package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/martinghunt/ichsm"
	"github.com/spf13/cobra"
)

var version = "local"
var newClient = ichsm.NewClient

const (
	outputFormatJSON  = "json"
	outputFormatTable = "table"
	outputFormatTSV   = "tsv"
)

func main() {
	log.SetPrefix("[ichsm] ")
	log.SetFlags(0)
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	cmd := newRootCommand(os.Stdout, os.Stderr)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}

func newRootCommand(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ichsm",
		Short:         "query the ENA",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.AddCommand(newSearchCommand(), newReadsCommand(), newOpenCommand(), newIdentifyCommand(), newGetFieldsCommand())
	return cmd
}

type searchOptions struct {
	accession string
	accFile   string
	columns   string
	level     string
	source    string
	apiKey    string
	email     string
	outfmt    string
	debug     bool
}

func newSearchCommand() *cobra.Command {
	opts := searchOptions{
		columns: "DEFAULT",
		outfmt:  outputFormatTSV,
	}
	cmd := &cobra.Command{
		Use:   "search",
		Short: "General search from an accession or file of accessions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeSearch(cmd, opts)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.debug, "debug", false, "More verbose logging")
	flags.StringVarP(&opts.accession, "accession", "a", "", "Accession to search for")
	flags.StringVarP(&opts.accFile, "acc-file", "f", "", "File of accessions to search for, one per line")
	flags.StringVar(&opts.accFile, "acc_file", "", "File of accessions to search for, one per line")
	flags.StringVarP(&opts.columns, "columns", "c", opts.columns, "Columns/fields to output, comma-separated, or SMALL, DEFAULT, BIG, ALL")
	flags.StringVar(&opts.columns, "fields", opts.columns, "Columns/fields to output, comma-separated, or SMALL, DEFAULT, BIG, ALL")
	flags.StringVar(&opts.level, "level", "", "Output level: study, sample, run, assembly, sequence, coding, contig_set, wgs_set, tsa_set, or tls_set. Default is the input accession level")
	flags.StringVar(&opts.source, "source", string(ichsm.SearchSourceAuto), "Metadata source: auto, ena, or ncbi")
	flags.StringVar(&opts.apiKey, "api-key", "", "NCBI API key; defaults to NCBI_API_KEY")
	flags.StringVar(&opts.email, "email", "", "Email sent to NCBI; defaults to NCBI_EMAIL")
	flags.StringVar(&opts.outfmt, "outfmt", opts.outfmt, "Output format: json, table, or tsv")
	_ = flags.MarkHidden("acc_file")

	return cmd
}

func executeSearch(cmd *cobra.Command, opts searchOptions) error {
	if (opts.accession == "") == (opts.accFile == "") {
		return fmt.Errorf("exactly one of -a/--accession or -f/--acc_file is required")
	}
	outfmt, err := parseOutputFormat(opts.outfmt, true)
	if err != nil {
		return err
	}

	accessions, err := accessionsFromInputs(opts.accession, opts.accFile)
	if err != nil {
		return err
	}

	fields := strings.Split(opts.columns, ",")
	level, err := parseSearchLevel(opts.level)
	if err != nil {
		return err
	}
	source, err := parseSearchSource(opts.source)
	if err != nil {
		return err
	}

	client := newClient()
	if opts.apiKey == "" {
		opts.apiKey = os.Getenv("NCBI_API_KEY")
	}
	if opts.email == "" {
		opts.email = os.Getenv("NCBI_EMAIL")
	}
	client.NCBIAPIKey = opts.apiKey
	client.NCBIEmail = opts.email

	results, err := searchAccessions(cmd.Context(), client, accessions, fields, level, source, opts.debug, cmd.ErrOrStderr())
	if err != nil {
		return err
	}

	if outfmt == outputFormatJSON {
		return writeJSON(cmd.OutOrStdout(), results)
	}
	if outfmt == outputFormatTable {
		return writeTable(cmd.OutOrStdout(), results, fields)
	}

	return writeTSV(cmd.OutOrStdout(), results, fields)
}

func parseOutputFormat(format string, allowJSON bool) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case outputFormatTSV:
		return outputFormatTSV, nil
	case outputFormatTable, "human":
		return outputFormatTable, nil
	case outputFormatJSON:
		if allowJSON {
			return outputFormatJSON, nil
		}
	}

	if allowJSON {
		return "", fmt.Errorf("unsupported --outfmt %q; expected json, table, or tsv", format)
	}
	return "", fmt.Errorf("unsupported --outfmt %q; expected table or tsv", format)
}

