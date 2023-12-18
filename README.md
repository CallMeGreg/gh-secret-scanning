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

Add flags to specify a GHES server, limit the number of secrets processed, filter for a specific secret provider, display the secret values, and generate a csv report:
```bash
gh secret-scanning alerts -u callmegreg-02c6a8a1b5d81e463.ghe-test.org -e github -l 100 -p slack -s -c
```

## Help
See available commands and flags by running:
```bash
gh secret-scanning -h
```

```
Interact with secret scanning alerts for a GHEC or GHES 3.7+ enterprise, organization, or repository

Usage:
  secret-scanning [command]

Available Commands:
  alerts      Get secret scanning alerts for an enterprise, organization, or repository
  help        Help about any command

Flags:
  -c, --csv                   Generate a csv report of the results
  -e, --enterprise string     GitHub enterprise slug
  -h, --help                  help for secret-scanning
  -l, --limit int             Limit the number of secrets processed (default 30)
  -o, --organization string   GitHub organization slug
  -p, --provider string       Filter for a specific secret provider
  -q, --quiet                 Minimize output to the console
  -r, --repository string     GitHub owner/repository slug
  -s, --show-secret           Display secret values
  -u, --url string            GitHub host to connect to (default "github.com")
  -v, --verbose               Generate verbose output

Use "secret-scanning [command] --help" for more information about a command.
```