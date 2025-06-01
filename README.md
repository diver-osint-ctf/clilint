# clilint - CTF Challenges YAML Linter

A Go-based linter for [ctfcli](https://github.com/CTFd/ctfcli) challenges.yaml files with GitHub Actions integration and automatic PR commenting.

## Features

- ✅ **YAML Format Validation**: Ensures valid YAML syntax
- ✅ **File Existence Check**: Verifies files listed in `files` field exist
- ✅ **Welcome Requirements**: Validates welcome dependencies for non-welcome challenges
- ✅ **Field Validation**: Checks `image`, `state`, `version`, and `tags` fields
- 🚀 **GitHub Integration**: Automatic PR detection and commenting
- 🎯 **Smart Detection**: Only processes directories with changes

## GitHub Actions Usage

### Quick Setup

1. Add this workflow to `.github/workflows/lint.yml`:

```yaml
name: CTF Challenges YAML Linter

on:
  pull_request:
    paths: ["**/challenges.yaml"]
  issue_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  lint-challenges:
    if: >
      (github.event_name == 'pull_request') ||
      (github.event_name == 'issue_comment' && 
       github.event.issue.pull_request &&
       contains(github.event.comment.body, '@github clilint'))

    runs-on: ubuntu-latest

    steps:
      - name: Set PR number
        run: |
          if [[ "${{ github.event_name }}" == "pull_request" ]]; then
            echo "PR_NUMBER=${{ github.event.number }}" >> $GITHUB_ENV
          else
            echo "PR_NUMBER=${{ github.event.issue.number }}" >> $GITHUB_ENV
          fi

      - name: Get Branch
        run: |
          if [[ "${{ github.event_name }}" == "pull_request" ]]; then
            echo "BRANCH=${{ github.event.pull_request.head.ref }}" >> $GITHUB_ENV
          else
            BRANCH_NAME=$(gh pr view ${{ env.PR_NUMBER }} --json headRefName --jq .headRefName --repo ${{ github.repository }})
            echo "BRANCH_NAME=${BRANCH_NAME}" >> $GITHUB_ENV
          fi
        env:
          GH_TOKEN: ${{ github.token }}

      - uses: actions/checkout@v4
        with:
          ref: ${{ env.BRANCH_NAME }}

      - name: Run CTF Challenges YAML Linter
        uses: diver-osint-ctf/clilint@v0.1.3
        with:
          repository: ${{ github.repository }}
          pr-number: ${{ env.PR_NUMBER }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

2. The linter automatically:
   - ✅ Detects changed directories with `challenges.yaml` files
   - ✅ Lints only affected challenges
   - ✅ Posts detailed results as PR comments
   - ✅ Triggers on PR changes or `@github clilint` comments

## Local Usage

```bash
# Install
go install github.com/diver-osint-ctf/clilint@latest

# Lint current directory
clilint

# Lint specific directories
clilint web osint crypto

# JSON output
clilint --json

# Help
clilint -h
```

## Validation Rules

| Rule                   | Description                                                           |
| ---------------------- | --------------------------------------------------------------------- |
| **YAML Format**        | Must be valid YAML syntax                                             |
| **File Existence**     | All files in `files[]` must exist                                     |
| **Welcome Dependency** | Non-welcome challenges must include "welcome" in `requirements[]`     |
| **Image Field**        | Must be `null`                                                        |
| **State Field**        | Must be `"visible"`                                                   |
| **Version Field**      | Must be `"0.1"`                                                       |
| **Tags Field**         | Must contain exactly one of: `introduction`, `easy`, `medium`, `hard` |

## Example challenges.yaml

```yaml
name: "web_challenge"
author: "author_name"
category: "web"
description: "Challenge description with **markdown** support"
flags: ["flag{example}"]
tags: ["medium"]
files: ["public/challenge.zip"]
requirements: ["welcome"]
value: 500
type: dynamic
extra:
  initial: 500
  decay: 100
  minimum: 100
image: null
host: null
state: visible
version: "0.1"
```

## PR Comment Example

The linter posts rich markdown comments:

```markdown
## 🎉 CTF Challenges YAML Linting Results

✅ All affected challenges.yaml files passed linting!

### 📋 Checked Challenges in This PR:

#### 🚩 **web_challenge** (`web/chall1/challenges.yaml`)

Challenge description with **markdown** support

---

✨ Great job! All challenges.yaml files follow the required format.
```

## Development

```bash
# Clone and build
git clone https://github.com/diver-osint-ctf/clilint.git
cd clilint
go build -o clilint .

# Run tests
go test -v

# Dependencies
go mod tidy
```

## Contributing

Contributions welcome! Please submit a Pull Request.

## License

MIT License