func parseSearchLevel(level string) (ichsm.AccessionType, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "":
		return "", nil
	case string(ichsm.AccessionTypeStudy):
		return ichsm.AccessionTypeStudy, nil
	case string(ichsm.AccessionTypeSample):
		return ichsm.AccessionTypeSample, nil
	case string(ichsm.AccessionTypeRun):
		return ichsm.AccessionTypeRun, nil
	case string(ichsm.AccessionTypeAssembly):
		return ichsm.AccessionTypeAssembly, nil
	case string(ichsm.AccessionTypeSequence):
		return ichsm.AccessionTypeSequence, nil
	case string(ichsm.AccessionTypeCoding):
		return ichsm.AccessionTypeCoding, nil
	case string(ichsm.AccessionTypeContigSet):
		return ichsm.AccessionTypeContigSet, nil
	case string(ichsm.AccessionTypeWGSSet):
		return ichsm.AccessionTypeWGSSet, nil
	case string(ichsm.AccessionTypeTSASet):
		return ichsm.AccessionTypeTSASet, nil
	case string(ichsm.AccessionTypeTLSSet):
		return ichsm.AccessionTypeTLSSet, nil
	default:
		return "", fmt.Errorf("unsupported --level %q; expected study, sample, run, assembly, sequence, coding, contig_set, wgs_set, tsa_set, or tls_set", level)
	}
}

func parseSearchSource(source string) (ichsm.SearchSource, error) {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "", string(ichsm.SearchSourceAuto):
		return ichsm.SearchSourceAuto, nil
	case string(ichsm.SearchSourceENA):
		return ichsm.SearchSourceENA, nil
	case string(ichsm.SearchSourceNCBI):
		return ichsm.SearchSourceNCBI, nil
	default:
		return "", fmt.Errorf("unsupported --source %q; expected auto, ena, or ncbi", source)
	}
}

func newGetFieldsCommand() *cobra.Command {
	var debug bool
	outfmt := outputFormatTSV

	cmd := &cobra.Command{
		Use:   "get_fields [data_type]",
		Short: "List ENA data types or fields for a given data type",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedOutfmt, err := parseOutputFormat(outfmt, false)
			if err != nil {
				return err
			}

			client := newClient()
			var text string
			if len(args) == 0 {
				if debug {
					log.Printf("getting available data types")
				}
				text, err = client.GetResultTypes(cmd.Context())
				if err == nil {
					text = appendICHSMSearchColumn(text)
				}
			} else {
				if debug {
					log.Printf("getting fields for %s", args[0])
				}
				text, err = client.GetAllowedFields(cmd.Context(), args[0])
			}
			if err != nil {
				return err
			}
			if parsedOutfmt == outputFormatTable {
				return writeAlignedRows(cmd.OutOrStdout(), tsvTextRows(text))
			}
			return writeTextWithTrailingNewline(cmd.OutOrStdout(), text)
		},
	}
	cmd.Flags().BoolVar(&debug, "debug", false, "More verbose logging")
	cmd.Flags().StringVar(&outfmt, "outfmt", outfmt, "Output format: table or tsv")
	return cmd
}

func writeTextWithTrailingNewline(out io.Writer, text string) error {
	fmt.Fprint(out, text)
	if !strings.HasSuffix(text, "\n") {
		fmt.Fprintln(out)
	}
	return nil
}

func appendICHSMSearchColumn(text string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return text
	}

	type resultTypeRow struct {
		resultType string
		supported  bool
		line       string
	}

	rows := make([]resultTypeRow, 0, len(lines)-1)
	for _, line := range lines[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) == 0 || fields[0] == "" {
			continue
		}
		rows = append(rows, resultTypeRow{
			resultType: fields[0],
			supported:  ichsmSearchSupportsResult(fields[0]),
			line:       line,
		})
	}
	sort.Slice(rows, func(i int, j int) bool {
		if rows[i].supported != rows[j].supported {
			return rows[i].supported
		}
		return rows[i].resultType < rows[j].resultType
	})

	out := make([]string, 0, len(lines))
	out = append(out, lines[0]+"\tichsm_search")
	for _, row := range rows {
		out = append(out, row.line+"\t"+yesNo(row.supported))
	}
	return strings.Join(out, "\n") + "\n"
}

