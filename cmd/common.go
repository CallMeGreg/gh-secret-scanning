package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
)

type User struct {
	Login string `json:"login"`
}

type Repository struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Full_name string `json:"full_name"`
}

type Alert struct {
	Number                      int        `json:"number"`
	Created_at                  string     `json:"created_at"`
	URL                         string     `json:"url"`
	HTML_URL                    string     `json:"html_url"`
	State                       string     `json:"state"`
	Resolution                  string     `json:"resolution"`
	Resolved_at                 string     `json:"resolved_at"`
	Resolved_by                 User       `json:"resolved_by"`
	Secret_type                 string     `json:"secret_type"`
	Secret_type_display_name    string     `json:"secret_type_display_name"`
	Secret                      string     `json:"secret"`
	Repository                  Repository `json:"repository"`
	Push_protection_bypassed    bool       `json:"push_protection_bypassed"`
	Push_protection_bypassed_at string     `json:"push_protection_bypassed_at"`
	Push_protection_bypassed_by User       `json:"push_protection_bypassed_by"`
	Validity_github             string     `json:"validity"`
	Validity_boolean            bool       `json:"validity_boolean"`
	Validity_response_code      string     `json:"validity_response_code"`
	Validity_endpoint           string     `json:"validity_endpoint"`
}

type HttpMethod int

const (
	GET HttpMethod = iota
	POST
	PUT
	DELETE
)

// template API URLs for the `list secret scanning alerts` endpoints (enterprise, organization, repository):
const (
	enterpriseAlertsURL   = "enterprises/{enterprise}/secret-scanning/alerts"
	organizationAlertsURL = "orgs/{org}/secret-scanning/alerts"
	repositoryAlertsURL   = "repos/{owner}/{repo}/secret-scanning/alerts"
)

func createGitHubSecretAlertsAPIPath(scope string, target string) (apiURL string, err error) {
	switch scope {
	case "enterprise":
		// replace {enterprise} with the target value:
		apiURL = strings.Replace(enterpriseAlertsURL, "{enterprise}", target, 1)
	case "organization":
		// replace {org} with the target value:
		apiURL = strings.Replace(organizationAlertsURL, "{org}", target, 1)
	case "repository":
		// split target into owner and repo and replace {owner} and {repo}:
		owner := strings.Split(target, "/")[0]
		repo := strings.Split(target, "/")[1]
		replacer := strings.NewReplacer("{owner}", owner, "{repo}", repo)
		apiURL = replacer.Replace(repositoryAlertsURL)
	default:
		err = fmt.Errorf("Invalid API target.")
	}
	return apiURL, err
}

func Red(s string) string {
	return "\x1b[31m" + s + "\x1b[m"
}

func Green(s string) string {
	return "\x1b[32m" + s + "\x1b[m"
}

func Yellow(s string) string {
	return "\x1b[33m" + s + "\x1b[m"
}

func Blue(s string) string {
	return "\x1b[34m" + s + "\x1b[m"
}

func Gray(s string) string {
	return "\x1b[90m" + s + "\x1b[m"
}

func setOptions() api.ClientOptions {
	var opts api.ClientOptions
	if quiet {
		opts = api.ClientOptions{
			Host:        host,
			Headers:     map[string]string{"Accept": "application/vnd.github+json"},
			LogColorize: true,
		}
	} else {
		opts = api.ClientOptions{
			Host:        host,
			Headers:     map[string]string{"Accept": "application/vnd.github+json"},
			Log:         os.Stdout,
			LogColorize: true,
		}
	}
	return opts
}

func callGitHubAPI(client *api.RESTClient, requestPath string, parseType interface{}, method HttpMethod, postBody ...[]byte) (int, string, error) {
	var httpMethod string
	switch method {
	case POST:
		httpMethod = http.MethodPost
	case PUT:
		httpMethod = http.MethodPut
	case DELETE:
		httpMethod = http.MethodDelete
	default:
		httpMethod = http.MethodGet
	}

	var body io.Reader
	if len(postBody) > 0 {
		body = bytes.NewReader(postBody[0])
	} else {
		body = nil
	}

	response, err := client.Request(httpMethod, requestPath, body)
	if err != nil {
		var httpError *api.HTTPError
		errors.As(err, &httpError)
		return httpError.StatusCode, "", err
	}

	defer response.Body.Close()
	nextPage := response.Header.Get("Link")
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("ERROR: Unable to read next page link")
		return response.StatusCode, nextPage, err
	}

	err = decodeJSONResponse(responseBody, &parseType)
	if err != nil {
		fmt.Println("ERROR: Unable to decode JSON response")
		return response.StatusCode, nextPage, err
	}

	return response.StatusCode, nextPage, nil
}

