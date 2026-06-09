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

	"github.com/martinghunt/ftep"
	"github.com/spf13/cobra"
)

var version = "local"
var newClient = ftep.NewClient

func main() {
	log.SetPrefix("[ftep] ")
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
		Use:           "ftep",
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
	cmd.AddCommand(newSearchCommand(), newReadsCommand(), newGetFieldsCommand())
	return cmd
}

type searchOptions struct {
	accession string
	accFile   string
	columns   string
	level     string
	outfmt    string
	debug     bool
}

func newSearchCommand() *cobra.Command {
	opts := searchOptions{
		columns: "DEFAULT",
		outfmt:  "tsv",
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
	flags.StringVar(&opts.level, "level", "", "Output level: study, sample, run, or assembly. Default is the input accession level")
	flags.StringVar(&opts.outfmt, "outfmt", opts.outfmt, "Output format: json or tsv")
	_ = flags.MarkHidden("acc_file")

	return cmd
}

func executeSearch(cmd *cobra.Command, opts searchOptions) error {
	if (opts.accession == "") == (opts.accFile == "") {
		return fmt.Errorf("exactly one of -a/--accession or -f/--acc_file is required")
	}
	if opts.outfmt != "tsv" && opts.outfmt != "json" {
		return fmt.Errorf("unsupported --outfmt %q; expected json or tsv", opts.outfmt)
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

	client := newClient()
	results, err := searchAccessions(cmd.Context(), client, accessions, fields, level, opts.debug, cmd.ErrOrStderr())
	if err != nil {
		return err
	}

	if opts.outfmt == "json" {
		return writeJSON(cmd.OutOrStdout(), results)
	}

	return writeTSV(cmd.OutOrStdout(), results, fields)
}

func parseSearchLevel(level string) (ftep.AccessionType, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "":
		return "", nil
	case string(ftep.AccessionTypeStudy):
		return ftep.AccessionTypeStudy, nil
	case string(ftep.AccessionTypeSample):
		return ftep.AccessionTypeSample, nil
	case string(ftep.AccessionTypeRun):
		return ftep.AccessionTypeRun, nil
	case string(ftep.AccessionTypeAssembly):
		return ftep.AccessionTypeAssembly, nil
	default:
		return "", fmt.Errorf("unsupported --level %q; expected study, sample, run, or assembly", level)
	}
}

func newGetFieldsCommand() *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:   "get_fields data_type",
		Short: "Get available fields for a given data type, such as read_run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if debug {
				log.Printf("getting fields for %s", args[0])
			}

			client := newClient()
			text, err := client.GetAllowedFields(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), text)
			if !strings.HasSuffix(text, "\n") {
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&debug, "debug", false, "More verbose logging")
	return cmd
}

func accessionsFromInputs(accession string, accFile string) ([]string, error) {
	if accession != "" {
		return []string{accession}, nil
	}
	return ftep.ReadAccessionsFile(accFile)
}

func searchAccessions(ctx context.Context, client *ftep.Client, accessions []string, fields []string, level ftep.AccessionType, debug bool, errOut io.Writer) ([]ftep.SearchResult, error) {
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
		if _, err := ftep.ResolveSearchLevel(accessionType, level); err != nil {
			return nil, err
		}

		toSearch = append(toSearch, accessionSearch{input: accession, fixed: fixedAccession, typ: accessionType})
	}

	results := make([]ftep.SearchResult, 0, len(toSearch))
	for _, accession := range toSearch {
		if debug {
			if level == "" {
				log.Printf("search for %s", accession.input)
			} else {
				log.Printf("search for %s at %s level", accession.input, level)
			}
		}

		resultType, newFields, records, err := client.Query(ctx, accession.fixed, accession.typ, fields, level)
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
			InputType:      accession.typ,
			ResultType:     resultType,
			Fields:         newFields,
			Records:        records,
		})
	}

	return results, nil
}

func writeJSON(out io.Writer, results []ftep.SearchResult) error {
	byAccession := make(map[string][]ftep.Record, len(results))
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
