# Overview
This project is a GitHub CLI (`gh`) extension that provides commands for interacting with secret scanning alerts. Primary uses include:
- Listing secret scanning alerts for an enterprise, organization, or repository
- Verifying if a secret is valid (for select providers)
- Opening issues in repos that contain verified secrets

# Pre-requisites
- [GitHub CLI](https://github.com/cli/cli#installation)
- GitHub Enterprise Server 3.7+ or GitHub Enterprise Cloud

# Installation
```bash
gh extension install CallMeGreg/gh-secret-scanning
```

# Usage
Authenticate with your GitHub Enterprise Server or GitHub Enterprise Cloud account:
```bash
gh auth login
```

## Alerts subcommand
List secret scanning alerts for an enterprise:
```bash
gh secret-scanning alerts -e <enterprise>
```

List secret scanning alerts for an organization:
```bash
gh secret-scanning alerts -o <organization>
```

List secret scanning alerts for a repository:
```bash
gh secret-scanning alerts -r <repository>
```

## Verify subcommand

:construction: This command is under development and not yet available.

## Help
See available commands and flags by running:
```bash
gh secret-scanning -h
```

```
Usage:
  secret-scanning [command]

Available Commands:
  alerts      Get secret scanning alerts for an enterprise, organization, or repository
  help        Help about any command

Flags:
  -e, --enterprise string     GitHub enterprise slug
  -h, --help                  help for secret-scanning
  -l, --limit int             Limit the number of secrets processed (default 20)
  -o, --organization string   GitHub organization slug
  -p, --provider string       Filter for a specific secret provider
  -r, --repository string     GitHub owner/repository slug
  -s, --secret                Display secret value
  -u, --url string            GitHub host to connect to (default "github.com")

Use "secret-scanning [command] --help" for more information about a command.
```