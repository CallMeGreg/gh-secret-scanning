package cmd

import (
	"github.com/spf13/cobra"
)

var host string
var enterprise string
var organization string
var repository string
var provider string
var limit int
var secret bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&host, "url", "u", "github.com", "GitHub host to connect to")
	rootCmd.PersistentFlags().StringVarP(&enterprise, "enterprise", "e", "", "GitHub enterprise slug")
	rootCmd.PersistentFlags().StringVarP(&organization, "organization", "o", "", "GitHub organization slug")
	rootCmd.PersistentFlags().StringVarP(&repository, "repository", "r", "", "GitHub owner/repository slug")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "", "Filter for a specific secret provider")
	rootCmd.PersistentFlags().IntVarP(&limit, "limit", "l", 20, "Limit the number of secrets processed")
	rootCmd.PersistentFlags().BoolVarP(&secret, "secret", "s", false, "Display secret value")

	// require exactly one (1) choice of enterprise, organization, or repository:
	rootCmd.MarkFlagsMutuallyExclusive("enterprise", "organization", "repository")
	rootCmd.MarkFlagsOneRequired("enterprise", "organization", "repository")

	// disable completion subcommand:
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

var rootCmd = &cobra.Command{
	Use:   "secret-scanning <subcommand> [flags]",
	Short: "Interact with secret scanning alerts",
	Long:  "Interact with secret scanning alerts for a GHEC or GHES 3.7+ enterprise, organization, or repository",
}

func Root() {
	rootCmd.AddCommand(alertsCmd)
	// TO DO: uncomment:
	// rootCmd.AddCommand(verifyCmd)
	rootCmd.Execute()
}
