# gh-secret-scanning - The GitHub Secret Scanning CLI Extension

This project is a GitHub CLI (`gh`) extension that provides commands for interacting with secret scanning alerts.

This extension helps GitHub Advanced Security (GHAS) customers prioritize remediation of their secret scanning alerts by identifying and focusing on those that are confirmed active first.

While this extension works for Enterprise Cloud (GHEC) customers, it is primarily intended for GitHub Enterprise Server (GHES) customers who do not have access to the [GitHub.com secret scanning validity check feature](https://docs.github.com/en/enterprise-cloud@latest/code-security/secret-scanning/managing-alerts-from-secret-scanning#validating-partner-patterns). Validity check on GHES is available as of `>=3.12` but currently limited to GitHub Personal Access Tokens (PAT).

Primary features include:

- Listing secret scanning alerts for an enterprise, organization, or repository
- Verifying if secret scanning alerts are still active
  - Expand the out-of-the-box secret scanning validity checks capabilities with custom validators
- Opening issues in repos that contain valid secrets

## Supported Token Types

- GitHub Personal Access Tokens (GHES + GHEC)
- Slack API Tokens

## Pre-requisites

- [GitHub CLI](https://github.com/cli/cli#installation)
- [GHES 3.7+](https://docs.github.com/en/enterprise-server@3.7/admin/all-releases#releases-of-github-enterprise-server) or [GHEC](https://docs.github.com/en/enterprise-cloud@latest/admin/overview/about-github-enterprise-cloud)
- [GHAS](https://docs.github.com/en/enterprise-cloud@latest/get-started/learning-about-github/about-github-advanced-security)

## Installation

```bash
gh extension install CallMeGreg/gh-secret-scanning
```

## Usage

Authenticate with your GitHub Enterprise Server or GitHub Enterprise Cloud account:

```bash
gh auth login
```

### Alerts subcommand

Target either an enterprise, organization, or repository by specifying the `-e`, `-o`, or `-r` flags respectively. _Exactly one selection from these three flags is required._

```bash
gh secret-scanning alerts -e <enterprise>
```

```bash
gh secret-scanning alerts -o <organization>
```

```bash
gh secret-scanning alerts -r <repository>
```

Optionally add flags to specify a GHES server, limit the number of secrets processed, filter for a specific secret provider, display the secret values, generate a csv report, include extra fields, and more:

```bash
gh secret-scanning alerts -e github --url my-github-server.com --limit 10 --provider slack --show-secret --csv --verbose
```

### Verify subcommand

Target either an enterprise, organization, or repository by specifying the `-e`, `-o`, or `-r` flags respectively. _Exactly one selection from these three flags is required._

```bash
gh secret-scanning verify -e <enterprise>
```

```bash
gh secret-scanning verify -o <organization>
```

```bash
gh secret-scanning verify -r <repository>
```

Optionally add flags to specify a GHES server, limit the number of secrets processed, filter for a specific secret provider, display the secret values, generate a csv report, include extra fields, and more:

```bash
gh secret-scanning verify -e github --url my-github-server.com --limit 10 --provider slack --show-secret --csv --verbose
```

Also, optionally create an issue in any repository that contains a valid secret by using the `--create-issues` (`-i`) flag:

```bash
gh secret-scanning verify -e github --url my-github-server.com --create-issues
```

### Help

See available commands and flags by running:

```bash
gh secret-scanning -h
```

```bash
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

## Demo

This example first lists the alerts for an organization with the `alerts` subcommand, and then verifies the secrets with the `verify` subcommand. The `--csv` flag is used to generate a csv report of the results, and the `--create-issues` flag is used to create issues in any repository that contains a valid secret.

https://github.com/CallMeGreg/gh-secret-scanning/assets/110078080/fa8d7b08-1a2c-4522-ae96-5c3aab60107d
