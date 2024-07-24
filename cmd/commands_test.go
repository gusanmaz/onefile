package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gusanmaz/onefile/utils"
)

const testProjectPath = "test_toy_project"

func setupTestProject(t *testing.T) {
	err := os.MkdirAll(testProjectPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}

	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
    fmt.Println("Hello from toy project!")
}`,
		"README.md": `# Toy Project

This is a simple toy project used for testing the onefile tool.`,
		".gitignore": `# Ignore compiled binaries
*.exe
*.out`,
	}

	for name, content := range files {
		err := ioutil.WriteFile(filepath.Join(testProjectPath, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}
}

func teardownTestProject(t *testing.T) {
	err := os.RemoveAll(testProjectPath)
	if err != nil {
		t.Fatalf("Failed to remove test project directory: %v", err)
	}
}

func TestDumpCommand(t *testing.T) {
	setupTestProject(t)
	defer teardownTestProject(t)

	cmd := NewDumpCmd()
	cmd.SetArgs([]string{"-p", testProjectPath, "-o", "test_output", "-t", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Dump command failed: %v", err)
	}

	// Check if the output file exists
	_, err = os.Stat("test_output.json")
	if os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Read and parse the JSON output
	data, err := ioutil.ReadFile("test_output.json")
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var projectData utils.ProjectData
	err = json.Unmarshal(data, &projectData)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check if the expected files are present
	expectedFiles := []string{"main.go", "README.md", ".gitignore"}
	for _, file := range expectedFiles {
		found := false
		for _, f := range projectData.Files {
			if f.Path == file {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %s not found in output", file)
		}
	}

	// Clean up
	os.Remove("test_output.json")
}

func TestReconstructCommand(t *testing.T) {
	setupTestProject(t)
	defer teardownTestProject(t)

	// First, use the dump command to create a JSON file
	dumpCmd := NewDumpCmd()
	dumpCmd.SetArgs([]string{"-p", testProjectPath, "-o", "test_dump", "-t", "json"})
	err := dumpCmd.Execute()
	if err != nil {
		t.Fatalf("Dump command failed: %v", err)
	}

	// Now use the reconstruct command
	reconstructCmd := NewReconstructCmd()
	reconstructCmd.SetArgs([]string{"-j", "test_dump.json", "-o", "test_reconstruct"})
	err = reconstructCmd.Execute()
	if err != nil {
		t.Fatalf("Reconstruct command failed: %v", err)
	}

	// Check if the reconstructed files exist
	expectedFiles := []string{"main.go", "README.md", ".gitignore"}
	for _, file := range expectedFiles {
		path := filepath.Join("test_reconstruct", file)
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("Expected file %s not found in reconstructed output", file)
			// Print directory contents for debugging
			files, _ := ioutil.ReadDir("test_reconstruct")
			t.Logf("Contents of test_reconstruct directory:")
			for _, f := range files {
				t.Logf("- %s", f.Name())
			}
		} else if err != nil {
			t.Errorf("Error checking file %s: %v", file, err)
		} else {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				t.Errorf("Error reading file %s: %v", file, err)
			} else {
				t.Logf("Content of %s:\n%s", file, string(content))
			}
		}
	}

	// Clean up
	os.RemoveAll("test_reconstruct")
	os.Remove("test_dump.json")
}

// Add more tests for other commands as needed
