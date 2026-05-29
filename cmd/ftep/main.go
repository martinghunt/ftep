package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/martinghunt/ftep"
)

var version = "local"
var newClient = ftep.NewClient

func main() {
	log.SetPrefix("[ftep] ")
	log.SetFlags(0)
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage(os.Stderr)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		usage(os.Stdout)
		return 0
	case "--version", "-version":
		fmt.Fprintln(os.Stdout, version)
		return 0
	case "search":
		return runSearch(args[1:])
	case "get_fields":
		return runGetFields(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		usage(os.Stderr)
		return 2
	}
}

func usage(w *os.File) {
	fmt.Fprintln(w, `ftep: query the ENA

Usage:
  ftep <command> <options>

Available commands:
  search       General search from an accession or file of accessions
  get_fields   Get available fields for a given data type, such as read_run

Use "ftep <command> -h" for command options.`)
}

func runSearch(args []string) int {
	var accession string
	var accFile string
	var columns string
	var outfmt string
	var sampleToRun bool
	var debug bool

	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), `Usage:
  ftep search [options]

Options:`)
		fs.PrintDefaults()
	}
	fs.BoolVar(&debug, "debug", false, "More verbose logging")
	fs.StringVar(&accession, "a", "", "Accession to search for")
	fs.StringVar(&accession, "accession", "", "Accession to search for")
	fs.StringVar(&accFile, "f", "", "File of accessions to search for, one per line")
	fs.StringVar(&accFile, "acc_file", "", "File of accessions to search for, one per line")
	fs.StringVar(&columns, "c", "DEFAULT", "Columns/fields to output, comma-separated, or SMALL, DEFAULT, BIG, ALL")
	fs.StringVar(&columns, "columns", "DEFAULT", "Columns/fields to output, comma-separated, or SMALL, DEFAULT, BIG, ALL")
	fs.StringVar(&columns, "fields", "DEFAULT", "Columns/fields to output, comma-separated, or SMALL, DEFAULT, BIG, ALL")
	fs.BoolVar(&sampleToRun, "s2r", false, "'sample to run': run data is reported for sample accessions")
	fs.StringVar(&outfmt, "outfmt", "tsv", "Output format: json or tsv")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", fs.Arg(0))
		return 2
	}
	if (accession == "") == (accFile == "") {
		fmt.Fprintln(os.Stderr, "exactly one of -a/--accession or -f/--acc_file is required")
		return 2
	}
	if outfmt != "tsv" && outfmt != "json" {
		fmt.Fprintf(os.Stderr, "unsupported --outfmt %q; expected json or tsv\n", outfmt)
		return 2
	}

	accessions, err := accessionsFromInputs(accession, accFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fields := strings.Split(columns, ",")
	client := newClient()
	results, err := searchAccessions(context.Background(), client, accessions, fields, sampleToRun, debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if outfmt == "json" {
		return writeJSON(os.Stdout, results)
	}

	if err := writeTSV(os.Stdout, results, fields); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func runGetFields(args []string) int {
	var debug bool

	fs := flag.NewFlagSet("get_fields", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), `Usage:
  ftep get_fields [options] data_type

Options:`)
		fs.PrintDefaults()
	}
	fs.BoolVar(&debug, "debug", false, "More verbose logging")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}

	if debug {
		log.Printf("getting fields for %s", fs.Arg(0))
	}

	client := newClient()
	text, err := client.GetAllowedFields(context.Background(), fs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Print(text)
	if !strings.HasSuffix(text, "\n") {
		fmt.Println()
	}
	return 0
}

func accessionsFromInputs(accession string, accFile string) ([]string, error) {
	if accession != "" {
		return []string{accession}, nil
	}
	return ftep.ReadAccessionsFile(accFile)
}

func searchAccessions(ctx context.Context, client *ftep.Client, accessions []string, fields []string, sampleToRun bool, debug bool) ([]ftep.SearchResult, error) {
	if len(accessions) == 0 {
		return nil, errors.New("no accessions provided")
	}

	type accessionSearch struct {
		input string
		fixed string
		typ   ftep.AccessionType
	}

	toSearch := make([]accessionSearch, 0, len(accessions))
	var firstType ftep.AccessionType
	for _, accession := range accessions {
		fixedAccession, accessionType, ok := ftep.IdentifyAccession(accession)
		if !ok {
			fmt.Fprintf(os.Stderr, "%s\t%s\n", accession, "")
			return nil, fmt.Errorf("error getting result types from accessions")
		}
		if firstType == "" {
			firstType = accessionType
		} else if accessionType != firstType {
			for _, searched := range toSearch {
				fmt.Fprintf(os.Stderr, "%s\t%s\n", searched.input, searched.typ)
			}
			fmt.Fprintf(os.Stderr, "%s\t%s\n", accession, accessionType)
			return nil, fmt.Errorf("error getting result types from accessions")
		}

		toSearch = append(toSearch, accessionSearch{input: accession, fixed: fixedAccession, typ: accessionType})
	}

	results := make([]ftep.SearchResult, 0, len(toSearch))
	for _, accession := range toSearch {
		if debug {
			log.Printf("search for %s", accession.input)
		}

		newFields, records, err := client.Query(ctx, accession.fixed, accession.typ, fields, sampleToRun)
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

		results = append(results, ftep.SearchResult{
			InputAccession: accession.input,
			FixedAccession: accession.fixed,
			Type:           accession.typ,
			Fields:         newFields,
			Records:        records,
		})
	}

	return results, nil
}

func writeJSON(out io.Writer, results []ftep.SearchResult) int {
	byAccession := make(map[string][]ftep.Record, len(results))
	for _, result := range results {
		byAccession[result.InputAccession] = result.Records
	}

	encoded, err := json.MarshalIndent(byAccession, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Fprintln(out, string(encoded))
	return 0
}

func writeTSV(out io.Writer, results []ftep.SearchResult, requestedFields []string) error {
	var columns []string
	for _, result := range results {
		if len(result.Records) == 0 {
			continue
		}

		if columns == nil {
			if requestedAllFields(requestedFields) {
				columns = ftep.SortedRecordKeys(result.Records[0])
			} else {
				columns = result.Fields
			}
			fmt.Fprintln(out, strings.Join(append([]string{"input_accession"}, columns...), "\t"))
		} else if !requestedAllFields(requestedFields) && !sameStringSet(columns, result.Fields) {
			return fmt.Errorf("field set changed between results")
		}

		for _, record := range result.Records {
			row := make([]string, 0, len(columns)+1)
			row = append(row, result.InputAccession)
			for _, column := range columns {
				row = append(row, formatValue(record[column]))
			}
			fmt.Fprintln(out, strings.Join(row, "\t"))
		}
	}
	return nil
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