func ichsmSearchSupportsResult(resultType string) bool {
	switch resultType {
	case "assembly", "coding", "read_run", "sample", "sequence", "study", "tls_set", "tsa_set", "wgs_set":
		return true
	default:
		return false
	}
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func accessionsFromInputs(accession string, accFile string) ([]string, error) {
	if accession != "" {
		return []string{accession}, nil
	}
	return ichsm.ReadAccessionsFile(accFile)
}

func searchAccessions(ctx context.Context, client *ichsm.Client, accessions []string, fields []string, level ichsm.AccessionType, source ichsm.SearchSource, debug bool, errOut io.Writer) ([]ichsm.SearchResult, error) {
	if len(accessions) == 0 {
		return nil, errors.New("no accessions provided")
	}

	type accessionSearch struct {
		input string
		fixed string
		typ   ichsm.AccessionType
	}

	toSearch := make([]accessionSearch, 0, len(accessions))
	var firstType ichsm.AccessionType
	for _, accession := range accessions {
		fixedAccession, accessionType, ok := ichsm.IdentifyAccession(accession)
		if !ok {
			fmt.Fprintf(errOut, "%s\t%s\n", accession, "")
			return nil, fmt.Errorf("error getting result types from accessions")
		}
		if firstType == "" {
			firstType = accessionType
		} else if accessionType != firstType {
			for _, searched := range toSearch {
				fmt.Fprintf(errOut, "%s\t%s\n", searched.input, searched.typ)
			}
			fmt.Fprintf(errOut, "%s\t%s\n", accession, accessionType)
			return nil, fmt.Errorf("error getting result types from accessions")
		}
		if _, err := ichsm.ResolveSearchLevel(accessionType, level); err != nil {
			return nil, err
		}

		toSearch = append(toSearch, accessionSearch{input: accession, fixed: fixedAccession, typ: accessionType})
	}

	results := make([]ichsm.SearchResult, 0, len(toSearch))
	for _, accession := range toSearch {
		if debug {
			if level == "" {
				log.Printf("search for %s", accession.input)
			} else {
				log.Printf("search for %s at %s level", accession.input, level)
			}
		}

		resultSource, resultType, newFields, records, err := client.QueryWithSource(ctx, accession.input, accession.fixed, accession.typ, fields, level, source)
		if err != nil {
			log.Printf("warning: error getting data for accession %s. Skipping", accession.input)
			if debug {
				log.Printf("warning: %v", err)
			}
			continue
		}
		if len(records) == 0 {
			log.Printf("warning: no results returned for accession %s. Skipping", accession.input)
			continue
		}

		results = append(results, ichsm.SearchResult{
			InputAccession: accession.input,
			FixedAccession: accession.fixed,
			InputType:      accession.typ,
			ResultType:     resultType,
			Source:         resultSource,
			Fields:         newFields,
			Records:        records,
		})
	}

	return results, nil
}

func writeJSON(out io.Writer, results []ichsm.SearchResult) error {
	byAccession := make(map[string][]ichsm.Record, len(results))
	for _, result := range results {
		byAccession[result.InputAccession] = result.Records
	}

	encoded, err := json.MarshalIndent(byAccession, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(out, string(encoded))
	return nil
}

func writeTSV(out io.Writer, results []ichsm.SearchResult, requestedFields []string) error {
	rows, err := searchRows(results, requestedFields)
	if err != nil {
		return err
	}
	return writeDelimitedRows(out, rows, "\t")
}

func writeTable(out io.Writer, results []ichsm.SearchResult, requestedFields []string) error {
	rows, err := searchRows(results, requestedFields)
	if err != nil {
		return err
	}
	return writeAlignedRows(out, rows)
}

func searchRows(results []ichsm.SearchResult, requestedFields []string) ([][]string, error) {
	var columns []string
	var rows [][]string
	for _, result := range results {
		if len(result.Records) == 0 {
			continue
		}

		if columns == nil {
			if requestedAllFields(requestedFields) {
				columns = ichsm.SortedRecordKeys(result.Records[0])
			} else {
				columns = result.Fields
			}
			rows = append(rows, append([]string{"input_accession"}, columns...))
		} else if !requestedAllFields(requestedFields) && !sameStringSet(columns, result.Fields) {
			return nil, fmt.Errorf("field set changed between results")
		}

		for _, record := range result.Records {
			row := make([]string, 0, len(columns)+1)
			row = append(row, result.InputAccession)
			for _, column := range columns {
				row = append(row, formatValue(record[column]))
			}
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func requestedAllFields(fields []string) bool {
	return len(fields) == 1 && fields[0] == "ALL"
}

func sameStringSet(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	ac := append([]string(nil), a...)
	bc := append([]string(nil), b...)
	sort.Strings(ac)
	sort.Strings(bc)
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}

func formatValue(value any) string {
	if value == nil {
		return "."
	}
	return fmt.Sprint(value)
}
