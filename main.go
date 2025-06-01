package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Challenge represents the structure of challenges.yaml
type Challenge struct {
	Name         string                 `yaml:"name"`
	Author       string                 `yaml:"author"`
	Category     string                 `yaml:"category"`
	Description  string                 `yaml:"description"`
	Flags        []string               `yaml:"flags"`
	Tags         []string               `yaml:"tags"`
	Files        []string               `yaml:"files"`
	Requirements []string               `yaml:"requirements"`
	Value        int                    `yaml:"value"`
	Type         string                 `yaml:"type"`
	Extra        map[string]interface{} `yaml:"extra"`
	Image        interface{}            `yaml:"image"`
	Host         interface{}            `yaml:"host"`
	State        string                 `yaml:"state"`
	Version      string                 `yaml:"version"`
	Hints        []interface{}          `yaml:"hints"`
}

type LintResult struct {
	File        string
	Errors      []string
	Name        string
	Description string
}

type Env struct {
	token     string
	owner     string
	repo      string
	prNumber  int
	commentPR bool
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Println("Usage: clilint [options] [directory...]")
		fmt.Println("Lints challenges.yaml files in the specified directories (default: current directory)")
		fmt.Println("Options:")
		fmt.Println("  --json           Output results in JSON format for GitHub Actions")
		fmt.Println("  --comment-pr     Post results as PR comment (requires GitHub environment)")
		return
	}

	jsonOutput := false
	commentPR := false
	var targetDirs []string

	// Parse arguments
	for _, arg := range os.Args[1:] {
		if arg == "--json" {
			jsonOutput = true
		} else if arg == "--comment-pr" {
			commentPR = true
		} else if !strings.HasPrefix(arg, "--") {
			targetDirs = append(targetDirs, arg)
		}
	}

	var allResults []LintResult

	// GitHub Actions mode: detect changed directories
	if commentPR {
		env, err := getEnv()
		if err != nil {
			log.Fatalf("Error getting environment: %v", err)
		}

		changedDirs, err := findChangedDirectories(env)
		if err != nil {
			log.Fatalf("Error finding changed directories: %v", err)
		}

		if len(changedDirs) == 0 {
			// No changes, post comment and exit
			err = postNoChangesComment(env)
			if err != nil {
				log.Fatalf("Error posting comment: %v", err)
			}
			return
		}

		// Lint changed directories
		for _, dir := range changedDirs {
			results, err := lintChallenges(dir)
			if err != nil {
				log.Fatalf("Error linting directory %s: %v", dir, err)
			}
			allResults = append(allResults, results...)
		}

		// Post PR comment
		hasErrors := hasLintErrors(allResults)
		err = postPRComment(allResults, hasErrors, env)
		if err != nil {
			log.Fatalf("Error posting PR comment: %v", err)
		}

		if hasErrors {
			os.Exit(1)
		}
		return
	}

	// Local mode: lint specified directories
	if len(targetDirs) == 0 {
		targetDirs = []string{"."}
	}

	for _, dir := range targetDirs {
		results, err := lintChallenges(dir)
		if err != nil {
			log.Fatalf("Error linting directory %s: %v", dir, err)
		}
		allResults = append(allResults, results...)
	}

	hasErrors := hasLintErrors(allResults)

	// Handle JSON output
	if jsonOutput {
		output := map[string]interface{}{
			"success": !hasErrors,
			"results": allResults,
		}

		jsonData, _ := json.Marshal(output)
		fmt.Println(string(jsonData))

		if hasErrors {
			os.Exit(1)
		}
		return
	}

	// Handle standard output
	for _, result := range allResults {
		if len(result.Errors) > 0 {
			fmt.Printf("❌ %s:\n", result.File)
			for _, err := range result.Errors {
				fmt.Printf("  - %s\n", err)
			}
			fmt.Println()
		} else {
			fmt.Printf("✅ %s: OK\n", result.File)
		}
	}

	if hasErrors {
		os.Exit(1)
	} else {
		fmt.Println("All challenges.yaml files passed linting! 🎉")
	}
}

