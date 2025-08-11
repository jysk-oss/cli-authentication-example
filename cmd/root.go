/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kerwood/cli-authentication-example/cmd/auth"
	"github.com/kerwood/cli-authentication-example/cmd/termpad"
	"github.com/kerwood/cli-authentication-example/internal/cliAuth"
	"github.com/kerwood/cli-authentication-example/internal/contextkeys"
	"github.com/spf13/cobra"
)

var version = "0.1.0"
var showVersion bool

var subCommandScopes = map[string]string{
	"termpad": "api://termpad/access",
}

var rootCmd = &cobra.Command{
	Use:   "cli-example",
	Short: "A brief description of your application",
	Long:  `A longer description that spans multiple lines`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Printf("butters v%s\n", version)
			os.Exit(0)
		}

		// Splits the full command path into a slice of strings.
		// If the slice has fewer than two elements, return and skip the rest.
		pathParts := strings.Split(cmd.CommandPath(), " ")
		if len(pathParts) < 2 {
			return nil
		}

		// The second element is the subcommand name. (termpad)
		cmdIdentifier := pathParts[1]

		// If the subcommand name exists in the subCommandScopes map
		// assign the scope to the scope variable. Else return and skip the rest.
		scope, hasScope := subCommandScopes[cmdIdentifier]
		if !hasScope || scope == "" {
			return nil
		}

		// Load the token store.
		// It reads the tokens.json file and deserialize into the TokenStore object.
		tokenStore, err := cliAuth.LoadTokenStore()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Request a delegated token for the subcommand from the TokenStore,
		// using the subcommand name and its associated scope.
		// If a valid delegated token already exists in the store, it will be returned.
		// Otherwise, the TokenStore will use the refresh token from your login
		// to request a new delegated token.
		token, err := tokenStore.GetDelegatedToken(cmdIdentifier, scope)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// The access token is saved into the command's context,
		// allowing subcommands to access it and use it for API calls.
		ctx := context.WithValue(cmd.Context(), ctxkeys.AccessToken, token.AccessToken)
		cmd.SetContext(ctx) // Update the context on the command

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		// No subcommand or version flag: show help
		if !cmd.Flags().Changed("version") {
			cmd.Help()
		}
	}}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version")
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(termpad.Commands)
}
