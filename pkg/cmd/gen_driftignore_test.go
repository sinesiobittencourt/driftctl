package cmd

import (
	"errors"
	"os"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGenDriftIgnoreCmd_Input(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		input  string
		output string
		err    error
	}{
		{
			name:   "test error on invalid input",
			args:   []string{},
			input:  "./testdata/input_stdin_invalid.json",
			output: "./testdata/output_stdin_empty.txt",
			err:    errors.New("invalid character 'i' looking for beginning of value"),
		},
		{
			name:   "test empty driftignore with valid input",
			args:   []string{},
			input:  "./testdata/input_stdin_empty.json",
			output: "./testdata/output_stdin_empty.txt",
			err:    nil,
		},
		{
			name:   "test driftignore content with valid input",
			args:   []string{},
			input:  "./testdata/input_stdin_valid.json",
			output: "./testdata/output_stdin_valid.txt",
			err:    nil,
		},
		{
			name:   "test driftignore content with valid input and filter missing & changed only",
			args:   []string{"--unmanaged=false"},
			input:  "./testdata/input_stdin_valid.json",
			output: "./testdata/output_stdin_valid_filter.txt",
			err:    nil,
		},
		{
			name:   "test driftignore content with valid input and filter unmanaged only",
			args:   []string{"--unmanaged", "--missing=false", "--changed=false"},
			input:  "./testdata/input_stdin_valid.json",
			output: "./testdata/output_stdin_valid_filter2.txt",
			err:    nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(NewGenDriftIgnoreCmd())

			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}

			stdin := os.Stdin
			defer func() { os.Stdin = stdin }()
			os.Stdin = r

			input, err := os.ReadFile(c.input)
			if err != nil {
				t.Fatal(err)
			}

			output, err := os.ReadFile(c.output)
			if err != nil {
				t.Fatal(err)
			}

			_, err = w.Write(input)
			if err != nil {
				t.Fatal(err)
			}
			err = w.Close()
			if err != nil {
				t.Fatal(err)
			}

			args := append([]string{"gen-driftignore"}, c.args...)

			_, err = test.Execute(rootCmd, args...)
			if c.err != nil {
				assert.EqualError(t, err, c.err.Error())
				return
			} else {
				assert.Equal(t, c.err, err)
			}

			got, err := os.ReadFile(".driftignore")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(".driftignore")

			assert.Equal(t, string(output), string(got))
		})
	}
}

func TestGenDriftIgnoreCmd_ValidFlags(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	genDriftIgnoreCmd := NewGenDriftIgnoreCmd()
	genDriftIgnoreCmd.RunE = func(_ *cobra.Command, args []string) error { return nil }
	rootCmd.AddCommand(genDriftIgnoreCmd)

	cases := []struct {
		args []string
	}{
		{args: []string{"gen-driftignore", "--unmanaged"}},
		{args: []string{"gen-driftignore", "--missing"}},
		{args: []string{"gen-driftignore", "--changed"}},
		{args: []string{"gen-driftignore", "--changed=false", "--missing=false", "--unmanaged=false"}},
	}

	for _, tt := range cases {
		output, err := test.Execute(rootCmd, tt.args...)
		if output != "" {
			t.Errorf("Unexpected output: %v", output)
		}
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestGenDriftIgnoreCmd_InvalidFlags(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	genDriftIgnoreCmd := NewGenDriftIgnoreCmd()
	genDriftIgnoreCmd.RunE = func(_ *cobra.Command, args []string) error { return nil }
	rootCmd.AddCommand(genDriftIgnoreCmd)

	cases := []struct {
		args []string
		err  error
	}{
		{args: []string{"gen-driftignore", "--deleted"}, err: errors.New("unknown flag: --deleted")},
		{args: []string{"gen-driftignore", "--drifted"}, err: errors.New("unknown flag: --drifted")},
		{args: []string{"gen-driftignore", "--from"}, err: errors.New("unknown flag: --from")},
	}

	for _, tt := range cases {
		_, err := test.Execute(rootCmd, tt.args...)
		assert.EqualError(t, err, tt.err.Error())
	}
}

func TestGenDriftIgnoreCmd_FileExist(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(NewGenDriftIgnoreCmd())

	_, err := os.OpenFile(filter.DriftIgnoreFilename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(".driftignore")

	_, err = test.Execute(rootCmd, "gen-driftignore")
	assert.EqualError(t, err, "An existing .driftignore file was found. Please rename or delete the existing one to generate a new one.")
}
