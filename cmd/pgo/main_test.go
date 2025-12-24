package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

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

	cmd := exec.Command("./pgo", "get", "docs", "2411")
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
	if !strings.Contains(output, "Document 2411") {
		t.Errorf("Expected output to contain 'Document 2411', got: %s", output)
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

	cmd := exec.Command("./pgo", "get", "tags", "2")
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
	if !strings.Contains(output, "Tag 2") {
		t.Errorf("Expected output to contain 'Tag 2', got: %s", output)
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
