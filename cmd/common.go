package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
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
	Validity_boolean            bool       `json:"validity_boolean"`
	Validity_response_code      string     `json:"validity_response_code"`
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

var per_page string = "100"
var per_page_int int = 100

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
		log.Fatal("Invalid API target.")
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

func callApi(client *api.RESTClient, requestPath string, parseType interface{}, method HttpMethod, postBody ...[]byte) (int, string, error) {
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
		log.Println("ERROR: Unable to read next page link")
		return response.StatusCode, nextPage, err
	}

	err = decodeJSONResponse(responseBody, &parseType)
	if err != nil {
		log.Println("ERROR: Unable to decode JSON response")
		return response.StatusCode, nextPage, err
	}

	return response.StatusCode, nextPage, nil
}

func decodeJSONResponse(body []byte, parseType interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	err := decoder.Decode(&parseType)
	if err != nil {
		log.Println("ERROR: Unable to decode JSON response")
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
	if provider != "" {
		// get keys for SupportedProviders[provider] map:
		var secret_types []string
		for key := range SupportedProviders[provider] {
			secret_types = append(secret_types, key)
		}
		secret_type_param = strings.Join(secret_types, ",")
	} else {
		// get keys for every SupportedProvider key:
		var secret_types []string
		for _, supported_provider := range SupportedProviders {
			for key := range supported_provider {
				secret_types = append(secret_types, key)
			}
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

func prettyPrintAlerts(alerts []Alert, validity_check bool) {
	counter := 0
	if len(alerts) > 0 {
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
		if validity_check {
			t.AddField("Confirmed Valid", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
			t.AddField("Validity Check Status Code", tableprinter.WithColor(Green), tableprinter.WithTruncate(nil))
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
			t.AddField(alert.Repository.Full_name, tableprinter.WithTruncate(nil))
			t.AddField(strconv.Itoa(alert.Number), tableprinter.WithTruncate(nil))
			t.AddField(alert.State, tableprinter.WithTruncate(nil))
			t.AddField(alert.Secret_type, tableprinter.WithTruncate(nil))
			if secret {
				t.AddField(alert.Secret, tableprinter.WithTruncate(nil))
			}
			if validity_check {
				t.AddField(strconv.FormatBool(alert.Validity_boolean), tableprinter.WithTruncate(nil))
				t.AddField(alert.Validity_response_code, tableprinter.WithTruncate(nil))
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
	if limit < len(alerts) {
		fmt.Println(Blue("Fetched " + strconv.Itoa(limit) + " secret alerts."))
	} else {
		fmt.Println(Blue("Fetched " + strconv.Itoa(len(alerts)) + " secret alerts."))
	}
}

func generateCSVReport(alerts []Alert, scope string) (err error) {
	fmt.Println(Blue("Generating CSV report..."))
	// reset counter
	counter := 0
	// get current date & time:
	now := time.Now()
	timestamp := now.Format("2006-01-02 15-04-05")
	filename := "Secret Scanning Report - " + scope + " - " + timestamp + ".csv"
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
	for counter < len(alerts) && counter < limit {
		alert := alerts[counter]
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
	return err
}

func verifyAlert(alert Alert) (status_code int, err error) {
	// verify that the alert is valid by making a request to its validation endpoint:
	provider := strings.Split(alert.Secret_type, "_")[0]
	secret_type := alert.Secret_type
	secret_validation_endpoint := SupportedProviders[provider][secret_type]["ValidationEndpoint"]
	secret_validation_method := SupportedProviders[provider][secret_type]["HttpMethod"]
	secret_validation_content_type := SupportedProviders[provider][secret_type]["ContentType"]
	auth_header := "Bearer " + alert.Secret

	// create a new client for the validation request:
	var opts api.ClientOptions
	opts.AuthToken = auth_header
	client, err := api.NewHTTPClient(opts)
	if err != nil {
		log.Println("ERROR: Unable to create HTTP client")
		return 0, err
	}
	// send a request to the validation endpoint:
	var response *http.Response
	var body io.Reader
	if secret_validation_method == "POST" {
		response, err := client.Post(secret_validation_endpoint, secret_validation_content_type, body)
		if err != nil {
			log.Println("ERROR: Unable to send " + secret_validation_method + " request to " + secret_validation_endpoint)
			return response.StatusCode, err
		}
		defer response.Body.Close()
	} else if secret_validation_method == "GET" {
		response, err := client.Get(secret_validation_endpoint)
		if err != nil {
			log.Println("ERROR: Unable to send " + secret_validation_method + " request to " + secret_validation_endpoint)
			return response.StatusCode, err
		}
		return response.StatusCode, nil
	} else {
		log.Println("ERROR: Invalid HTTP method for validation endpoint")
		return 0, err
	}
	return response.StatusCode, err
}