func decodeJSONResponse(body []byte, parseType interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	err := decoder.Decode(&parseType)
	if err != nil {
		fmt.Println("ERROR: Unable to decode JSON response")
		return err
	}

	return nil
}

func findNextPage(nextPageLink string) (string, bool) {
	var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)
	for _, m := range linkRE.FindAllStringSubmatch(nextPageLink, -1) {
		if len(m) > 2 && m[2] == "next" {
			return m[1], true
		}
	}
	return "", false
}

func validateProvider(provider string) (err error) {
	// get top level keys from SupportedProviders map:
	var providerList []string
	for key := range SupportedProviders {
		providerList = append(providerList, key)
	}
	for _, item := range providerList {
		if strings.ToLower(item) == strings.ToLower(provider) {
			return nil
		}
	}
	err = fmt.Errorf(Red("Invalid provider: " + provider + "\nValid providers are: " + strings.Join(providerList, ", ")))
	return err
}

func sortAlerts(alerts []Alert) []Alert {
	// sort alerts by repo name and then alert number
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].Repository.Full_name == alerts[j].Repository.Full_name {
			return alerts[i].Number < alerts[j].Number
		}
		return alerts[i].Repository.Full_name < alerts[j].Repository.Full_name
	})
	return alerts
}

func addRepoFullNameToAlerts(alerts []Alert) []Alert {
	// handle repo name for repo endpoint which doesn't return the repo name field
	for i := range alerts {
		alerts[i].Repository.Full_name = repository
	}
	return alerts
}

func getSecretTypeParameter() (secret_type_param string) {
	// if provider was specified, only return the secret types for that provider:
	secret_type_param = "all"
	if provider != "" {
		// get keys for SupportedProviders[provider] map:
		var secret_types []string
		for key := range SupportedProviders[provider] {
			secret_types = append(secret_types, key)
		}
		secret_type_param = strings.Join(secret_types, ",")
	}
	return secret_type_param
}

func getScopeAndTarget() (scope string, target string, err error) {
	if enterprise != "" {
		scope = "enterprise"
		target = enterprise
	} else if organization != "" {
		scope = "organization"
		target = organization
	} else if repository != "" {
		scope = "repository"
		target = repository
	}
	return scope, target, err
}

