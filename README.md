# Overview
This project is a GitHub CLI (`gh`) extension that provides commands for interacting with secret scanning alerts. Primary uses include:
- Listing secret scanning alerts for an enterprise, organization, or repository
- Verifying if secret alerts are still active
- Opening issues in repos that contain valid secrets

# Supported Token Types
- GitHub Personal Access Tokens
- Slack API Tokens

# Pre-requisites
- [GitHub CLI](https://github.com/cli/cli#installation)
- GitHub Enterprise Server 3.7+ or GitHub Enterprise Cloud

# Installation
```
gh extension install CallMeGreg/gh-secret-scanning
```

# Usage
Authenticate with your GitHub Enterprise Server or GitHub Enterprise Cloud account:
```
gh auth login
```

## Alerts subcommand
Target either an enterprise, organization, or repository by specifying the `-e`, `-o`, or `-r` flags respectively. _Exactly one selection from these three flags is required._

```
gh secret-scanning alerts -e <enterprise>
```

```
gh secret-scanning alerts -o <organization>
```

```
gh secret-scanning alerts -r <repository>
```

Optionally add flags to specify a GHES server, limit the number of secrets processed, filter for a specific secret provider, display the secret values, generate a csv report, include extra fields, and more:
```
gh secret-scanning alerts -e github --url my-github-server.com --limit 10 --provider slack --show-secret --csv --verbose
```

## Verify subcommand
Target either an enterprise, organization, or repository by specifying the `-e`, `-o`, or `-r` flags respectively. _Exactly one selection from these three flags is required._

```
gh secret-scanning verify -e <enterprise>
```

```
gh secret-scanning verify -o <organization>
```

```
gh secret-scanning verify -r <repository>
```

Optionally add flags to specify a GHES server, limit the number of secrets processed, filter for a specific secret provider, display the secret values, generate a csv report, include extra fields, and more:
```
gh secret-scanning verify -e github --url my-github-server.com --limit 10 --provider slack --show-secret --csv --verbose
```

Also, optionally create issue in any repository that contains a valid secret using the `--create-issues` (`-i`) flag:
```
gh secret-scanning verify -e github --url my-github-server.com --create-issues
```


## Help
See available commands and flags by running:
```
gh secret-scanning -h
```

```
Interact with secret scanning alerts for a GHEC or GHES 3.7+ enterprise, organization, or repository

Usage:
  secret-scanning [command]

Available Commands:
  alerts      Get secret scanning alerts for an enterprise, organization, or repository
  help        Help about any command
  verify      Verify alerts for an enterprise, organization, or repository

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
  -v, --verbose               Include additional secret alert fields

Use "secret-scanning [command] --help" for more information about a command.
```

# Demo
This example first lists the alerts for an organization with the `alerts` subcommand, and then verifies the secrets with the `verify` subcommand. The `-c` flag is used to generate a csv report of the results, and the `-i` flag is used to create issues in any repository that contains a valid secret.

https://github.com/CallMeGreg/gh-secret-scanning/assets/110078080/fa8d7b08-1a2c-4522-ae96-5c3aab60107d

