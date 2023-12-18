package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

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

type Alerts struct {
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

var ProviderTokenMapping = map[string][]string{
	"github": {"github_personal_access_token"},
	"slack":  {"slack_api_token"},
}

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

func callApi(requestPath string, parseType interface{}, method HttpMethod, postBody ...[]byte) (int, string, error) {
	opts := setOptions()
	client, err := api.NewRESTClient(opts)
	if err != nil {
		fmt.Println(err)
		return 0, "", err
	}

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
	for supportedProvider := range ProviderTokenMapping {
		if strings.ToLower(provider) == supportedProvider {
			fmt.Println(Blue("Filtering for alerts from:"), provider)
			return
		}
	}
	keys := make([]string, 0, len(ProviderTokenMapping))
	for k := range ProviderTokenMapping {
		keys = append(keys, k)
	}
	providers := strings.Join(keys, ", ")
	log.Fatal(Red("Invalid provider specified. Please choose from the list of supported providers: "), providers)
	return
}

func sortAlerts(alerts []Alerts) []Alerts {
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].Repository.Full_name == alerts[j].Repository.Full_name {
			return alerts[i].Number < alerts[j].Number
		}
		return alerts[i].Repository.Full_name < alerts[j].Repository.Full_name
	})
	return alerts
}
