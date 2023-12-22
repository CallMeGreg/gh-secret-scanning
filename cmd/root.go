package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var host string
var enterprise string
var organization string
var repository string
var provider string
var limit int
var secret bool
var csvReport bool
var verbose bool
var quiet bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&host, "url", "u", "github.com", "GitHub host to connect to")
	rootCmd.PersistentFlags().StringVarP(&enterprise, "enterprise", "e", "", "GitHub enterprise slug")
	rootCmd.PersistentFlags().StringVarP(&organization, "organization", "o", "", "GitHub organization slug")
	rootCmd.PersistentFlags().StringVarP(&repository, "repository", "r", "", "GitHub owner/repository slug")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "", "Filter for a specific secret provider")
	rootCmd.PersistentFlags().IntVarP(&limit, "limit", "l", 30, "Limit the number of secrets processed")
	rootCmd.PersistentFlags().BoolVarP(&secret, "show-secret", "s", false, "Display secret values")
	rootCmd.PersistentFlags().BoolVarP(&csvReport, "csv", "c", false, "Generate a csv report of the results")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Include additional secret alert fields")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Minimize output to the console")

	// require exactly one (1) choice of enterprise, organization, or repository:
	rootCmd.MarkFlagsMutuallyExclusive("enterprise", "organization", "repository")
	rootCmd.MarkFlagsOneRequired("enterprise", "organization", "repository")

	// disable completion subcommand:
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		// warn user about --show-secret flag:
		if secret {
			fmt.Println(Yellow("WARNING: --show-secret flag is enabled. Full secret values will be displayed in PLAIN TEXT in the output. Would you like to continue? (y/n)"))
			var response string
			fmt.Scanln(&response)
			if response != "y" {
				log.Fatal("Exiting...")
			}
		}
		// check if provider is in supportedProviders:
		if provider != "" {
			return validateProvider(provider)
		}
		return
	}

}

var rootCmd = &cobra.Command{
	Use:   "secret-scanning <subcommand> [flags]",
	Short: "Interact with secret scanning alerts",
	Long:  "Interact with secret scanning alerts for a GHEC or GHES 3.7+ enterprise, organization, or repository",
}

func Root() {
	rootCmd.AddCommand(alertsCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.Execute()
}
