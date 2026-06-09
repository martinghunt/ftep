package main

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/martinghunt/ftep"
	"github.com/spf13/cobra"
)

const enaBrowserViewBaseURL = "https://www.ebi.ac.uk/ena/browser/view/"

var openBrowser = openURLInBrowser

type openOptions struct {
	accession string
	printURL  bool
}

func newOpenCommand() *cobra.Command {
	var opts openOptions

	cmd := &cobra.Command{
		Use:   "open [accession]",
		Short: "Open an accession in the ENA browser",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeOpen(cmd, args, opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.accession, "accession", "a", "", "Accession to open")
	flags.BoolVar(&opts.printURL, "print-url", false, "Print the ENA browser URL instead of opening it")

	return cmd
}

func executeOpen(cmd *cobra.Command, args []string, opts openOptions) error {
	accession, err := openAccessionFromInputs(opts.accession, args)
	if err != nil {
		return err
	}

	browserURL, err := enaBrowserURL(accession)
	if err != nil {
		return err
	}

	if opts.printURL {
		fmt.Fprintln(cmd.OutOrStdout(), browserURL)
		return nil
	}

	if err := openBrowser(browserURL); err != nil {
		return fmt.Errorf("error opening browser: %w; try --print-url", err)
	}
	return nil
}

func openAccessionFromInputs(flagAccession string, args []string) (string, error) {
	if flagAccession != "" && len(args) > 0 {
		return "", fmt.Errorf("provide accession either as an argument or with -a/--accession, not both")
	}
	if flagAccession != "" {
		return flagAccession, nil
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("accession is required")
}

func enaBrowserURL(accession string) (string, error) {
	accession = strings.TrimSpace(accession)
	if accession == "" {
		return "", fmt.Errorf("accession is required")
	}

	if _, _, ok := ftep.IdentifyAccession(accession); !ok {
		return "", fmt.Errorf("accession format not recognised: %s", accession)
	}

	return enaBrowserViewBaseURL + url.PathEscape(accession), nil
}

func openURLInBrowser(browserURL string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", browserURL).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", browserURL).Run()
	default:
		return exec.Command("xdg-open", browserURL).Run()
	}
}
