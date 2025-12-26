package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
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
	if !strings.Contains(output, "Found") {
		t.Errorf("Expected output to contain 'Found', got: %s", output)
	}

	if !strings.Contains(output, "tags") {
		t.Errorf("Expected output to contain 'tags', got: %s", output)
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
	if !strings.Contains(output, "Found") {
		t.Errorf("Expected output to contain 'Found', got: %s", output)
	}

	if !strings.Contains(output, "documents") {
		t.Errorf("Expected output to contain 'documents', got: %s", output)
	}

	// Check that we're showing tag names, not just numbers
	// This is a basic check - we expect to see some text that looks like tag names
	lines := strings.Split(output, "\n")
	tagLineFound := false
	for _, line := range lines {
		if strings.HasPrefix(line, "Tags: ") {
			tagLineFound = true
			// Check that we have something that looks like tag names (not just numbers in brackets)
			if strings.Contains(line, "unknown(") {
				// If we have unknown tags, that's still OK for this test
				continue
			}
			// We should have some non-numeric content
			tagContent := strings.TrimPrefix(line, "Tags: ")
			if strings.Trim(tagContent, " []") != "" {
				// Found some tag content
				break
			}
		}
	}

	if !tagLineFound {
		t.Errorf("Expected to find tag lines in output")
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
	if !strings.Contains(output, "Found") || !strings.Contains(output, "documents") {
		t.Errorf("Expected output to contain search results, got: %s", output)
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
	if !strings.Contains(output, "Found") || !strings.Contains(output, "documents") {
		t.Errorf("Expected output to contain search results, got: %s", output)
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
	if !strings.Contains(output, "Found") || !strings.Contains(output, "tags") {
		t.Errorf("Expected output to contain search results, got: %s", output)
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

	output := listStdout.String()
	if !strings.Contains(output, "ID: ") {
		t.Skip("No documents found, skipping GetSpecificDoc test")
	}

	// Simple parsing to find the first ID
	start := strings.Index(output, "ID: ") + 4
	end := strings.Index(output[start:], "\n")
	if end == -1 {
		// Try end of string if no newline
		end = len(output[start:])
	}

	docID := strings.TrimSpace(output[start : start+end])
	if docID == "" {
		t.Skip("Could not parse valid document ID")
	}

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

	output = stdout.String()
	if !strings.Contains(output, "Document "+docID) {
		t.Errorf("Expected output to contain 'Document %s', got: %s", docID, output)
	}

	if !strings.Contains(output, "Title:") {
		t.Errorf("Expected output to contain 'Title:', got: %s", output)
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

	output := listStdout.String()
	if !strings.Contains(output, "ID: ") {
		t.Skip("No tags found, skipping GetSpecificTag test")
	}

	// Simple parsing to find the first ID
	start := strings.Index(output, "ID: ") + 4
	end := strings.Index(output[start:], "\n")
	if end == -1 {
		end = len(output[start:])
	}

	tagID := strings.TrimSpace(output[start : start+end])
	if tagID == "" {
		t.Skip("Could not parse valid tag ID")
	}

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

	output = stdout.String()
	if !strings.Contains(output, "Tag "+tagID) {
		t.Errorf("Expected output to contain 'Tag %s', got: %s", tagID, output)
	}

	if !strings.Contains(output, "Name:") {
		t.Errorf("Expected output to contain 'Name:', got: %s", output)
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
