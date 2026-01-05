package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintChallengeFile(t *testing.T) {
	// Create a temporary directory for tests
	tempDir := t.TempDir()

	// Create lintrc.yaml in temp directory
	lintrcContent := `tags:
  condition: and
  patterns:
    - type: static
      values:
        - easy
        - medium
        - hard
requirements:
  condition: and
  patterns:
    - type: static
      values:
        - "welcome"`
	lintrcPath := filepath.Join(tempDir, "lintrc.yaml")
	err := os.WriteFile(lintrcPath, []byte(lintrcContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create lintrc.yaml: %v", err)
	}

	// Change to temp directory so lintrc.yaml is found
	origDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(origDir)
	}()
	_ = os.Chdir(tempDir)

	tests := []struct {
		name         string
		yamlContent  string
		files        []string
		wantErrors   []string
		wantWarnings []string
	}{
		{
			name: "valid challenge",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{},
		},
		{
			name: "challenge without welcome requires welcome in requirements",
			yamlContent: `
name: "test_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements:
  - welcome
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
`,
			files:      []string{},
			wantErrors: []string{},
		},
		{
			name: "missing welcome requirement",
			yamlContent: `
name: "test_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{"Requirements validation failed for pattern type 'static'"},
		},
		{
			name: "non-null image",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements: []
value: 500
type: dynamic
extra:
  initial: 500
  decay: 100
  minimum: 100
image: "some-image"
host: null
state: visible
version: "0.1"
`,
			files:      []string{},
			wantErrors: []string{"Field 'image' should be null"},
		},
		{
			name: "wrong state",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements: []
value: 500
type: dynamic
extra:
  initial: 500
  decay: 100
  minimum: 100
image: null
host: null
state: hidden
version: "0.1"
`,
			files:      []string{},
			wantErrors: []string{"Field 'state' should be 'visible'"},
		},
		{
			name: "wrong version",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements: []
value: 500
type: dynamic
extra:
  initial: 500
  decay: 100
  minimum: 100
image: null
host: null
state: visible
version: "1.0"
`,
			files:      []string{},
			wantErrors: []string{"Field 'version' should be '0.1'"},
		},
		{
			name: "invalid tags - no valid tag",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - invalid
files: []
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{"Tags should contain exactly one of: easy, medium, hard"},
		},
		{
			name: "invalid tags - multiple valid tags",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
  - medium
files: []
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{"Tags should contain exactly one of: easy, medium, hard"},
		},
		{
			name: "missing file",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files:
  - missing.txt
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{"File specified in 'files' does not exist: missing.txt"},
		},
		{
			name: "file too large",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files:
  - large_file.txt
requirements: []
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
`,
			files:      []string{"large_file.txt:large"},
			wantErrors: []string{"File 'large_file.txt' is too large"},
		},
		{
			name: "requirements condition none - no validation",
			yamlContent: `
name: "test_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{},
		},
		{
			name: "tags condition none - no validation",
			yamlContent: `
name: "test_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - invalid_tag
files: []
requirements:
  - welcome
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
`,
			files:        []string{},
			wantErrors:   []string{},
			wantWarnings: []string{},
		},
		{
			name: "type standard - should warn",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements: []
value: 500
type: standard
extra:
  initial: 500
  decay: 100
  minimum: 100
image: null
host: null
state: visible
version: "0.1"
`,
			files:        []string{},
			wantErrors:   []string{},
			wantWarnings: []string{"Field 'type' is 'standard', did you intend to use 'dynamic'?"},
		},
		{
			name: "flags as map with inline style",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - {
      type: "static",
      content: "flag{wat}",
      data: "case_insensitive",
    }
  - {
      type: "regex",
      content: "(.*)STUFF(.*)",
      data: "case_insensitive",
    }
  - {
      type: "static",
      content: "flag{wat}",
    }
tags:
  - easy
files: []
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{},
		},
		{
			name: "flags as map with multiline style",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - type: static
    content: "flag{wat}"
    data: "case_insensitive"
  - type: regex
    content: "(.*)STUFF(.*)"
    data: "case_insensitive"
  - type: static
    content: "flag{wat}"
tags:
  - easy
files: []
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{},
		},
		{
			name: "invalid flag format - number",
			yamlContent: `
name: "welcome_challenge"
author: "test"
category: "intro"
description: "test description"
flags:
  - 123
tags:
  - easy
files: []
requirements: []
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
`,
			files:      []string{},
			wantErrors: []string{"Invalid YAML format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Create lintrc.yaml for this specific test
			var lintrcContent string
			switch tt.name {
			case "requirements condition none - no validation":
				lintrcContent = `tags:
  condition: and
  patterns:
    - type: static
      values:
        - easy
        - medium
        - hard
requirements:
  condition: none
  patterns:
    - type: static
      values:
        - "welcome"`
			case "tags condition none - no validation":
				lintrcContent = `tags:
  condition: none
  patterns:
    - type: static
      values:
        - easy
        - medium
        - hard
requirements:
  condition: and
  patterns:
    - type: static
      values:
        - "welcome"`
			default:
				lintrcContent = `tags:
  condition: and
  patterns:
    - type: static
      values:
        - easy
        - medium
        - hard
requirements:
  condition: and
  patterns:
    - type: static
      values:
        - "welcome"`
			}
			lintrcPath := filepath.Join(testDir, "lintrc.yaml")
			err = os.WriteFile(lintrcPath, []byte(lintrcContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create lintrc.yaml: %v", err)
			}

			// Create challenge.yml file
			yamlPath := filepath.Join(testDir, "challenge.yml")
			err = os.WriteFile(yamlPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create challenge.yml: %v", err)
			}

			// Create required files
			for _, fileName := range tt.files {
				if strings.Contains(fileName, ":large") {
					// Create a large file (over 1MB)
					actualFileName := strings.Split(fileName, ":")[0]
					filePath := filepath.Join(testDir, actualFileName)
					file, err := os.Create(filePath)
					if err != nil {
						t.Fatalf("Failed to create large file %s: %v", actualFileName, err)
					}
					// Write 1.5MB of data
					data := make([]byte, 1024*1024+1024*512) // 1.5MB
					for i := range data {
						data[i] = 'A'
					}
					_, err = file.Write(data)
					if err != nil {
						_ = file.Close()
						t.Fatalf("Failed to write large file %s: %v", actualFileName, err)
					}
					_ = file.Close()
				} else {
					filePath := filepath.Join(testDir, fileName)
					err = os.WriteFile(filePath, []byte("test content"), 0644)
					if err != nil {
						t.Fatalf("Failed to create test file %s: %v", fileName, err)
					}
				}
			}

			// Change to test directory so lintrc.yaml is found
			origDir, _ := os.Getwd()
			defer func() {
				_ = os.Chdir(origDir)
			}()
			_ = os.Chdir(testDir)

			// Run linting
			result := lintChallengeFile(yamlPath)

			// Check errors
			if len(tt.wantErrors) == 0 {
				if len(result.Errors) != 0 {
					t.Errorf("Expected no errors, but got: %v", result.Errors)
				}
			} else {
				if len(result.Errors) == 0 {
					t.Errorf("Expected errors %v, but got none", tt.wantErrors)
				} else {
					for _, wantError := range tt.wantErrors {
						found := false
						for _, gotError := range result.Errors {
							if strings.Contains(gotError, wantError) {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Expected error containing '%s' not found in: %v", wantError, result.Errors)
						}
					}
				}
			}

			// Check warnings
			if len(tt.wantWarnings) == 0 {
				if len(result.Warnings) != 0 {
					t.Errorf("Expected no warnings, but got: %v", result.Warnings)
				}
			} else {
				if len(result.Warnings) == 0 {
					t.Errorf("Expected warnings %v, but got none", tt.wantWarnings)
				} else {
					for _, wantWarning := range tt.wantWarnings {
						found := false
						for _, gotWarning := range result.Warnings {
							if strings.Contains(gotWarning, wantWarning) {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Expected warning containing '%s' not found in: %v", wantWarning, result.Warnings)
						}
					}
				}
			}
		})
	}
}

func TestLintChallenges(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create lintrc.yaml in temp directory
	lintrcContent := `tags:
  condition: and
  patterns:
    - type: static
      values:
        - easy
        - medium
        - hard
requirements:
  condition: and
  patterns:
    - type: static
      values:
        - "welcome"`
	lintrcPath := filepath.Join(tempDir, "lintrc.yaml")
	err := os.WriteFile(lintrcPath, []byte(lintrcContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create lintrc.yaml: %v", err)
	}

	// Change to temp directory so lintrc.yaml is found
	origDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(origDir)
	}()
	_ = os.Chdir(tempDir)

	// Create some test directories and files
	dirs := []string{
		"genre1/chall1",
		"genre1/chall2",
		"genre2/chall1",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		yamlContent := `
name: "welcome_test"
author: "test"
category: "intro"
description: "test description"
flags:
  - "flag{test}"
tags:
  - easy
files: []
requirements: []
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
`
		yamlPath := filepath.Join(dirPath, "challenge.yml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create challenge.yml in %s: %v", dirPath, err)
		}
	}

	// Run linting
	results, err := lintChallenges(tempDir)
	if err != nil {
		t.Fatalf("lintChallenges failed: %v", err)
	}

	// Should find 3 challenge.yml files
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// All should pass
	for _, result := range results {
		if len(result.Errors) > 0 {
			t.Errorf("Unexpected errors in %s: %v", result.File, result.Errors)
		}
	}
}

func TestInvalidYAML(t *testing.T) {
	tempDir := t.TempDir()

	invalidYAML := `
name: "test"
invalid yaml content:
  - not: properly
    - formatted
`

	yamlPath := filepath.Join(tempDir, "challenge.yml")
	err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid yaml file: %v", err)
	}

	result := lintChallengeFile(yamlPath)

	if len(result.Errors) == 0 {
		t.Error("Expected error for invalid YAML, but got none")
	}

	found := false
	for _, err := range result.Errors {
		if strings.Contains(err, "Invalid YAML format") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'Invalid YAML format' error, got: %v", result.Errors)
	}
}
