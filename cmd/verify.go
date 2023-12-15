package cmd

import "github.com/spf13/cobra"

var createIssues bool

func init() {
	verifyCmd.PersistentFlags().BoolP("create-issues", "c", false, "Create issues in repos that contain verified secret alerts")
}

var verifyCmd = &cobra.Command{
	Use:   "verify [flags]",
	Short: "Verify alerts for an enterprise, organization, or repository",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runAlerts(cmd, args)
	},
}

func runVerify(target string, provider string) (err error) {
	// TO DO: get secret alerts for specified target (optionally filtered by provider):

	// TO DO: verify that the secret alerts are valid:

	// TO DO: optionally create issues in repos that contain verified secret alerts:

	return
}
