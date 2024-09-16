package cmd

import (
	"fmt"
	"math"
	"net/url"
	"strconv"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

var createIssues bool

func init() {
	verifyCmd.PersistentFlags().BoolVarP(&createIssues, "create-issues", "i", false, "Create issues in repos that contain valid secret alerts")
}

var verifyCmd = &cobra.Command{
	Use:   "verify [flags]",
	Short: "Verify alerts for an enterprise, organization, or repository",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runVerify(cmd, args)
	},
}

func runVerify(cmd *cobra.Command, args []string) (err error) {
	// set scope & target based on the flag that was used:
	scope, target, err := getScopeAndTarget()
	if err != nil {
		return err
	}

	// set the API URL based on the target:
	requestPath, err := createGitHubSecretAlertsAPIPath(scope, target)
	if err != nil {
		return err
	}

	// update the URL to include query parameters based on specified flags:
	parsedURL, err := url.Parse(requestPath)
	if err != nil {
		return err
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
		return err
	}
	values.Set("per_page", per_page)
	// if provider was specified, filter results for just that provider. Otherwise, target all supported providers:
	secret_type := getSecretTypeParameter()
	if secret_type != "all" {
		values.Set("secret_type", secret_type)
		parsedURL.RawQuery = values.Encode()
	}

	// update the request path
	requestPath = parsedURL.String()

	// loop through calls to the API until all pages of results have been fetched or limit has been reached:
	var allSecretAlerts []Alert
	var pageOfSecretAlerts []Alert
	var pages = int(math.Ceil(float64(limit) / float64(per_page_int)))

	opts := setOptions()
	client, err := api.NewRESTClient(opts)
	if err != nil {
		return err
	}

	for page := 1; page <= pages; page++ {
		fmt.Println("Processing page: " + strconv.Itoa(page))
		_, nextPage, err := callGitHubAPI(client, requestPath, &pageOfSecretAlerts, GET)
		if err != nil {
			return err
		}
		// add each secret alert in the response page to allSecretAlerts array
		allSecretAlerts = append(allSecretAlerts, pageOfSecretAlerts...)
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

	// verify which secret alerts are confirmed valid:
	verifiedAlerts, err := verifyAlerts(sortedAlerts)
	if err != nil {
		fmt.Println("WARNING: issues encountered while sending verify requests.")
	}

	// pretty print with validity status
	if !quiet {
		prettyPrintAlerts(verifiedAlerts, true)
	}

	// optionally generate a csv report of the results:
	if len(sortedAlerts) > 0 && csvReport {
		err = generateCSVReport(sortedAlerts, scope, true)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// optionally create an issue for each repository that contains at least one valid secret alert:
	if createIssues {
		err = createIssuesForValidAlerts(verifiedAlerts)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return
}
