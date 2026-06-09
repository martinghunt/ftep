package main

import (
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/martinghunt/ichsm"
	"github.com/spf13/cobra"
)

const (
	readsFormatManifest = "manifest"
	readsFormatTable    = "table"
	readsFormatURLs     = "urls"
	readsFormatWget     = "wget"
	readsFormatCurl     = "curl"
	readsFormatMD5      = "md5"

	readsProtocolHTTPS = "https"
	readsProtocolFTP   = "ftp"
)

var readsFields = []string{"run_accession", "fastq_ftp", "fastq_md5", "fastq_bytes"}

type readsOptions struct {
	accession string
	accFile   string
	outfmt    string
	protocol  string
	outputDir string
	debug     bool
}

type readFile struct {
	InputAccession string
	RunAccession   string
	Filename       string
	OutputPath     string
	URL            string
	MD5            string
	Bytes          string
}

func newReadsCommand() *cobra.Command {
	opts := readsOptions{
		outfmt:   readsFormatManifest,
		protocol: readsProtocolHTTPS,
	}

	cmd := &cobra.Command{
		Use:   "reads",
		Short: "Print FASTQ download manifests or commands for an accession",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeReads(cmd, opts)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.debug, "debug", false, "More verbose logging")
	flags.StringVarP(&opts.accession, "accession", "a", "", "Accession to find reads for")
	flags.StringVarP(&opts.accFile, "acc-file", "f", "", "File of accessions to find reads for, one per line")
	flags.StringVar(&opts.accFile, "acc_file", "", "File of accessions to find reads for, one per line")
	flags.StringVar(&opts.outfmt, "outfmt", opts.outfmt, "Output format: manifest, table, urls, wget, curl, or md5")
	flags.StringVar(&opts.protocol, "protocol", opts.protocol, "Download URL protocol: https or ftp")
	flags.StringVarP(&opts.outputDir, "output-dir", "o", "", "Directory to use in printed output filenames")
	_ = flags.MarkHidden("acc_file")

	return cmd
}

func executeReads(cmd *cobra.Command, opts readsOptions) error {
	if (opts.accession == "") == (opts.accFile == "") {
		return fmt.Errorf("exactly one of -a/--accession or -f/--acc-file is required")
	}

	outfmt, err := parseReadsOutfmt(opts.outfmt)
	if err != nil {
		return err
	}
	protocol, err := parseReadsProtocol(opts.protocol)
	if err != nil {
		return err
	}

	accessions, err := accessionsFromInputs(opts.accession, opts.accFile)
	if err != nil {
		return err
	}

	client := newClient()
	results, err := searchAccessions(cmd.Context(), client, accessions, readsFields, ichsm.AccessionTypeRun, ichsm.SearchSourceENA, opts.debug, cmd.ErrOrStderr())
	if err != nil {
		return err
	}

	files, err := readFilesFromResults(results, protocol, opts.outputDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("no FASTQ files found")
	}

	return writeReads(cmd.OutOrStdout(), files, outfmt)
}

func parseReadsOutfmt(outfmt string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(outfmt)) {
	case readsFormatManifest:
		return readsFormatManifest, nil
	case readsFormatTable, "human":
		return readsFormatTable, nil
	case readsFormatURLs:
		return readsFormatURLs, nil
	case readsFormatWget:
		return readsFormatWget, nil
	case readsFormatCurl:
		return readsFormatCurl, nil
	case readsFormatMD5:
		return readsFormatMD5, nil
	default:
		return "", fmt.Errorf("unsupported --outfmt %q; expected manifest, table, urls, wget, curl, or md5", outfmt)
	}
}

func parseReadsProtocol(protocol string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case readsProtocolHTTPS:
		return readsProtocolHTTPS, nil
	case readsProtocolFTP:
		return readsProtocolFTP, nil
	default:
		return "", fmt.Errorf("unsupported --protocol %q; expected https or ftp", protocol)
	}
}

func readFilesFromResults(results []ichsm.SearchResult, protocol string, outputDir string) ([]readFile, error) {
	var files []readFile
	for _, result := range results {
		for _, record := range result.Records {
			recordFiles, err := readFilesFromRecord(result.InputAccession, record, protocol, outputDir)
			if err != nil {
				return nil, err
			}
			files = append(files, recordFiles...)
		}
	}
	return files, nil
}

func readFilesFromRecord(inputAccession string, record ichsm.Record, protocol string, outputDir string) ([]readFile, error) {
	runAccession := recordString(record, "run_accession")
	urls := splitENAList(recordString(record, "fastq_ftp"))
	md5s := splitENAList(recordString(record, "fastq_md5"))
	byteCounts := splitENAList(recordString(record, "fastq_bytes"))

	files := make([]readFile, 0, len(urls))
	for i, rawURL := range urls {
		url, err := normalizeDownloadURL(rawURL, protocol)
		if err != nil {
			return nil, err
		}

		filename := path.Base(strings.TrimPrefix(rawURL, "ftp://"))
		if filename == "." || filename == "/" || filename == "" {
			return nil, fmt.Errorf("could not determine filename from FASTQ URL %q", rawURL)
		}

		outputPath := filename
		if outputDir != "" {
			outputPath = filepath.Join(outputDir, filename)
		}

		files = append(files, readFile{
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

func recordString(record ichsm.Record, key string) string {
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

func normalizeDownloadURL(rawURL string, protocol string) (string, error) {
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

func writeReads(out io.Writer, files []readFile, format string) error {
	switch format {
	case readsFormatManifest:
		return writeReadsManifest(out, files)
	case readsFormatTable:
		return writeAlignedRows(out, readFilesRows(files))
	case readsFormatURLs:
		for _, file := range files {
			fmt.Fprintln(out, file.URL)
		}
	case readsFormatWget:
		for _, file := range files {
			fmt.Fprintf(out, "wget -c -O %s %s\n", shellQuote(file.OutputPath), shellQuote(file.URL))
		}
	case readsFormatCurl:
		for _, file := range files {
			fmt.Fprintf(out, "curl -L --fail --continue-at - --output %s %s\n", shellQuote(file.OutputPath), shellQuote(file.URL))
		}
	case readsFormatMD5:
		for _, file := range files {
			if file.MD5 == "" {
				return fmt.Errorf("missing MD5 checksum for %s", file.URL)
			}
			fmt.Fprintf(out, "%s  %s\n", file.MD5, file.OutputPath)
		}
	default:
		return fmt.Errorf("unsupported reads format %q", format)
	}

	return nil
}

func writeReadsManifest(out io.Writer, files []readFile) error {
	return writeDelimitedRows(out, readFilesRows(files), "\t")
}

func readFilesRows(files []readFile) [][]string {
	rows := [][]string{{"input_accession", "run_accession", "filename", "url", "md5", "bytes"}}
	for _, file := range files {
		rows = append(rows, []string{
			file.InputAccession,
			file.RunAccession,
			file.Filename,
			file.URL,
			emptyAsDot(file.MD5),
			emptyAsDot(file.Bytes),
		})
	}
	return rows
}

func emptyAsDot(value string) string {
	if value == "" {
		return "."
	}
	return value
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
