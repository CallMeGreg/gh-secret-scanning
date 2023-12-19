package cmd

import "strings"

type Registry struct {
	Providers []Provider
}

type Provider struct {
	Name       string
	TokenTypes []TokenType
}

type TokenType struct {
	Name               string
	ValidationEndpoint string
	HttpMethod         string
}

func (r *Registry) GetProviders() string {
	names := make([]string, len(r.Providers))
	for i, provider := range r.Providers {
		names[i] = provider.Name
	}
	return strings.Join(names, ",")
}

func (p *Provider) GetTokenTypes() string {
	names := make([]string, len(p.TokenTypes))
	for i, tokenType := range p.TokenTypes {
		names[i] = tokenType.Name
	}
	return strings.Join(names, ",")
}

var MainRegistry = Registry{
	Providers: []Provider{
		GitHub,
		Slack,
	},
}

var GitHub = Provider{
	Name: "github",
	TokenTypes: []TokenType{
		{
			Name:               "github_personal_access_token",
			ValidationEndpoint: "https://api.github.com",
			HttpMethod:         "GET",
		},
		{
			Name:               "github_app_installation_access_token",
			ValidationEndpoint: "https://api.github.com",
			HttpMethod:         "GET",
		},
		{
			Name:               "github_oauth_access_token",
			ValidationEndpoint: "https://api.github.com",
			HttpMethod:         "GET",
		},
	},
}

var Slack = Provider{
	Name: "slack",
	TokenTypes: []TokenType{
		{
			Name:               "slack_api_token",
			ValidationEndpoint: "https://slack.com/api/auth.test",
			HttpMethod:         "POST",
		},
	},
}
