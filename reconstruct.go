package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Function to reconstruct project from JSON
func reconstructProjectFromJSON(jsonPath, rootPath string) {
	data, err := ioutil.ReadFile(jsonPath)
	check(err, "reading JSON file")

	var projectData ProjectData
	err = json.Unmarshal(data, &projectData)
	check(err, "unmarshaling JSON")

	for _, dir := range projectData.Directories {
		err := createDirsIfNotExist(filepath.Join(rootPath, dir))
		check(err, "creating directories")
	}

	for _, file := range projectData.Files {
		err := writeFileContent(filepath.Join(rootPath, file.Path), file.Content)
		check(err, "writing file content")
	}

	fmt.Printf("Project reconstructed successfully in: %s\n", rootPath)
}

func main() {
	jsonPath := flag.String("json", "project_data.json", "The input JSON file")
	rootPath := flag.String("path", ".", "The root path to reconstruct the project")

	flag.Parse()

	absRootPath, err := filepath.Abs(strings.TrimSpace(*rootPath))
	check(err, "getting absolute root path")

	reconstructProjectFromJSON(*jsonPath, absRootPath)
}
