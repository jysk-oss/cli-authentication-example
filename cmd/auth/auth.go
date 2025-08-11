package auth

import (
	"github.com/spf13/cobra"
)

func init() {
}

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication",
	Long:  "Authentication Sub Command",
}
