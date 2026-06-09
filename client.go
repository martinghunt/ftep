package ftep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Record is one metadata record returned by the ENA portal.
type Record map[string]any

// Client queries the ENA portal API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// SearchOptions configures a multi-accession search.
type SearchOptions struct {
	Accessions []string
	Fields     []string
	Level      AccessionType
}

// SearchResult contains records for one input accession.
type SearchResult struct {
	InputAccession string        `json:"input_accession"`
	FixedAccession string        `json:"fixed_accession"`
	InputType      AccessionType `json:"input_type"`
	ResultType     AccessionType `json:"result_type"`
	Fields         []string      `json:"fields"`
	Records        []Record      `json:"records"`
}

// NewClient returns a client configured for the public ENA portal.
func NewClient() *Client {
	return &Client{
		BaseURL: BasePortalURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchKeyValue returns the ENA search parameter for an input accession type
// at a requested output level.
func SearchKeyValue(queryType AccessionType, resultType AccessionType, accession string) (string, string, error) {
	switch queryType {
	case AccessionTypeAssembly:
		if resultType != AccessionTypeAssembly {
			return "", "", unsupportedSearchLevel(queryType, resultType)
		}
		return "query", "accession=" + accession, nil
	case AccessionTypeStudy:
		switch resultType {
		case AccessionTypeStudy:
			return "query", "study_accession=" + accession + " OR secondary_study_accession=" + accession, nil
		case AccessionTypeSample, AccessionTypeRun, AccessionTypeAssembly:
			return "query", "study_accession=" + accession, nil
		default:
			return "", "", unsupportedSearchLevel(queryType, resultType)
		}
	case AccessionTypeSample:
		switch resultType {
		case AccessionTypeSample, AccessionTypeRun, AccessionTypeAssembly:
			return "query", "sample_accession=" + accession + " OR secondary_sample_accession=" + accession, nil
		default:
			return "", "", unsupportedSearchLevel(queryType, resultType)
		}
	case AccessionTypeRun:
		if resultType != AccessionTypeRun && resultType != AccessionTypeAssembly {
			return "", "", unsupportedSearchLevel(queryType, resultType)
		}
		return "query", "run_accession=" + accession, nil
	case AccessionTypeExperiment:
		if resultType != AccessionTypeRun {
			return "", "", unsupportedSearchLevel(queryType, resultType)
		}
		return "query", "experiment_accession=" + accession, nil
	default:
		return "", "", fmt.Errorf("unsupported accession type %q", queryType)
	}
}

// ResolveFields expands SMALL, DEFAULT, and BIG field presets for an accession
// type. Unknown presets, including ALL, are passed through unchanged.
func ResolveFields(accessionType AccessionType, fields []string) ([]string, error) {
	if len(fields) == 0 {
		fields = []string{"DEFAULT"}
	}

	presets, ok := fieldPresets[accessionType]
	if !ok {
		return nil, fmt.Errorf("unsupported accession type %q", accessionType)
	}

	if preset, ok := presets[fields[0]]; ok {
		return copyStrings(preset), nil
	}

	return copyStrings(fields), nil
}

// Query searches ENA for one normalized accession at a requested output level.
func (c *Client) Query(ctx context.Context, accession string, accessionType AccessionType, fields []string, level AccessionType) (AccessionType, []string, []Record, error) {
	resultType, err := ResolveSearchLevel(accessionType, level)
	if err != nil {
		return "", nil, nil, err
	}

	if accessionType == AccessionTypeStudy && resultType != AccessionTypeStudy {
		accession, err = c.resolvePrimaryStudyAccession(ctx, accession)
		if err != nil {
			return "", nil, nil, err
		}
	}

	searchKey, searchValue, err := SearchKeyValue(accessionType, resultType, accession)
	if err != nil {
		return "", nil, nil, err
	}

	endpoint, ok := urlSearchData[resultType]
	if !ok {
		return "", nil, nil, fmt.Errorf("unsupported accession type %q", resultType)
	}

	resolvedFields, err := ResolveFields(resultType, fields)
	if err != nil {
		return "", nil, nil, err
	}

	params := url.Values{}
	params.Set("result", endpoint.result)
	params.Set(searchKey, searchValue)
	params.Set("format", "json")
	params.Set("fields", strings.Join(resolvedFields, ","))

	results, err := c.requestJSON(ctx, endpoint.mainType, params)
	if err != nil {
		return "", nil, nil, err
	}

	return resultType, resolvedFields, results, nil
}

// Search identifies and queries a set of accessions. As in the original CLI,
// all accessions must have the same inferred type.
func (c *Client) Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	if len(opts.Accessions) == 0 {
		return nil, fmt.Errorf("no accessions provided")
	}

	type accessionSearch struct {
		input string
		fixed string
		typ   AccessionType
	}

	toSearch := make([]accessionSearch, 0, len(opts.Accessions))
	var firstType AccessionType
	for _, accession := range opts.Accessions {
		fixedAccession, accessionType, ok := IdentifyAccession(accession)
		if !ok {
			return nil, fmt.Errorf("accession format not recognised: %s", accession)
		}
		if firstType == "" {
			firstType = accessionType
		} else if accessionType != firstType {
			return nil, fmt.Errorf("accessions must all be the same type: got %s and %s", firstType, accessionType)
		}
		toSearch = append(toSearch, accessionSearch{input: accession, fixed: fixedAccession, typ: accessionType})
	}

	results := make([]SearchResult, 0, len(toSearch))
	for _, accession := range toSearch {
		resultType, fields, records, err := c.Query(ctx, accession.fixed, accession.typ, opts.Fields, opts.Level)
		if err != nil {
			return nil, err
		}
		results = append(results, SearchResult{
			InputAccession: accession.input,
			FixedAccession: accession.fixed,
			InputType:      accession.typ,
			ResultType:     resultType,
			Fields:         fields,
			Records:        records,
		})
	}

	return results, nil
}

// ResolveSearchLevel returns the ENA result level to search. A zero level means
// the closest report level for the input accession type.
func ResolveSearchLevel(inputType AccessionType, level AccessionType) (AccessionType, error) {
	if level == "" {
		if inputType == AccessionTypeExperiment {
			return AccessionTypeRun, nil
		}
		return inputType, nil
	}

	switch level {
	case AccessionTypeAssembly, AccessionTypeStudy, AccessionTypeSample, AccessionTypeRun:
	default:
		return "", fmt.Errorf("unsupported search level %q; expected study, sample, run, or assembly", level)
	}

	if _, _, err := SearchKeyValue(inputType, level, ""); err != nil {
		return "", err
	}
	return level, nil
}

func unsupportedSearchLevel(inputType AccessionType, level AccessionType) error {
	return fmt.Errorf("cannot search %s accessions at %s level", inputType, level)
}

func (c *Client) resolvePrimaryStudyAccession(ctx context.Context, accession string) (string, error) {
	searchKey, searchValue, err := SearchKeyValue(AccessionTypeStudy, AccessionTypeStudy, accession)
	if err != nil {
		return "", err
	}

	endpoint := urlSearchData[AccessionTypeStudy]
	params := url.Values{}
	params.Set("result", endpoint.result)
	params.Set(searchKey, searchValue)
	params.Set("format", "json")
	params.Set("fields", "study_accession")

	results, err := c.requestJSON(ctx, endpoint.mainType, params)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no study found for accession %s", accession)
	}

	studyAccession, ok := results[0]["study_accession"].(string)
	if !ok || studyAccession == "" {
		return "", fmt.Errorf("no primary study accession found for accession %s", accession)
	}
	return studyAccession, nil
}

