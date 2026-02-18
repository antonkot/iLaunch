package cmd

import (
	"context"
	"fmt"

	"ilaunch/internal/app"

	"github.com/spf13/cobra"
)

type exitCodeError struct {
	code int
	err  error
}

func (e exitCodeError) Error() string { return e.err.Error() }
func (e exitCodeError) Unwrap() error { return e.err }
func (e exitCodeError) ExitCode() int { return e.code }

var nonInteractive bool

var rootCmd = &cobra.Command{
	Use:   "ilaunch",
	Short: "Interactive Node.js project bootstrap utility",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var code int
		var err error
		if nonInteractive {
			code, err = app.RunNonInteractive(ctx)
		} else {
			code, err = app.RunInteractive(ctx)
		}
		if err != nil {
			return exitCodeError{code: code, err: err}
		}
		if code != 0 {
			return exitCodeError{code: code, err: fmt.Errorf("execution finished with code %d", code)}
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "Run without TUI (CI mode)")
	rootCmd.SilenceUsage = true
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	type coder interface{ ExitCode() int }
	if c, ok := err.(coder); ok {
		return c.ExitCode()
	}
	return 1
}
