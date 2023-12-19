package cmd

// type Registry struct {
// 	Providers []Provider
// }

// type Provider struct {
// 	Name       string
// 	TokenTypes []TokenType
// }

// type TokenType struct {
// 	Name               string
// 	ValidationEndpoint string
// 	HttpMethod         string
// }

// func (r *Registry) GetProviders() string {
// 	names := make([]string, len(r.Providers))
// 	for i, provider := range r.Providers {
// 		names[i] = provider.Name
// 	}
// 	return strings.Join(names, ",")
// }

// func (p *Provider) GetTokenTypes() string {
// 	names := make([]string, len(p.TokenTypes))
// 	for i, tokenType := range p.TokenTypes {
// 		names[i] = tokenType.Name
// 	}
// 	return strings.Join(names, ",")
// }

// var MainRegistry = Registry{
// 	Providers: []Provider{
// 		Github_provider,
// 		Slack_provider,
// 	},
// }

// var Github_provider = Provider{
// 	Name: "github",
// 	TokenTypes: []TokenType{
// 		{
// 			Name:               "github_personal_access_token",
// 			ValidationEndpoint: "https://api.github.com",
// 			HttpMethod:         "GET",
// 		},
// 	},
// }

// var Slack_provider = Provider{
// 	Name: "slack",
// 	TokenTypes: []TokenType{
// 		{
// 			Name:               "slack_api_token",
// 			ValidationEndpoint: "https://slack.com/api/auth.test",
// 			HttpMethod:         "POST",
// 		},
// 	},
// }

var SupportedProviders = map[string]map[string]map[string]string{
	"github": {
		"github_personal_access_token": {
			"ValidationEndpoint": "https://api.github.com",
			"HttpMethod":         "GET",
			"ContentType":        "application/json",
			"ResponseFormat":     "json",
		},
	},
	"slack": {
		"slack_api_token": {
			"ValidationEndpoint": "https://slack.com/api/auth.test",
			"HttpMethod":         "POST",
			"ContentType":        "application/json",
		},
	},
}
