package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// newHelpCmd replaces cobra's built-in help command.
//
// The built-in one prints "Unknown help topic" for a command that does not
// exist and then exits 0, so `wtx help no-such-command` looks successful to a
// script. This one returns an error instead. Help for a command that does
// exist behaves as it does upstream.
func newHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any wtx command.

Type "wtx help [path to command]" for full details.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _, err := cmd.Root().Find(args)
			if target == nil || err != nil {
				return fmt.Errorf("unknown help topic %q", strings.Join(args, " "))
			}

			// Make the target's own --help and --version flags visible in the
			// usage text, as cobra's built-in help command does.
			target.InitDefaultHelpFlag()
			target.InitDefaultVersionFlag()

			return target.Help()
		},
	}
}