// GetAllowedFields returns the ENA searchFields response for a result type,
// such as read_run.
func (c *Client) GetAllowedFields(ctx context.Context, dataType string) (string, error) {
	params := url.Values{}
	params.Set("result", dataType)
	return c.requestText(ctx, "searchFields", params)
}

// SortedRecordKeys returns record keys in deterministic order. It is useful
// when ENA's ALL field preset is requested and the output columns come from the
// returned JSON object.
func SortedRecordKeys(record Record) []string {
	keys := make([]string, 0, len(record))
	for key := range record {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c *Client) requestJSON(ctx context.Context, path string, params url.Values) ([]Record, error) {
	body, err := c.request(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var results []Record
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&results); err != nil {
		return nil, fmt.Errorf("error parsing json from query: %w", err)
	}

	for _, result := range results {
		for key, value := range result {
			if value == "" {
				result[key] = nil
			}
		}
	}

	return results, nil
}

func (c *Client) requestText(ctx context.Context, path string, params url.Values) (string, error) {
	body, err := c.request(ctx, path, params)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *Client) request(ctx context.Context, path string, params url.Values) ([]byte, error) {
	requestURL, err := c.requestURL(path, params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("error requesting data from %s: %w", requestURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error requesting data: status=%d url=%s body=%s", resp.StatusCode, requestURL, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func (c *Client) requestURL(path string, params url.Values) (string, error) {
	baseURL := BasePortalURL
	if c != nil && c.BaseURL != "" {
		baseURL = c.BaseURL
	}

	parsed, err := url.Parse(strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/"))
	if err != nil {
		return "", err
	}
	parsed.RawQuery = params.Encode()
	return parsed.String(), nil
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}

	return &http.Client{Timeout: 30 * time.Second}
}
