package auth

import (
	"github.com/kerwood/cli-authentication-example/internal/cliAuth"
	"github.com/spf13/cobra"
)

func init() {
	AuthCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with your user account",
	Run: func(cmd *cobra.Command, args []string) {
		cliAuth.Login()
	},
}
