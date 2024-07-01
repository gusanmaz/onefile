package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Function to generate markdown from JSON
func generateMarkdownFromJSON(jsonPath, outputPath string) {
	data, err := ioutil.ReadFile(jsonPath)
	check(err, "reading JSON file")

	var projectData ProjectData
	err = json.Unmarshal(data, &projectData)
	check(err, "unmarshaling JSON")

	var mdContent strings.Builder

	mdContent.WriteString("# Project Structure\n\n")
	mdContent.WriteString("## Directory Tree\n\n")
	mdContent.WriteString("```\n")

	for _, dir := range projectData.Directories {
		mdContent.WriteString(dir + "/\n")
	}

	for _, file := range projectData.Files {
		mdContent.WriteString(file.Path + "\n")
	}

	mdContent.WriteString("```\n\n")
	mdContent.WriteString("## File Contents\n\n")

	for _, file := range projectData.Files {
		ext := filepath.Ext(file.Path)
		lang := getLanguageFromExtension(ext)
		mdContent.WriteString(fmt.Sprintf("### File: %s\n\n```%s\n%s\n```\n\n", file.Path, lang, file.Content))
	}

	err = ioutil.WriteFile(outputPath, []byte(mdContent.String()), 0644)
	check(err, "writing markdown to file")
}

func main() {
	jsonPath := flag.String("json", "project_data.json", "The input JSON file")
	outputPath := flag.String("output", "project_structure.md", "The output markdown file")

	flag.Parse()

	generateMarkdownFromJSON(*jsonPath, *outputPath)

	fmt.Printf("Markdown generated successfully: %s\n", *outputPath)
}
