package utils

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

func SaveAsJSON(projectData ProjectData, outputPath string, includeGit, includeNonText bool) error {
	// Filter directories
	filteredDirs := make([]string, 0, len(projectData.Directories))
	for _, dir := range projectData.Directories {
		if includeGit || !strings.HasPrefix(dir, ".git") {
			filteredDirs = append(filteredDirs, dir)
		}
	}

	// Filter files
	filteredFiles := make([]FileData, 0, len(projectData.Files))
	for _, file := range projectData.Files {
		if (includeGit || !strings.HasPrefix(file.Path, ".git/")) && (includeNonText || isTextFile(file.Path)) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	// Create filtered project data
	filteredProjectData := ProjectData{
		Directories: filteredDirs,
		Files:       filteredFiles,
	}

	// Marshal the filtered data to JSON
	data, err := json.MarshalIndent(filteredProjectData, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON data to the output file
	return ioutil.WriteFile(outputPath, data, 0644)
}
