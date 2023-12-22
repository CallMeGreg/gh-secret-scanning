package cmd

import (
	"fmt"
	"log"
	"math"
	"net/url"
	"strconv"

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

	var per_page string
	if limit < 100 {
		per_page = strconv.Itoa(limit)
	} else {
		per_page = "100"
	}
	per_page_int, err := strconv.Atoi(per_page)
	if err != nil {
		fmt.Println(err)
	}
	values.Set("per_page", per_page)
	// if provider was specified, filter results. Otherwise, return all results:
	var secret_type string
	if provider != "" {
		secret_type = getSecretTypeParameter()
	} else {
		secret_type = ""
	}
	values.Set("secret_type", secret_type)
	parsedURL.RawQuery = values.Encode()

	// update the request path
	requestPath = parsedURL.String()

	// loop through calls to the API until all pages of results have been fetched or limit has been reached:
	var allSecretAlerts []Alert
	var pageOfSecretAlerts []Alert
	var pages = int(math.Ceil(float64(limit) / float64(per_page_int)))

	opts := setOptions()
	client, err := api.NewRESTClient(opts)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for page := 1; page <= pages; page++ {
		log.Printf("Processing page: %d\n", page)
		_, nextPage, err := callGitHubAPI(client, requestPath, &pageOfSecretAlerts, GET)
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

	// if a specific repo endpoint was targeted, add the repo field to the alerts:
	if repository != "" {
		sortedAlerts = addRepoFullNameToAlerts(sortedAlerts)
	}

	// pretty print all of the response details:
	if !quiet {
		prettyPrintAlerts(sortedAlerts, false)
	}

	// optionally generate a csv report of the results:
	if len(sortedAlerts) > 0 && csvReport {
		err = generateCSVReport(sortedAlerts, scope, false)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return
}
