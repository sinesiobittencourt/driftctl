package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewGenDriftIgnoreCmd() *cobra.Command {
	opts := &filter.AnalysisListOptions{}

	cmd := &cobra.Command{
		Use:     "gen-driftignore",
		Short:   "Generate a driftignore file based on your scan result",
		Long:    "This command will generate a new driftignore file containing your current drifts",
		Example: "driftctl scan -o json://stdout | driftctl gen-driftignore",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Throw an error if a .driftignore file already exists
			_, err := os.Stat(filter.DriftIgnoreFilename)
			if !os.IsNotExist(err) {
				return errors.New("An existing .driftignore file was found. Please rename or delete the existing one to generate a new one.")
			}

			input, err := io.ReadAll(bufio.NewReader(os.Stdin))
			if err != nil {
				return err
			}

			analysis := &analyser.Analysis{}
			err = analysis.UnmarshalJSON(input)
			if err != nil {
				return err
			}

			// Sort resources in a predictable order
			analysis.SortResources()

			f, err := os.OpenFile(filter.DriftIgnoreFilename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}

			n, list := filter.AnalysisToList(analysis, *opts)
			_, err = f.WriteString(list)
			if err != nil {
				return err
			}

			fmt.Printf("Added %v resources to .driftignore\n", n)

			return nil
		},
	}

	fl := cmd.Flags()

	fl.BoolVar(&opts.IncludeUnmanaged, "unmanaged", true, "Include resources not managed by IaC")
	fl.BoolVar(&opts.IncludeDeleted, "missing", true, "Include missing resources")
	fl.BoolVar(&opts.IncludeDrifted, "changed", true, "Include resources that changed on cloud provider")

	return cmd
}
