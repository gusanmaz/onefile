package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sabhiram/go-gitignore"
	"github.com/schollz/progressbar/v3"
)

func FetchGithubRepo(owner, repo, path string, gitIgnore *ignore.GitIgnore, useGit bool, githubToken string, includeGit, includeNonText bool) (ProjectData, error) {
	if useGit {
		return fetchWithGit(owner, repo, path, gitIgnore, includeGit, includeNonText)
	}
	return fetchWithAPI(owner, repo, path, gitIgnore, githubToken, includeGit, includeNonText)
}

func fetchWithGit(owner, repo, path string, gitIgnore *ignore.GitIgnore, includeGit, includeNonText bool) (ProjectData, error) {
	tmpDir, err := ioutil.TempDir("", "github-clone-")
	if err != nil {
		return ProjectData{}, err
	}
	defer os.RemoveAll(tmpDir)

	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, tmpDir)
	err = cmd.Run()
	if err != nil {
		return ProjectData{}, fmt.Errorf("git clone failed: %v", err)
	}

	projectPath := filepath.Join(tmpDir, path)
	return DumpProject(projectPath, gitIgnore, includeGit, includeNonText)
}

func fetchWithAPI(owner, repo, path string, gitIgnore *ignore.GitIgnore, githubToken string, includeGit, includeNonText bool) (ProjectData, error) {
	var projectData ProjectData
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)

	bar := progressbar.Default(-1, "Fetching repository")

	client := &http.Client{}
	err := fetchContents(apiURL, path, &projectData, gitIgnore, bar, client, githubToken, includeGit, includeNonText)
	if err != nil {
		return ProjectData{}, err
	}

	bar.Finish()

	sort.Strings(projectData.Directories)
	sort.Slice(projectData.Files, func(i, j int) bool {
		return projectData.Files[i].Path < projectData.Files[j].Path
	})

	return projectData, nil
}

func fetchContents(url, path string, projectData *ProjectData, gitIgnore *ignore.GitIgnore, bar *progressbar.ProgressBar, client *http.Client, githubToken string, includeGit, includeNonText bool) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if githubToken != "" {
		req.Header.Set("Authorization", "token "+githubToken)
	}

	resp, err := client.Do(req)
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
		var singleFile GithubContent
		err = json.Unmarshal(body, &singleFile)
		if err != nil {
			return err
		}
		contents = []GithubContent{singleFile}
	}

	for _, content := range contents {
		if !includeGit && (strings.HasPrefix(content.Path, ".git/") || content.Path == ".git") {
			continue
		}

		if content.Type == "dir" {
			projectData.Directories = append(projectData.Directories, content.Path)
			err = fetchContents(content.URL, content.Path, projectData, gitIgnore, bar, client, githubToken, includeGit, includeNonText)
			if err != nil {
				return err
			}
		} else if content.Type == "file" {
			if MatchesPatterns(content.Path, gitIgnore, includeGit, includeNonText) {
				fileContent, err := fetchFileContent(content.DownloadURL, client, githubToken)
				if err != nil {
					return err
				}
				projectData.Files = append(projectData.Files, FileData{Path: content.Path, Content: fileContent})
			} else {
				projectData.Files = append(projectData.Files, FileData{Path: content.Path, Content: ""})
			}
			bar.Describe(fmt.Sprintf("Downloaded: %s", filepath.Base(content.Path)))
			bar.Add(1)
			fmt.Printf("Processed: %s\n", filepath.Base(content.Path))
		}
	}

	return nil
}

func fetchFileContent(url string, client *http.Client, githubToken string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if githubToken != "" {
		req.Header.Set("Authorization", "token "+githubToken)
	}

	resp, err := client.Do(req)
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

func FetchUserRepos(username, githubToken string) ([]GithubRepo, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if githubToken != "" {
		req.Header.Set("Authorization", "token "+githubToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []GithubRepo
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func ParseGitHubURL(url string) (owner string, repo string, path string, err error) {
	// Remove any leading "https://" or "http://"
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Remove any leading "github.com/"
	url = strings.TrimPrefix(url, "github.com/")

	parts := strings.Split(url, "/")

	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid GitHub URL or repository format: %s", url)
	}

	owner = parts[0]
	repo = parts[1]

	if len(parts) > 2 {
		if parts[2] == "tree" && len(parts) > 3 {
			path = strings.Join(parts[4:], "/")
		} else {
			path = strings.Join(parts[2:], "/")
		}
	}

	return owner, repo, path, nil
}
