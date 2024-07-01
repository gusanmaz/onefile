package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Struct to store file data
type FileData struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// Struct to store directory and file structure
type ProjectData struct {
	Directories []string   `json:"directories"`
	Files       []FileData `json:"files"`
}

// Function to check for errors and log them
func check(err error, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", message, err)
		os.Exit(1)
	}
}

// Function to create directories if they do not exist
func createDirsIfNotExist(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// Function to write content to a file
func writeFileContent(path string, content string) error {
	dir := filepath.Dir(path)
	err := createDirsIfNotExist(dir)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// Function to read file content
func readFileContent(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Function to check if a file is a text file based on its extension
func isTextFile(ext string) bool {
	textFileExtensions := []string{".txt", ".md", ".html", ".css", ".js", ".py", ".go", ".mod", ".java", ".rb", ".rs", ".cpp", ".c", ".sh"}
	for _, textExt := range textFileExtensions {
		if ext == textExt {
			return true
		}
	}
	return false
}

// Function to get the programming language from the file extension
func getLanguageFromExtension(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".c", ".h":
		return "c"
	case ".mod":
		return "mod"
	case ".md":
		return "md"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".rs":
		return "rust"
	case ".cpp":
		return "cpp"
	case ".sh":
		return "bash"
	case ".txt":
		return ""
	default:
		return ""
	}
}
