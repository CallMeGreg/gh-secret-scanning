package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts [flags]",
	Short: "Get secret scanning alerts for an enterprise, organization, or repository",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runAlerts(cmd, args)
	},
}

func runAlerts(cmd *cobra.Command, args []string) (err error) {
	opts := api.ClientOptions{
		Host:    host,
		Headers: map[string]string{"Accept": "application/vnd.github+json"},
		Log:     os.Stdout,
	}
	client, err := api.NewRESTClient(opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	// set scope & target based on the flag that was used:
	var scope string
	var target string
	if enterprise != "" {
		scope = "enterprise"
		target = enterprise
	} else if organization != "" {
		scope = "organization"
		target = organization
	} else if repository != "" {
		scope = "repository"
		target = repository
	} else {
		fmt.Println("No enterprise/organization/repository specified.")
		return
	}

	// set the API URL based on the target:
	apiURL, err := createGitHubSecretAlertsAPIPath(scope, target)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fetch secret data:
	response := []Alerts{}
	err = client.Get(apiURL, &response)
	if err != nil {
		fmt.Println(err)
		return
	}

	// TO DO: pretty print all of the response details:

	terminal := term.FromEnv()
	termWidth, _, _ := terminal.Size()
	t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
	green := func(s string) string {
		return "\x1b[32m" + s + "\x1b[m"
	}
	t.AddField("Repo", tableprinter.WithColor(green), tableprinter.WithTruncate(nil))
	t.AddField("ID", tableprinter.WithColor(green), tableprinter.WithTruncate(nil))
	t.AddField("State", tableprinter.WithColor(green), tableprinter.WithTruncate(nil))
	t.AddField("Secret Type", tableprinter.WithColor(green), tableprinter.WithTruncate(nil))
	if secret {
		t.AddField("Secret", tableprinter.WithColor(green), tableprinter.WithTruncate(nil))
	}
	t.EndRow()

	counter := 0
	for counter < len(response) && counter < limit {
		alert := response[counter]

		t.AddField(alert.Repository.Full_name, tableprinter.WithTruncate(nil))
		t.AddField(strconv.Itoa(alert.Number), tableprinter.WithTruncate(nil))
		t.AddField(alert.State, tableprinter.WithTruncate(nil))
		t.AddField(alert.Secret_type, tableprinter.WithTruncate(nil))
		if secret {
			t.AddField(alert.Secret, tableprinter.WithTruncate(nil))
		}
		t.EndRow()
		counter++
	}
	if err := t.Render(); err != nil {
		log.Fatal(err)
	}
	return
}
