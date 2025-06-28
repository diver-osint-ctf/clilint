# clilint - CTF Challenges YAML Linter

A Go-based linter for [ctfcli](https://github.com/CTFd/ctfcli) challenge.yml files with GitHub Actions integration and automatic PR commenting.

## Features

- ✅ **YAML Format Validation**: Ensures valid YAML syntax
- ✅ **File Existence Check**: Verifies files listed in `files` field exist
- ✅ **File Size**: Checks that all files in `files[]` are 1.00 MB or smaller
- ✅ **Welcome Requirements**: Validates welcome dependencies for non-welcome challenges
- ✅ **Field Validation**: Checks `image`, `state`, `version`, and `tags` fields
- 🚀 **GitHub Integration**: Automatic PR detection and commenting
- 🎯 **Smart Detection**: Only processes directories with changes

## GitHub Actions Usage

### Quick Setup

1. Add this workflow to [`.github/workflows/lint.yml`](./.github/workflows/lint.yml):

2. The linter automatically:
   - ✅ Detects changed directories with `challenge.yml` files
   - ✅ Lints only affected challenges
   - ✅ Posts detailed results as PR comments
   - ✅ Triggers on PR changes or `@github clilint` comments

## Validation Rules

| Rule                   | Description                                                           |
| ---------------------- | --------------------------------------------------------------------- |
| **YAML Format**        | Must be valid YAML syntax                                             |
| **File Existence**     | All files in `files[]` must exist                                     |
| **File Size**          | All files in `files[]` must be 1.00 MB or smaller                     |
| **Welcome Dependency** | Non-welcome challenges must include "welcome" in `requirements[]`     |
| **Image Field**        | Must be `null`                                                        |
| **State Field**        | Must be `"visible"`                                                   |
| **Version Field**      | Must be `"0.1"`                                                       |
| **Tags Field**         | Must contain exactly one of: `introduction`, `easy`, `medium`, `hard` |

## Example challenge.yml

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

## Example lintrc.yaml

[lintrc.yaml](./lintrc.yaml)

```yaml
tags:
  condition: and
  patterns:
    - type: regex
      values:
        - "author: *"
    - type: static
      values:
        - easy
        - medium
        - hard
```

## PR Comment Example

The linter posts rich markdown comments:

```markdown
## 🎉 CTF Challenges YAML Linting Results

✅ All affected challenge.yml files passed linting!

### 📋 Checked Challenges in This PR:

#### 🚩 **web_challenge** (`web/chall1/challenge.yml`)

Challenge description with **markdown** support

---

✨ Great job! All challenge.yml files follow the required format.
```
