package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	build := exec.Command("go", "build", "-o", "./pgo", ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build pgo binary: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.Remove("./pgo")
	os.Exit(code)
}

func TestCLI_GetTags(t *testing.T) {
	// Skip this test if we don't have environment variables set
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	cmd := exec.Command("./pgo", "get", "tags")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	output := stdout.String()
	// Check that output is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", output)
	}

	// Check for expected JSON fields
	if _, ok := result["count"]; !ok {
		t.Errorf("Expected 'count' field in JSON output")
	}

	if _, ok := result["results"]; !ok {
		t.Errorf("Expected 'results' field in JSON output")
	}
}

func TestCLI_GetDocs_WithTagNames(t *testing.T) {
	// Skip this test if we don't have environment variables set
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	cmd := exec.Command("./pgo", "get", "docs")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	output := stdout.String()
	// Check that output is valid JSON
	var result DocumentListOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", output)
	}

	// Check that tag_names field exists and is populated correctly
	for _, doc := range result.Results {
		if doc.TagNames == nil {
			t.Errorf("Expected tag_names field to be present")
		}
		// TagNames should match the length of Tags
		if len(doc.Tags) != len(doc.TagNames) {
			t.Errorf("Expected tag_names to match tags length, got %d vs %d", len(doc.TagNames), len(doc.Tags))
		}
	}
}

func TestCLI_SearchDocs(t *testing.T) {
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	cmd := exec.Command("./pgo", "search", "docs", "invoice")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	output := stdout.String()
	// Check that output is valid JSON
	var result DocumentListOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", output)
	}
}

func TestCLI_SearchDocs_TitleOnly(t *testing.T) {
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	cmd := exec.Command("./pgo", "search", "docs", "-title-only", "invoice")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	output := stdout.String()
	// Check that output is valid JSON
	var result DocumentListOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", output)
	}
}

func TestCLI_SearchTags(t *testing.T) {
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	cmd := exec.Command("./pgo", "search", "tags", "invoice")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	output := stdout.String()
	// Check that output is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", output)
	}
}

func TestCLI_InvalidCommand(t *testing.T) {
	cmd := exec.Command("./pgo", "invalid", "command")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL=dummy",
		"PAPERLESS_TOKEN=dummy",
	)

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Errorf("Expected command to fail with invalid arguments")
	}

	errorOutput := stderr.String()
	if !strings.Contains(errorOutput, "unknown command") {
		t.Errorf("Expected 'unknown command' in error output, got: %s", errorOutput)
	}
}

func TestCLI_InvalidResource(t *testing.T) {
	cmd := exec.Command("./pgo", "get", "invalid")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL=dummy",
		"PAPERLESS_TOKEN=dummy",
	)

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Errorf("Expected command to fail with invalid resource")
	}

	errorOutput := stderr.String()
	if !strings.Contains(errorOutput, "unknown resource") {
		t.Errorf("Expected 'unknown resource' in error output, got: %s", errorOutput)
	}
}

func TestCLI_GetSpecificDoc(t *testing.T) {
	// Skip this test if we don't have environment variables set
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	// First, list documents to get a valid ID
	listCmd := exec.Command("./pgo", "get", "docs")
	listCmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)
	var listStdout bytes.Buffer
	listCmd.Stdout = &listStdout
	listCmd.Stderr = os.Stderr // Capture stderr for debugging

	err := listCmd.Run()
	if err != nil {
		t.Fatalf("List docs failed: %v", err)
	}

	// Parse JSON to get the first document ID
	var listResult DocumentListOutput
	if err := json.Unmarshal(listStdout.Bytes(), &listResult); err != nil {
		t.Fatalf("Failed to parse list output: %v", err)
	}

	if len(listResult.Results) == 0 {
		t.Skip("No documents found, skipping GetSpecificDoc test")
	}

	docID := fmt.Sprintf("%d", listResult.Results[0].ID)

	cmd := exec.Command("./pgo", "get", "docs", docID)
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	// Parse the JSON output
	var doc DocumentWithTagNames
	if err := json.Unmarshal(stdout.Bytes(), &doc); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", stdout.String())
	}

	// Check that ID matches
	if fmt.Sprintf("%d", doc.ID) != docID {
		t.Errorf("Expected document ID %s, got %d", docID, doc.ID)
	}

	// Check that title is present
	if doc.Title == "" {
		t.Errorf("Expected document to have a title")
	}
}

