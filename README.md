# Overview
This project is a GitHub CLI (`gh`) extension that provides commands for interacting with secret scanning alerts. Primary uses include:
- Listing secret scanning alerts for an enterprise, organization, or repository
- Verifying if a secret is valid (for select providers)

# Pre-requisites
- [GitHub CLI](https://github.com/cli/cli#installation)
- GitHub Enterprise Server 3.7+ or GitHub Enterprise Cloud

# Installation

```bash
gh extension install CallMeGreg/gh-secret-scanning
```

# Usage
Authenticate to GitHub Enterprise Server or GitHub Enterprise Cloud using `gh auth login`

Then, run `gh secret-scanning --help` to see the available commands.