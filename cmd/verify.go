package cmd

import (
	"fmt"
	"log"
	"math"
	"net/url"

	"github.com/spf13/cobra"
)

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
	// set scope & target based on the flag that was used:
	scope, target, err := getScopeAndTarget()
	if err != nil {
		fmt.Println(err)
		return
	}

	// set the API URL based on the target:
	requestPath, err := createGitHubSecretAlertsAPIPath(scope, target)
	if err != nil {
		fmt.Println(err)
		return
	}

	// update the URL to include query parameters based on specified flags:
	parsedURL, err := url.Parse(requestPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	values := parsedURL.Query()
	values.Set("per_page", per_page)
	// if provider was specified, filter results for just that provider. Otherwise, target all supported providers:
	secret_type := getSecretTypeParameter()
	values.Set("secret_type", secret_type)
	parsedURL.RawQuery = values.Encode()

	// update the request path
	requestPath = parsedURL.String()

	// loop through calls to the API until all pages of results have been fetched or limit has been reached:
	var allSecretAlerts []Alert
	var pageOfSecretAlerts []Alert
	var pages = int(math.Ceil(float64(limit) / float64(per_page_int)))
	for page := 1; page <= pages; page++ {
		log.Printf("Processing page: %d\n", page)
		_, nextPage, err := callApi(requestPath, &pageOfSecretAlerts, GET)
		if err != nil {
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

	// TO DO: verify that the secret alerts are valid:
	for _, alert := range sortedAlerts {
		verifyAlert(alert)
		// TO DO: optionally create issues in repos that contain verified secret alerts:
	}

	// TO DO: optionally create issues in repos that contain verified secret alerts:

	return
}