func TestCLI_GetSpecificTag(t *testing.T) {
	// Skip this test if we don't have environment variables set
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	// First, list tags to get a valid ID
	listCmd := exec.Command("./pgo", "get", "tags")
	listCmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)
	var listStdout bytes.Buffer
	listCmd.Stdout = &listStdout
	listCmd.Stderr = os.Stderr

	err := listCmd.Run()
	if err != nil {
		t.Fatalf("List tags failed: %v", err)
	}

	// Parse JSON to get the first tag ID using a proper struct
	type TagListResult struct {
		Results []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"results"`
	}
	var listResult TagListResult
	if err := json.Unmarshal(listStdout.Bytes(), &listResult); err != nil {
		t.Fatalf("Failed to parse list output: %v", err)
	}

	if len(listResult.Results) == 0 {
		t.Skip("No tags found, skipping GetSpecificTag test")
	}

	tagID := fmt.Sprintf("%d", listResult.Results[0].ID)

	cmd := exec.Command("./pgo", "get", "tags", tagID)
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	// Parse the JSON output using a proper struct
	type TagResult struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var tag TagResult
	if err := json.Unmarshal(stdout.Bytes(), &tag); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", stdout.String())
	}

	// Check that ID matches
	if fmt.Sprintf("%d", tag.ID) != tagID {
		t.Errorf("Expected tag ID %s, got %d", tagID, tag.ID)
	}

	// Check that name is present
	if tag.Name == "" {
		t.Errorf("Expected tag to have a name")
	}
}

func TestCLI_InvalidID(t *testing.T) {
	cmd := exec.Command("./pgo", "get", "docs", "invalid")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL=dummy",
		"PAPERLESS_TOKEN=dummy",
	)

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Errorf("Expected command to fail with invalid ID")
	}

	errorOutput := stderr.String()
	if !strings.Contains(errorOutput, "invalid ID format") {
		t.Errorf("Expected 'invalid ID format' in error output, got: %s", errorOutput)
	}
}

func TestCLI_TagCache(t *testing.T) {
	cmd := exec.Command("./pgo", "tagcache")
	// No env vars needed - tagcache doesn't require auth

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	// Should print a path ending with tags.json
	if !strings.HasSuffix(strings.TrimSpace(output), "tags.json") {
		t.Errorf("Expected output to end with 'tags.json', got: %s", output)
	}

	// Should contain paperless-go in the path
	if !strings.Contains(output, "paperless-go") {
		t.Errorf("Expected output to contain 'paperless-go', got: %s", output)
	}
}

func TestCLI_TagCache_WithCustomXDG(t *testing.T) {
	cmd := exec.Command("./pgo", "tagcache")
	cmd.Env = append(os.Environ(), "XDG_CACHE_HOME=/tmp/test-cache")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v, stderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	expected := "/tmp/test-cache/paperless-go/tags.json"
	if output != expected {
		t.Errorf("Expected output to be %s, got: %s", expected, output)
	}
}

func TestCLI_OutputFormat_InvalidFormat(t *testing.T) {
	cmd := exec.Command("./pgo", "-output-format=xml", "get", "tags")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL=dummy",
		"PAPERLESS_TOKEN=dummy",
	)

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Errorf("Expected command to fail with invalid output format")
	}

	errorOutput := stderr.String()
	if !strings.Contains(errorOutput, "unsupported output format") {
		t.Errorf("Expected 'unsupported output format' in error output, got: %s", errorOutput)
	}
}

func TestCLI_OutputFormat_JSON(t *testing.T) {
	// Skip this test if we don't have environment variables set
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	cmd := exec.Command("./pgo", "-output-format=json", "get", "tags")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	output := stdout.String()
	// Check that output is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", output)
	}
}

func TestCLI_ApplyDocs_MissingFlags(t *testing.T) {
	cmd := exec.Command("./pgo", "apply", "docs", "1")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL=dummy",
		"PAPERLESS_TOKEN=dummy",
	)

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Errorf("Expected command to fail with missing flags")
	}

	errorOutput := stderr.String()
	if !strings.Contains(errorOutput, "missing required flag: --tags") {
		t.Errorf("Expected 'missing required flag: --tags' in error output, got: %s", errorOutput)
	}
}

func TestCLI_ApplyDocs_InvalidID(t *testing.T) {
	cmd := exec.Command("./pgo", "apply", "docs", "invalid", "--tags=1")
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL=dummy",
		"PAPERLESS_TOKEN=dummy",
	)

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Errorf("Expected command to fail with invalid ID")
	}

	errorOutput := stderr.String()
	if !strings.Contains(errorOutput, "invalid ID format") {
		t.Errorf("Expected 'invalid ID format' in error output, got: %s", errorOutput)
	}
}

func TestCLI_ApplyDocs_Integration(t *testing.T) {
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	// Get Doc ID
	listCmd := exec.Command("./pgo", "get", "docs")
	listCmd.Env = append(os.Environ(), "PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"), "PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"))
	var listStdout bytes.Buffer
	listCmd.Stdout = &listStdout
	listCmd.Stderr = os.Stderr
	if err := listCmd.Run(); err != nil {
		t.Fatalf("List docs failed: %v", err)
	}

	var docsResult DocumentListOutput
	if err := json.Unmarshal(listStdout.Bytes(), &docsResult); err != nil {
		t.Fatalf("Failed to parse list docs output: %v", err)
	}
	if len(docsResult.Results) == 0 {
		t.Skip("No docs found")
	}
	docID := docsResult.Results[0].ID

	// Get Tag ID
	tagCmd := exec.Command("./pgo", "get", "tags")
	tagCmd.Env = append(os.Environ(), "PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"), "PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"))
	var tagStdout bytes.Buffer
	tagCmd.Stdout = &tagStdout
	tagCmd.Stderr = os.Stderr
	if err := tagCmd.Run(); err != nil {
		t.Fatalf("List tags failed: %v", err)
	}

	type TagListResult struct {
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}
	var tagResult TagListResult
	if err := json.Unmarshal(tagStdout.Bytes(), &tagResult); err != nil {
		t.Fatalf("Failed to parse list tags output: %v", err)
	}
	if len(tagResult.Results) == 0 {
		t.Skip("No tags found")
	}
	tagID := tagResult.Results[0].ID

	// Apply
	cmd := exec.Command("./pgo", "apply", "docs", fmt.Sprintf("%d", docID), "--tags="+fmt.Sprintf("%d", tagID))
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Apply command failed: %v, stderr: %s", err, stderr.String())
	}

	var updated DocumentWithTagNames
	if err := json.Unmarshal(stdout.Bytes(), &updated); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", stdout.String())
	}
	if updated.ID != docID {
		t.Errorf("Expected document ID %d, got %d", docID, updated.ID)
	}
}

func TestCLI_AddTag_Integration(t *testing.T) {
	if os.Getenv("PAPERLESS_URL") == "" || os.Getenv("PAPERLESS_TOKEN") == "" {
		t.Skip("Skipping integration test - PAPERLESS_URL and PAPERLESS_TOKEN not set")
	}

	// Create unique tag name
	tagName := "test-tag-" + time.Now().Format("20060102150405")

	cmd := exec.Command("./pgo", "add", "tag", tagName)
	cmd.Env = append(os.Environ(),
		"PAPERLESS_URL="+os.Getenv("PAPERLESS_URL"),
		"PAPERLESS_TOKEN="+os.Getenv("PAPERLESS_TOKEN"),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Add tag command failed: %v, stderr: %s", err, stderr.String())
	}

	var result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s", stdout.String())
	}
	if result.Name != tagName {
		t.Errorf("Expected tag name %s, got: %s", tagName, result.Name)
	}
	if result.ID == 0 {
		t.Errorf("Expected non-zero tag ID")
	}
}
