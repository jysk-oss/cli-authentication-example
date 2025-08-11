package termpad

import (
	"strings"

	"github.com/kerwood/cli-authentication-example/internal/contextkeys"
	"github.com/kerwood/cli-authentication-example/internal/termpad"
	"github.com/spf13/cobra"
)

var Commands = &cobra.Command{
	Use:   "termpad",
	Short: "Post or get code from termpad.",
}

func init() {
	Commands.AddCommand(postCmd)
	Commands.AddCommand(getCmd)
}

var postCmd = &cobra.Command{
	Use:                   "post [string]",
	Short:                 "Post code to termpad",
	Example:               "  cli-example termpad post \"Hello world\"",
	Args:                  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		token := cmd.Context().Value(ctxkeys.AccessToken).(string)
		input := strings.Join(args, " ")
		termpad.PostData(token, input)
	},
}

var getCmd = &cobra.Command{
	Use:                   "get [termpad-identifier]",
	Short:                 "Get code from termpad",
	Example:               "  cli-example termpad get WorrisomeFriendlyJewellery",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		token := cmd.Context().Value(ctxkeys.AccessToken).(string)
		termpad.GetData(token, args[0])
	},
}
