package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
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
	requestPath, err := createGitHubSecretAlertsAPIPath(scope, target)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Update the URL to include the per_page query parameter:
	parsedURL, err := url.Parse(requestPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set the query parameters
	values := parsedURL.Query()
	values.Set("per_page", per_page)
	secret_type := strings.Join(ProviderTokenMapping[provider], ",")
	values.Set("secret_type", secret_type)
	parsedURL.RawQuery = values.Encode()

	// Update the request path
	requestPath = parsedURL.String()

	// loop through calls to the API until all pages of results have been fetched or limit has been reached:
	var allSecretAlerts []Alerts
	var pageOfSecretAlerts []Alerts
	var pages = int(math.Ceil(float64(limit) / float64(per_page_int)))
	for page := 1; page <= pages; page++ {
		log.Printf("Processing page: %d\n", page)
		_, nextPage, err := callApi(requestPath, &pageOfSecretAlerts, GET)
		if err != nil {
			// check if the error is a 404
			log.Printf("ERROR: Unable to get alerts for target: %s\n", requestPath)
			return err
		}
		for _, secretAlert := range pageOfSecretAlerts {
			// add each secret alert in the response page to allSecretAlerts array
			allSecretAlerts = append(allSecretAlerts, secretAlert)
		}
		var hasNextPage bool
		if requestPath, hasNextPage = findNextPage(nextPage); !hasNextPage {
			break
		}
		if page*per_page_int >= limit {
			break
		}
	}

	// sort allSecretAlerts by repository name, and then by secret alert ID:
	sortedAlerts := sortAlerts(allSecretAlerts)

	// pretty print all of the response details:
	counter := 0
	if len(sortedAlerts) > 0 && !quiet {
		terminal := term.FromEnv()
		termWidth, _, _ := terminal.Size()
		t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
		t.AddField("Repository", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		t.AddField("ID", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		t.AddField("State", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		t.AddField("Secret Type", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		if secret {
			t.AddField("Secret", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		}
		if verbose {
			t.AddField("Created At", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Resolution", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Resolved At", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Resolved By", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Push Protection Bypassed", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Push Protection Bypassed At", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Push Protection Bypassed By", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("URL", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		}
		t.EndRow()

		for counter < len(sortedAlerts) && counter < limit {
			alert := sortedAlerts[counter]
			t.AddField(alert.Repository.Full_name, tableprinter.WithTruncate(nil))
			t.AddField(strconv.Itoa(alert.Number), tableprinter.WithTruncate(nil))
			t.AddField(alert.State, tableprinter.WithTruncate(nil))
			t.AddField(alert.Secret_type, tableprinter.WithTruncate(nil))
			if secret {
				t.AddField(alert.Secret, tableprinter.WithTruncate(nil))
			}
			if verbose {
				t.AddField(alert.Created_at, tableprinter.WithTruncate(nil))
				t.AddField(alert.Resolution, tableprinter.WithTruncate(nil))
				t.AddField(alert.Resolved_at, tableprinter.WithTruncate(nil))
				t.AddField(alert.Resolved_by.Login, tableprinter.WithTruncate(nil))
				t.AddField(strconv.FormatBool(alert.Push_protection_bypassed), tableprinter.WithTruncate(nil))
				t.AddField(alert.Push_protection_bypassed_at, tableprinter.WithTruncate(nil))
				t.AddField(alert.Push_protection_bypassed_by.Login, tableprinter.WithTruncate(nil))
				t.AddField(alert.HTML_URL, tableprinter.WithTruncate(nil))
			}
			t.EndRow()
			counter++
		}
		if err := t.Render(); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(Blue("Evaluated " + strconv.Itoa(len(sortedAlerts)) + " secret alerts."))

	// optionally generate a csv report of the results:
	if len(sortedAlerts) > 0 && csvReport {
		fmt.Println(Blue("Generating CSV report..."))
		// reset counter
		counter = 0
		// get current date & time:
		now := time.Now()
		timestamp := now.Format("2006-01-02_15-04-05")
		filename := "secret-scanning-report-" + timestamp + ".csv"
		if provider != "" {
			filename = "secret-scanning-report-" + provider + "-" + timestamp + ".csv"
		}
		// Create a CSV file
		file, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		// Initialize CSV writer
		writer := csv.NewWriter(file)
		defer writer.Flush()
		// Write headers to CSV file
		headers := []string{"Repository", "ID", "State", "Secret Type"}
		if secret {
			headers = append(headers, "Secret")
		}
		if verbose {
			headers = append(headers, "Created At", "Resolution", "Resolved At", "Resolved By", "Push Protection Bypassed", "Push Protection Bypassed At", "Push Protection Bypassed By", "URL")
		}
		writer.Write(headers)
		// Write data to CSV file
		for counter < len(sortedAlerts) && counter < limit {
			alert := sortedAlerts[counter]
			row := []string{alert.Repository.Full_name, strconv.Itoa(alert.Number), alert.State, alert.Secret_type}
			if secret {
				row = append(row, alert.Secret)
			}
			if verbose {
				row = append(row, alert.Created_at, alert.Resolution, alert.Resolved_at, alert.Resolved_by.Login, strconv.FormatBool(alert.Push_protection_bypassed), alert.Push_protection_bypassed_at, alert.Push_protection_bypassed_by.Login, alert.HTML_URL)
			}
			writer.Write(row)
			counter++
		}
		if err := writer.Error(); err != nil {
			log.Fatal(err)
		}
		fmt.Println(Blue("CSV report generated: " + filename))

	}
	return
}
