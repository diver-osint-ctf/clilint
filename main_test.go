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

	tests := []struct {
		name        string
		yamlContent string
		files       []string
		wantErrors  []string
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
  - introduction
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
  - introduction
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
  - introduction
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
			wantErrors: []string{"Challenges without 'welcome' in name must have 'welcome' in requirements"},
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
  - introduction
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
  - introduction
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
  - introduction
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
			wantErrors: []string{"Tags should contain exactly one of: introduction, easy, medium, hard"},
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
  - introduction
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
			wantErrors: []string{"Tags should contain exactly one of: introduction, easy, medium, hard"},
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
  - introduction
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Create challenges.yaml file
			yamlPath := filepath.Join(testDir, "challenges.yaml")
			err = os.WriteFile(yamlPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create challenges.yaml: %v", err)
			}

			// Create required files
			for _, fileName := range tt.files {
				filePath := filepath.Join(testDir, fileName)
				err = os.WriteFile(filePath, []byte("test content"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", fileName, err)
				}
			}

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
		})
	}
}

func TestLintChallenges(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

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
  - introduction
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
		yamlPath := filepath.Join(dirPath, "challenges.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create challenges.yaml in %s: %v", dirPath, err)
		}
	}

	// Run linting
	results, err := lintChallenges(tempDir)
	if err != nil {
		t.Fatalf("lintChallenges failed: %v", err)
	}

	// Should find 3 challenges.yaml files
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

	yamlPath := filepath.Join(tempDir, "challenges.yaml")
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
