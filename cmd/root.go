// Package cmd holds the wtx command-line surface.
//
// Commands are built by constructor functions rather than declared as
// package-level variables: a cobra.Command carries mutable state (parsed flag
// values, output writers, the argument slice), and sharing one across
// invocations lets a flag set in one execution leak into the next. Tests
// depend on each execution starting clean.
package cmd

import "github.com/spf13/cobra"

// newRootCmd builds the `wtx` command and everything hanging off it.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "wtx",
		Short: "Tmux workspaces on top of wt-managed git worktrees",
		Long: `wtx manages tmux workspaces on top of wt-managed git worktrees,
with per-pane code-assistant layouts.

Run "wtx help <command>" for details on a command.`,
		Version: Version(),

		// Errors are reported once, by main. Without these, cobra prints the
		// error itself and follows it with a full usage dump on every
		// failure, including failures that have nothing to do with usage.
		SilenceErrors: true,
		SilenceUsage:  true,

		// Set explicitly rather than left nil. An unset Run also makes cobra
		// print help, but by a path that reports an error; bare `wtx` is not
		// a failure and must exit 0.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	// Cobra's default version template prefixes the program name. `wtx
	// version` prints the bare string, and the two are required to be
	// byte-identical.
	root.SetVersionTemplate("{{.Version}}\n")

	root.SetHelpCommand(newHelpCmd())
	root.AddCommand(newVersionCmd())

	return root
}

// Execute builds a fresh root command and runs it, returning any error for the
// caller to report and turn into an exit status.
func Execute() error {
	return newRootCmd().Execute()
}
