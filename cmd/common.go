package cmd

import (
	"log"
	"strings"
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
		log.Fatal("Invalid option")
	}
	return apiURL, err
}