func getEnv() (Env, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return Env{}, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	repository := os.Getenv("INPUT_REPOSITORY")
	if repository == "" {
		repository = os.Getenv("GITHUB_REPOSITORY")
	}
	if repository == "" {
		return Env{}, fmt.Errorf("INPUT_REPOSITORY or GITHUB_REPOSITORY environment variable is required")
	}

	repoPath := strings.Split(repository, "/")
	if len(repoPath) != 2 {
		return Env{}, fmt.Errorf("invalid repository format: %s", repository)
	}
	owner, repo := repoPath[0], repoPath[1]

	prNumberStr := os.Getenv("INPUT_PR_NUMBER")
	if prNumberStr == "" {
		prNumberStr = os.Getenv("PR_NUMBER")
	}
	if prNumberStr == "" {
		return Env{}, fmt.Errorf("INPUT_PR_NUMBER or PR_NUMBER environment variable is required")
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		return Env{}, fmt.Errorf("invalid PR number: %v", err)
	}

	return Env{
		token:     token,
		owner:     owner,
		repo:      repo,
		prNumber:  prNumber,
		commentPR: true,
	}, nil
}

func getGitHubClient(token string) (*github.Client, context.Context) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client, ctx
}

func findChangedDirectories(env Env) ([]string, error) {
	client, ctx := getGitHubClient(env.token)

	var allFiles []string
	opt := &github.ListOptions{PerPage: 100}

	for {
		files, resp, err := client.PullRequests.ListFiles(ctx, env.owner, env.repo, env.prNumber, opt)
		if err != nil {
			return nil, fmt.Errorf("error getting PR files: %v", err)
		}

		for _, file := range files {
			allFiles = append(allFiles, file.GetFilename())
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Find directories containing challenges.yaml files
	dirSet := make(map[string]bool)

	for _, file := range allFiles {
		dir := filepath.Dir(file)

		// Check if the file is challenges.yaml or if the directory contains challenges.yaml
		if filepath.Base(file) == "challenges.yaml" {
			dirSet[dir] = true
		} else {
			// Check parent directories for challenges.yaml
			current := dir
			for current != "." && current != "/" {
				if _, err := os.Stat(filepath.Join(current, "challenges.yaml")); err == nil {
					dirSet[current] = true
					break
				}
				current = filepath.Dir(current)
			}
		}
	}

	var directories []string
	for dir := range dirSet {
		directories = append(directories, dir)
	}

	return directories, nil
}

func hasLintErrors(results []LintResult) bool {
	for _, result := range results {
		if len(result.Errors) > 0 {
			return true
		}
	}
	return false
}

func postNoChangesComment(env Env) error {
	commentBody := "## 📋 CTF Challenges YAML Linting Results\n\n🔍 No challenges.yaml files were affected by this PR.\n\nNo linting required for this change."
	return createComment(env, commentBody)
}

func postPRComment(results []LintResult, hasErrors bool, env Env) error {
	commentBody := generateCommentBody(results, hasErrors)
	return createComment(env, commentBody)
}

func generateCommentBody(results []LintResult, hasErrors bool) string {
	var body strings.Builder

	if hasErrors {
		body.WriteString("## ❌ CTF Challenges YAML Linting Results\n\n")
		body.WriteString("### 🔍 Linting Results for Changes in This PR:\n\n")
	} else {
		body.WriteString("## 🎉 CTF Challenges YAML Linting Results\n\n")
		body.WriteString("✅ All affected challenges.yaml files passed linting!\n\n")
		body.WriteString("### 📋 Checked Challenges in This PR:\n\n")
	}

	for _, result := range results {
		if len(result.Errors) > 0 {
			body.WriteString(fmt.Sprintf("#### ❌ **%s** (`%s`)\n\n", result.Name, result.File))
			if result.Description != "" {
				body.WriteString("**Description:**\n")
				body.WriteString(result.Description)
				body.WriteString("\n\n")
			}
			body.WriteString("**Issues found:**\n")
			for _, err := range result.Errors {
				body.WriteString(fmt.Sprintf("- %s\n", err))
			}
			body.WriteString("\n---\n\n")
		} else {
			body.WriteString(fmt.Sprintf("#### 🚩 **%s** (`%s`)\n\n", result.Name, result.File))
			if result.Description != "" {
				body.WriteString(result.Description)
				body.WriteString("\n\n---\n\n")
			}
		}
	}

	if hasErrors {
		body.WriteString("⚠️ Please fix the issues above and try again.")
	} else {
		body.WriteString("✨ Great job! All challenges.yaml files in the changed directories follow the required format and standards.")
	}

	return body.String()
}

func createComment(env Env, body string) error {
	client, ctx := getGitHubClient(env.token)

	comment := &github.IssueComment{
		Body: github.String(body),
	}

	_, _, err := client.Issues.CreateComment(ctx, env.owner, env.repo, env.prNumber, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment: %v", err)
	}

	fmt.Printf("Successfully posted comment to PR #%d\n", env.prNumber)
	return nil
}

func lintChallenges(rootDir string) ([]LintResult, error) {
	var results []LintResult

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "challenges.yaml" {
			result := lintChallengeFile(path)
			results = append(results, result)
		}

		return nil
	})

	return results, err
}

