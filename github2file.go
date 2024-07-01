package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const githubAPIURL = "https://api.github.com/repos/%s/%s/contents"

type GithubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

func fetchGithubRepo(owner, repo string) (ProjectData, error) {
	var projectData ProjectData
	url := fmt.Sprintf(githubAPIURL, owner, repo)

	err := fetchContents(url, "", &projectData)
	if err != nil {
		return ProjectData{}, err
	}

	return projectData, nil
}

func fetchContents(url, path string, projectData *ProjectData) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var contents []GithubContent
	err = json.Unmarshal(body, &contents)
	if err != nil {
		return err
	}

	for _, content := range contents {
		if content.Type == "dir" {
			projectData.Directories = append(projectData.Directories, content.Path)
			err = fetchContents(fmt.Sprintf("%s/%s", url, content.Name), content.Path, projectData)
			if err != nil {
				return err
			}
		} else if content.Type == "file" && isTextFile(filepath.Ext(content.Name)) {
			fileContent, err := fetchFileContent(content.DownloadURL)
			if err != nil {
				return err
			}
			projectData.Files = append(projectData.Files, FileData{Path: content.Path, Content: fileContent})
			fmt.Printf("Downloaded: %s\n", content.Path)
		}
	}

	return nil
}

func fetchFileContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseGitHubURL(url string) (string, string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 5 || parts[2] != "github.com" {
		return "", "", fmt.Errorf("invalid GitHub URL")
	}
	return parts[3], parts[4], nil
}

func main() {
	repoURL := flag.String("url", "", "GitHub repository URL")
	owner := flag.String("owner", "", "GitHub repository owner")
	repo := flag.String("repo", "", "GitHub repository name")
	outputType := flag.String("type", "json", "Output type: json or md")
	outputFile := flag.String("output", "", "Output file name")

	flag.Parse()

	if *repoURL != "" {
		var err error
		*owner, *repo, err = parseGitHubURL(*repoURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing GitHub URL: %v\n", err)
			os.Exit(1)
		}
	} else if *owner == "" || *repo == "" {
		fmt.Println("Please provide either a GitHub URL or both owner and repo flags")
		os.Exit(1)
	}

	if *outputFile == "" {
		*outputFile = fmt.Sprintf("%s_%s.%s", *owner, *repo, *outputType)
	}

	fmt.Printf("Fetching repository: %s/%s\n", *owner, *repo)
	startTime := time.Now()

	projectData, err := fetchGithubRepo(*owner, *repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching GitHub repo: %v\n", err)
		os.Exit(1)
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("Repository fetched in %v\n", elapsedTime)

	if *outputType == "json" {
		data, err := json.MarshalIndent(projectData, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		err = ioutil.WriteFile(*outputFile, data, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON file: %v\n", err)
			os.Exit(1)
		}
	} else if *outputType == "md" {
		var mdContent strings.Builder
		mdContent.WriteString("# Project Structure\n\n")
		mdContent.WriteString("## Directory Tree\n\n```\n")
		for _, dir := range projectData.Directories {
			mdContent.WriteString(dir + "/\n")
		}
		for _, file := range projectData.Files {
			mdContent.WriteString(file.Path + "\n")
		}
		mdContent.WriteString("```\n\n## File Contents\n\n")
		for _, file := range projectData.Files {
			ext := filepath.Ext(file.Path)
			lang := getLanguageFromExtension(ext)
			mdContent.WriteString(fmt.Sprintf("### File: %s\n\n```%s\n%s\n```\n\n", file.Path, lang, file.Content))
		}
		err = ioutil.WriteFile(*outputFile, []byte(mdContent.String()), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Markdown file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Invalid output type. Use 'json' or 'md'\n")
		os.Exit(1)
	}

	fmt.Printf("Output file created successfully: %s\n", *outputFile)
}