func prettyPrintAlerts(alerts []Alert, validity_check bool) (err error) {
	counter := 0
	if len(alerts) > 0 {
		terminal := term.FromEnv()
		termWidth, _, _ := terminal.Size()
		t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
		t.AddField("Repository", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		t.AddField("ID", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		t.AddField("State", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		t.AddField("Secret Type", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		t.AddField("Validity GitHub", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		if secret {
			t.AddField("Secret", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
		}
		if validity_check {
			t.AddField("Confirmed Valid", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Status Code", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Validity Endpoint", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
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

		for counter < len(alerts) && counter < limit {
			alert := alerts[counter]
			var color func(string) string
			if alert.Validity_boolean {
				color = Yellow
			} else {
				color = Gray
			}
			t.AddField(alert.Repository.Full_name, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
			t.AddField(strconv.Itoa(alert.Number), tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
			t.AddField(alert.State, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
			t.AddField(alert.Secret_type, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
			t.AddField(alert.Validity_github, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
			if secret {
				t.AddField(alert.Secret, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
			}
			if validity_check {
				t.AddField(strconv.FormatBool(alert.Validity_boolean), tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(alert.Validity_response_code, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(alert.Validity_endpoint, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
			}
			if verbose {
				t.AddField(alert.Created_at, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(alert.Resolution, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(alert.Resolved_at, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(alert.Resolved_by.Login, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(strconv.FormatBool(alert.Push_protection_bypassed), tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(alert.Push_protection_bypassed_at, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(alert.Push_protection_bypassed_by.Login, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
				t.AddField(alert.HTML_URL, tableprinter.WithColor(color), tableprinter.WithTruncate(nil))
			}
			t.EndRow()
			counter++
		}
		if err := t.Render(); err != nil {
			err = fmt.Errorf("error rendering table: %v", err)
			return err
		}
	}
	if limit < len(alerts) {
		fmt.Println(Blue("Fetched " + strconv.Itoa(limit) + " secret alerts."))
	} else {
		fmt.Println(Blue("Fetched " + strconv.Itoa(len(alerts)) + " secret alerts."))
	}
	return err
}

func generateCSVReport(alerts []Alert, scope string, validity_check bool) (err error) {
	fmt.Println(Blue("Generating CSV report..."))
	// reset counter
	counter := 0
	// get current date & time:
	now := time.Now()
	// Format the time as YYYYMMDD-HHMMSS
	timestamp := now.Format("20060102-150405")
	filename := "secretscanningreport-" + scope + "-" + timestamp + ".csv"
	// Create a CSV file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("ERROR: Error creating CSV file.")
		return err
	}
	defer file.Close()
	// Initialize CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// Write headers to CSV file
	headers := []string{"Repository", "ID", "State", "Secret Type", "GitHub Validity"}
	if secret {
		headers = append(headers, "Secret")
	}
	if validity_check {
		headers = append(headers, "Confirmed Valid", "Status Code", "Validation Endpoint")
	}
	if verbose {
		headers = append(headers, "Created At", "Resolution", "Resolved At", "Resolved By", "Push Protection Bypassed", "Push Protection Bypassed At", "Push Protection Bypassed By", "URL")
	}
	writer.Write(headers)
	// Write data to CSV file
	for counter < len(alerts) && counter < limit {
		alert := alerts[counter]
		row := []string{alert.Repository.Full_name, strconv.Itoa(alert.Number), alert.State, alert.Secret_type, alert.Validity_github}
		if secret {
			row = append(row, alert.Secret)
		}
		if validity_check {
			row = append(row, strconv.FormatBool(alert.Validity_boolean), alert.Validity_response_code, alert.Validity_endpoint)
		}
		if verbose {
			row = append(row, alert.Created_at, alert.Resolution, alert.Resolved_at, alert.Resolved_by.Login, strconv.FormatBool(alert.Push_protection_bypassed), alert.Push_protection_bypassed_at, alert.Push_protection_bypassed_by.Login, alert.HTML_URL)
		}
		writer.Write(row)
		counter++
	}
	if err := writer.Error(); err != nil {
		fmt.Println("ERROR: Error writing to CSV file.")
		return err
	}
	fmt.Println(Blue("CSV report generated: " + filename))
	return err
}

func verifyAlerts(alerts []Alert) (alertsOutput []Alert, err error) {
	// Print Supported providers for reference when verbose flag is enabled
	if verbose {
		fmt.Println(Blue("Supported Providers:"))
		for provider, secretTypes := range SupportedProviders {
			for secretType := range secretTypes {
				fmt.Printf("- %s - %s\n", provider, secretType)
			}
		}
	}
	for i, alert := range alerts {
		// verify that the alert is valid by making a request to its validation endpoint:
		provider := strings.Split(alert.Secret_type, "_")[0]

		// Skip alert if provider is not supported
		if _, ok := SupportedProviders[provider]; !ok {
			continue
		}

		secret_type := alert.Secret_type
		secret_validation_method := SupportedProviders[provider][secret_type]["HttpMethod"]
		secret_validation_content_type := SupportedProviders[provider][secret_type]["ContentType"]
		alert.Validity_endpoint = SupportedProviders[provider][secret_type]["ValidationEndpoint"]

		// create a new client for the validation request:
		var opts api.ClientOptions
		opts.AuthToken = alert.Secret
		opts.Headers = map[string]string{
			"User-Agent": "gh-secret-scanning",
		}
		client, err := api.NewHTTPClient(opts)
		if err != nil {
			fmt.Println("ERROR: Unable to create HTTP client.")
			return alerts, err
		}
		// send a request to the validation endpoint:
		var response *http.Response
		if secret_validation_method == "POST" {
			var body io.Reader
			req, err := http.NewRequest("POST", alert.Validity_endpoint, body)
			req.Header.Set("Authorization", "Bearer "+alert.Secret)
			req.Header.Set("Content-Type", secret_validation_content_type)
			req.Header.Set("User-Agent", "gh-secret-scanning")
			response, err = client.Do(req)
			// response, err = client.Post(alert.Validity_endpoint, secret_validation_content_type, body)
			if err != nil {
				fmt.Println("WARNING: Unable to send " + secret_validation_method + " request to " + alert.Validity_endpoint)
				continue
			}
			alert.Validity_response_code = strconv.Itoa(response.StatusCode)
			defer response.Body.Close()
		} else if secret_validation_method == "GET" {
			response, err = client.Get(alert.Validity_endpoint)
			if err != nil {
				fmt.Println("WARNING: Unable to send " + secret_validation_method + " request to " + alert.Validity_endpoint)
				continue
			}
			alert.Validity_response_code = strconv.Itoa(response.StatusCode)
		} else {
			fmt.Println("WARNING: Invalid HTTP method for validation endpoint for " + alert.Secret_type + " secret type.")
			continue
		}
		expected_body_key := SupportedProviders[provider][secret_type]["ExpectedBodyKey"]
		expected_body_value := SupportedProviders[provider][secret_type]["ExpectedBodyValue"]
		if expected_body_key != "" {
			alert.Validity_boolean = checkForExpectedBody(response, expected_body_key, expected_body_value, alert)
		} else if alert.Validity_response_code == "200" {
			alert.Validity_boolean = true
			if verbose {
				fmt.Println(Yellow("CONFIRMED: Alert " + strconv.Itoa(alert.Number) + " in " + alert.Repository.Full_name + " is valid."))
			}
		} else {
			alert.Validity_boolean = false
		}
		if provider == "github" && alert.Validity_boolean == false && host != "github.com" {
			// also confirm validity with the provided GitHub Enterprise Server API:
			alert = checkEnterpriseServerAPI(alert, client, secret_validation_method, secret_validation_content_type)
			if alert.Validity_response_code == "200" {
				alert.Validity_boolean = true
				if verbose {
					fmt.Println(Yellow("CONFIRMED: Alert " + strconv.Itoa(alert.Number) + " in " + alert.Repository.Full_name + " is valid."))
				}
			} else {
				alert.Validity_boolean = false
			}
		}
		alerts[i] = alert
	}
	return alerts, err
}

func checkForExpectedBody(response *http.Response, expected_body_key string, expected_body_value string, alert Alert) (validity_boolean bool) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("ERROR: Unable to read response body for alert " + strconv.Itoa(alert.Number) + " in " + alert.Repository.Full_name)
		return false
	}
	var response_body map[string]interface{}
	err = json.Unmarshal(body, &response_body)
	if err != nil {
		fmt.Println("ERROR: Unable to unmarshal response body for alert " + strconv.Itoa(alert.Number) + " in " + alert.Repository.Full_name)
		return false
	}
	body_value := response_body[expected_body_key]
	if body_value.(bool) {
		// convert response_value to a string
		body_value = strconv.FormatBool(body_value.(bool))
	}
	if body_value == expected_body_value {
		alert.Validity_boolean = true
		if verbose {
			fmt.Println(Yellow("CONFIRMED: Alert " + strconv.Itoa(alert.Number) + " in " + alert.Repository.Full_name + " is valid."))
		}
	} else {
		alert.Validity_boolean = false
	}
	return alert.Validity_boolean
}

func checkEnterpriseServerAPI(alert Alert, client *http.Client, secret_validation_method string, secret_validation_content_type string) (alertOutput Alert) {
	enterprise_server_api_endpoint := "https://" + host + "/api/v3/"
	// create a new http request:
	req, err := http.NewRequest("GET", enterprise_server_api_endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+alert.Secret)
	req.Header.Set("Content-Type", secret_validation_content_type)
	req.Header.Set("User-Agent", "gh-secret-scanning")
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("WARNING: Unable to send " + secret_validation_method + " request to " + alert.Validity_endpoint)
	}
	alert.Validity_response_code = strconv.Itoa(response.StatusCode)
	alert.Validity_endpoint = enterprise_server_api_endpoint
	return alert
}

func createIssuesForValidAlerts(alerts []Alert) (err error) {
	fmt.Println(Blue("Creating issues for valid alerts..."))
	issue_count := 0
	alertsByRepo := make(map[string][]Alert)
	for _, alert := range alerts {
		alertsByRepo[alert.Repository.Full_name] = append(alertsByRepo[alert.Repository.Full_name], alert)
	}
	for repo, alerts := range alertsByRepo {
		// check if there is at least one confirmed valid secret alert
		hasValidAlert := false
		for _, alert := range alerts {
			if alert.Validity_boolean {
				hasValidAlert = true
				break
			}
		}
		// if there is no confirmed valid secret alert, skip this repository
		if !hasValidAlert {
			continue
		}
		// create a string with the details of the alerts
		details := "**Please promptly revoke the following secrets and confirm in the provider's logs that they have not been used maliciously:**\n\n"
		details += "| Alert ID | Secret Type | Alert Link |\n"
		details += "| --- | --- | --- |\n"
		for _, alert := range alerts {
			if alert.Validity_boolean {
				details += fmt.Sprintf("| %d | %s | [Link](%s) |\n", alert.Number, alert.Secret_type, alert.HTML_URL)
			}
		}
		// create the issue
		var repo_with_host string
		if host == "" {
			repo_with_host = repo
		} else {
			repo_with_host = host + "/" + repo
		}
		_, _, err := gh.Exec("issue", "create", "--title", "IMMEDIATE ACTION REQUIRED: Active Secrets Detected", "--body", details, "--repo", repo_with_host)
		if err != nil {
			fmt.Println(err)
		} else {
			issue_count++
			if verbose {
				fmt.Println("Created issue in " + repo)
			}
		}
	}
	fmt.Println(Blue("Created " + strconv.Itoa(issue_count) + " issue(s)."))
	return err
}