func lintChallengeFile(filePath string) LintResult {
	result := LintResult{
		File:        filePath,
		Errors:      []string{},
		Name:        "",
		Description: "",
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to read file: %v", err))
		return result
	}

	// Parse YAML
	var challenge Challenge
	err = yaml.Unmarshal(data, &challenge)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid YAML format: %v", err))
		return result
	}

	// Store challenge info for PR display
	result.Name = challenge.Name
	result.Description = challenge.Description

	// Lint checks
	result.Errors = append(result.Errors, checkFiles(filePath, challenge.Files)...)
	result.Errors = append(result.Errors, checkRequirements(challenge.Name, challenge.Requirements)...)
	result.Errors = append(result.Errors, checkImage(challenge.Image)...)
	result.Errors = append(result.Errors, checkState(challenge.State)...)
	result.Errors = append(result.Errors, checkVersion(challenge.Version)...)
	result.Errors = append(result.Errors, checkTags(challenge.Tags)...)

	return result
}

func checkFiles(challengePath string, files []string) []string {
	var errors []string
	baseDir := filepath.Dir(challengePath)

	for _, file := range files {
		fullPath := filepath.Join(baseDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("File specified in 'files' does not exist: %s", file))
		}
	}

	return errors
}

func checkRequirements(name string, requirements []string) []string {
	var errors []string

	// Check if name contains "welcome"
	if !strings.Contains(strings.ToLower(name), "welcome") {
		// Name doesn't contain "welcome", so requirements should include "welcome"
		hasWelcome := false
		for _, req := range requirements {
			if strings.ToLower(req) == "welcome" {
				hasWelcome = true
				break
			}
		}
		if !hasWelcome {
			errors = append(errors, "Challenges without 'welcome' in name must have 'welcome' in requirements")
		}
	}

	return errors
}

func checkImage(image interface{}) []string {
	var errors []string

	if image != nil {
		errors = append(errors, "Field 'image' should be null")
	}

	return errors
}

func checkState(state string) []string {
	var errors []string

	if state != "visible" {
		errors = append(errors, "Field 'state' should be 'visible'")
	}

	return errors
}

func checkVersion(version string) []string {
	var errors []string

	if version != "0.1" {
		errors = append(errors, "Field 'version' should be '0.1'")
	}

	return errors
}

func checkTags(tags []string) []string {
	var errors []string
	validTags := []string{"introduction", "easy", "medium", "hard"}

	foundValidTags := []string{}
	for _, tag := range tags {
		for _, validTag := range validTags {
			if tag == validTag {
				foundValidTags = append(foundValidTags, tag)
				break
			}
		}
	}

	if len(foundValidTags) != 1 {
		errors = append(errors, "Tags should contain exactly one of: introduction, easy, medium, hard")
	}

	return errors
}
