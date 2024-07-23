package internal

import (
	"fmt"
	"github.com/sabhiram/go-gitignore"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

func DumpProject(rootPath string, gitIgnore *ignore.GitIgnore, includeGit, includeNonText bool) (ProjectData, error) {
	var projectData ProjectData

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			projectData.Directories = append(projectData.Directories, relPath)
		} else {
			if MatchesPatterns(relPath, gitIgnore, includeGit, includeNonText) {
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				projectData.Files = append(projectData.Files, FileData{Path: relPath, Content: string(content)})
			} else {
				projectData.Files = append(projectData.Files, FileData{Path: relPath, Content: ""})
			}
		}
		return nil
	})

	if err != nil {
		return ProjectData{}, err
	}

	sort.Strings(projectData.Directories)
	sort.Slice(projectData.Files, func(i, j int) bool {
		return projectData.Files[i].Path < projectData.Files[j].Path
	})

	return projectData, nil
}

func ReconstructProject(projectData ProjectData, outputPath string) error {
	for _, dir := range projectData.Directories {
		err := os.MkdirAll(filepath.Join(outputPath, dir), 0755)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %v", dir, err)
		}
	}

	for _, file := range projectData.Files {
		filePath := filepath.Join(outputPath, file.Path)
		err := ioutil.WriteFile(filePath, []byte(file.Content), 0644)
		if err != nil {
			return fmt.Errorf("error writing file %s: %v", file.Path, err)
		}
	}

	return nil
}