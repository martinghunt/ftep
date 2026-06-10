package ichsm

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

const (
	ReadProtocolHTTPS = "https"
	ReadProtocolFTP   = "ftp"
)

var ReadFileFields = []string{"run_accession", "fastq_ftp", "fastq_md5", "fastq_bytes"}

type ReadFileOptions struct {
	Accessions []string
	Protocol   string
	OutputDir  string
}

type ReadFile struct {
	InputAccession string
	RunAccession   string
	Filename       string
	OutputPath     string
	URL            string
	MD5            string
	Bytes          string
}

func (c *Client) ReadFiles(ctx context.Context, opts ReadFileOptions) ([]ReadFile, error) {
	if len(opts.Accessions) == 0 {
		return nil, errors.New("no accessions provided")
	}

	protocol, err := NormalizeReadFileProtocol(opts.Protocol)
	if err != nil {
		return nil, err
	}

	results, err := c.Search(ctx, SearchOptions{
		Accessions: opts.Accessions,
		Fields:     ReadFileFields,
		Level:      AccessionTypeRun,
		Source:     SearchSourceENA,
	})
	if err != nil {
		return nil, err
	}

	opts.Protocol = protocol
	return ReadFilesFromSearchResults(results, opts)
}

func ReadFilesFromSearchResults(results []SearchResult, opts ReadFileOptions) ([]ReadFile, error) {
	protocol, err := NormalizeReadFileProtocol(opts.Protocol)
	if err != nil {
		return nil, err
	}
	opts.Protocol = protocol

	var files []ReadFile
	for _, result := range results {
		for _, record := range result.Records {
			recordFiles, err := ReadFilesFromRecord(result.InputAccession, record, opts)
			if err != nil {
				return nil, err
			}
			files = append(files, recordFiles...)
		}
	}
	return files, nil
}

func ReadFilesFromRecord(inputAccession string, record Record, opts ReadFileOptions) ([]ReadFile, error) {
	protocol, err := NormalizeReadFileProtocol(opts.Protocol)
	if err != nil {
		return nil, err
	}

	runAccession := recordString(record, "run_accession")
	urls := splitENAList(recordString(record, "fastq_ftp"))
	md5s := splitENAList(recordString(record, "fastq_md5"))
	byteCounts := splitENAList(recordString(record, "fastq_bytes"))

	files := make([]ReadFile, 0, len(urls))
	for i, rawURL := range urls {
		url, err := NormalizeReadDownloadURL(rawURL, protocol)
		if err != nil {
			return nil, err
		}

		filename := path.Base(strings.TrimPrefix(rawURL, "ftp://"))
		if filename == "." || filename == "/" || filename == "" {
			return nil, fmt.Errorf("could not determine filename from FASTQ URL %q", rawURL)
		}

		outputPath := filename
		if opts.OutputDir != "" {
			outputPath = filepath.Join(opts.OutputDir, filename)
		}

		files = append(files, ReadFile{
			InputAccession: inputAccession,
			RunAccession:   runAccession,
			Filename:       filename,
			OutputPath:     outputPath,
			URL:            url,
			MD5:            listValue(md5s, i),
			Bytes:          listValue(byteCounts, i),
		})
	}

	return files, nil
}

func NormalizeReadFileProtocol(protocol string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "", ReadProtocolHTTPS:
		return ReadProtocolHTTPS, nil
	case ReadProtocolFTP:
		return ReadProtocolFTP, nil
	default:
		return "", fmt.Errorf("unsupported read file protocol %q; expected https or ftp", protocol)
	}
}

func NormalizeReadDownloadURL(rawURL string, protocol string) (string, error) {
	protocol, err := NormalizeReadFileProtocol(protocol)
	if err != nil {
		return "", err
	}

	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("empty FASTQ URL")
	}

	if strings.HasPrefix(rawURL, "ftp://") {
		rawURL = strings.TrimPrefix(rawURL, "ftp://")
	}
	if strings.HasPrefix(rawURL, "https://") {
		rawURL = strings.TrimPrefix(rawURL, "https://")
	}
	if strings.HasPrefix(rawURL, "http://") {
		rawURL = strings.TrimPrefix(rawURL, "http://")
	}

	return protocol + "://" + rawURL, nil
}

func recordString(record Record, key string) string {
	value := record[key]
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func splitENAList(value string) []string {
	if value == "" || value == "." {
		return nil
	}

	parts := strings.Split(value, ";")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func listValue(values []string, index int) string {
	if index >= len(values) {
		return ""
	}
	return values[index]
}
