# clilint - CTF Challenges YAML Linter

A Go-based linter for [ctfcli](https://github.com/CTFd/ctfcli) challenges.yaml files, designed to run on GitHub Actions.

## Features

This linter validates the following aspects of `challenges.yaml` files:

- ✅ **YAML Format**: Ensures the file is valid YAML
- ✅ **File Existence**: Verifies that files specified in the `files` field actually exist
- ✅ **Welcome Requirements**: Checks that challenges without "welcome" in their name have "welcome" in their requirements
- ✅ **Null Fields**: Ensures `image` and `host` fields are set to `null`
- ✅ **State Validation**: Confirms `state` is set to "visible"
- ✅ **Version Check**: Validates that `version` is set to "0.1"
- ✅ **Tag Validation**: Ensures exactly one difficulty tag is present (`introduction`, `easy`, `medium`, or `hard`)
- 🚀 **GitHub Integration**: Native PR comment posting with markdown support
- 🎯 **Genre Filtering**: Automatically filters directories based on `config.yaml` genre list

## Configuration

### config.yaml

The linter uses a `config.yaml` file to define valid challenge genres. Only directories matching these genre names will be processed.

```yaml
genre:
  - web
  - misc
  - osint
  - crypto
  - pwn
```

If a `config.yaml` file is not found or cannot be read, the linter will proceed without genre validation and process all directories.

## Installation & Usage

### GitHub Actions Integration (Copy Sample Workflow)

This repository includes a sample workflow file (`.github/workflows/lint.yml`) that you can copy to your repository. The workflow will automatically:

- **Run on Pull Requests**: When `challenges.yaml` files are modified
- **Run on Command**: When someone comments `@github clilint` on a Pull Request
- **Smart Detection**: Only lints directories that contain changes
- **Rich Comments**: Posts detailed results including challenge descriptions in markdown format
- **Genre Filtering**: Only processes directories listed in `config.yaml`

#### Setup Instructions

1. Copy `.github/workflows/lint.yml` to your repository
2. Create a `config.yaml` file in your repository root with your challenge genres
3. The linter will automatically run on:
   - Pull Requests that modify `challenges.yaml` files
   - When someone comments `@github clilint` on a Pull Request

#### Sample Workflow Features

- 🎯 **Smart Change Detection**: Only processes directories with actual changes
- 📝 **Markdown Display**: Shows challenge descriptions in full markdown format
- 🚀 **Go-powered Comments**: Uses native Go GitHub API for efficient PR commenting
- ⚡ **Streamlined**: Single binary handles linting and commenting
- 🎪 **Genre Aware**: Respects `config.yaml` genre restrictions

### Local Usage

1. Clone this repository
2. Build the binary:
   ```bash
   go build -o clilint .
   ```
3. Run the linter:
   ```bash
   ./clilint [directory]
   ```

## Usage

### Command Line

```bash
# Lint all challenges.yaml files in current directory and subdirectories
./clilint

# Lint all challenges.yaml files in a specific directory
./clilint /path/to/challenges

# Lint multiple directories
./clilint osint web crypto

# Output in JSON format
./clilint --json

# Post results as PR comment (requires GitHub environment variables)
./clilint --comment-pr

# Use custom config file
./clilint --config my-config.yaml

# Show help
./clilint -h
```

### GitHub Environment Variables

When using `--comment-pr`, the following environment variables are required:

- `GITHUB_TOKEN`: GitHub personal access token or `${{ secrets.GITHUB_TOKEN }}`
- `GITHUB_REPOSITORY`: Repository in format `owner/repo` or `${{ github.repository }}`
- `PR_NUMBER`: Pull request number (optional for issue comments)

## Expected Directory Structure

The linter expects the following directory structure:

```
repository/
├── config.yaml          # Genre definitions
├── genre_1/             # Must match config.yaml genres
│   ├── chall_1_1/
│   │   ├── challenges.yaml
│   │   └── public/
│   │       └── challenge_files...
│   ├── chall_1_2/
│   │   └── challenges.yaml
│   └── ...
├── genre_2/             # Must match config.yaml genres
│   ├── chall_2_1/
│   │   └── challenges.yaml
│   └── ...
└── ...
```

## Genre Filtering Example

With this `config.yaml`:

```yaml
genre:
  - web
  - osint
  - misc
```

The linter will process:

- ✅ `web/chall_1/challenges.yaml`
- ✅ `osint/chall_2/challenges.yaml`
- ✅ `misc/chall_3/challenges.yaml`

But skip:

- ❌ `crypto/chall_1/challenges.yaml` (not in config.yaml)
- ❌ `invalid_genre/chall_1/challenges.yaml` (not in config.yaml)

## Example challenges.yaml

Here's an example of a valid `challenges.yaml` file:

```yaml
name: "example_challenge"
author: "author_name"
category: "web"
description: |
  This is an example challenge description.

  You can use **markdown** formatting here!

  - List items
  - [Links](https://example.com)
  - Code blocks: `flag{example}`
flags:
  - "flag{example_flag}"
tags:
  - medium
files:
  - public/example.txt
requirements:
  - "welcome"
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

## Sample PR Comment Output

When linting succeeds, the GitHub Action will post a comment like this:

```markdown
## 🎉 CTF Challenges YAML Linting Results

✅ All affected challenges.yaml files passed linting!

### 📋 Checked Challenges in This PR:

#### 🚩 **sample_chall** (`osint/chall_1/challenges.yaml`)

sample\_**description**

[sample](https://example.com)

---

✨ Great job! All challenges.yaml files in the changed directories follow the required format and standards.
```

## Testing

Run the test suite:

```bash
go test -v
```

## Linting Rules

### 1. YAML Format

The file must be valid YAML syntax.

### 2. File Existence

All files listed in the `files` field must exist relative to the directory containing the `challenges.yaml` file.

### 3. Welcome Requirements

If the challenge name doesn't contain "welcome" (case-insensitive), the `requirements` field must include "welcome".

### 4. Null Fields

Both `image` and `host` fields must be set to `null`.

### 5. State Field

The `state` field must be set to "visible".

### 6. Version Field

The `version` field must be set to "0.1".

### 7. Tags Field

The `tags` field must contain exactly one of the following difficulty tags:

- `introduction`
- `easy`
- `medium`
- `hard`

### 8. Genre Validation

Directories must match one of the genres defined in `config.yaml`. If no config file is found, this validation is skipped.

## Error Examples

```
❌ genre1/chall1/challenges.yaml:
  - File specified in 'files' does not exist: missing.txt
  - Challenges without 'welcome' in name must have 'welcome' in requirements
  - Field 'image' should be null
  - Field 'version' should be '0.1'
  - Tags should contain exactly one of: introduction, easy, medium, hard

Skipping crypto/chall1/challenges.yaml: genre 'crypto' not found in config.yaml
```

## Development

### Building

```bash
go build -o clilint .
```

### Running Tests

```bash
go test -v
```

### Dependencies

- Go 1.21+
- `gopkg.in/yaml.v3` for YAML parsing
- `github.com/google/go-github/v65/github` for GitHub API integration
- `golang.org/x/oauth2` for OAuth2 authentication

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
