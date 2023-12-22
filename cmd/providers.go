package cmd

var SupportedProviders = map[string]map[string]map[string]string{
	"github": {
		"github_personal_access_token": {
			"ValidationEndpoint": "https://api.github.com",
			"HttpMethod":         "GET",
			"ContentType":        "application/vnd.github.v3+json",
			"ExpectedBodyKey":    "",
			"ExpectedBodyValue":  "",
		},
	},
	"slack": {
		"slack_api_token": {
			"ValidationEndpoint": "https://slack.com/api/auth.test",
			"HttpMethod":         "POST",
			"ContentType":        "application/json",
			"ExpectedBodyKey":    "ok",
			"ExpectedBodyValue":  "true",
		},
	},
}
