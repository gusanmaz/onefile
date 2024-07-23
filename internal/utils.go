package internal

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/lexers"
)

type FileData struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ProjectData struct {
	Directories []string   `json:"directories"`
	Files       []FileData `json:"files"`
}

type GithubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
	URL         string `json:"url"`
}

type GithubRepo struct {
	Name string `json:"name"`
}

func getLanguageFromExtension(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		return ""
	}
	return strings.ToLower(lexer.Config().Name)
}

func isTextFile(path string) bool {
	lexer := lexers.Match(path)
	return lexer != nil
}

func matchesPatterns(path string, includePatterns, excludePatterns []string, includeGit, includeNonText bool) bool {
	if !includeGit && (strings.HasPrefix(path, ".git"+string(os.PathSeparator)) || path == ".git") {
		return false
	}

	for _, pattern := range excludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return false
		}
	}

	for _, pattern := range includePatterns {
		if pattern == "*" {
			return includeNonText || isTextFile(path)
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return includeNonText || isTextFile(path)
		}
		if strings.HasPrefix(pattern, ".") && strings.HasSuffix(path, pattern) {
			return includeNonText || isTextFile(path)
		}
	}
	return false
}

func DumpProject(rootPath string, includePatterns, excludePatterns []string, includeGit, includeNonText bool) (ProjectData, error) {
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
			if includeGit || !strings.HasPrefix(relPath, ".git") {
				projectData.Directories = append(projectData.Directories, relPath)
			}
		} else if matchesPatterns(relPath, includePatterns, excludePatterns, includeGit, includeNonText) {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			projectData.Files = append(projectData.Files, FileData{Path: relPath, Content: string(content)})
		}
		return nil
	})

	if err != nil {
		return ProjectData{}, err
	}

	return projectData, nil
}
func SaveAsJSON(projectData ProjectData, outputPath string) error {
	data, err := json.MarshalIndent(projectData, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputPath, data, 0644)
}

func SaveAsMarkdown(projectData ProjectData, outputPath string) error {
	markdown := GenerateMarkdown(projectData)
	return ioutil.WriteFile(outputPath, []byte(markdown), 0644)
}

func GenerateMarkdown(projectData ProjectData) string {
	var md strings.Builder

	md.WriteString("# Project Structure\n\n")
	md.WriteString("```\n")
	md.WriteString(generateProjectTree(projectData))
	md.WriteString("```\n\n")

	md.WriteString("## Shell Commands to Create Project Structure\n\n")
	md.WriteString("```bash\n")
	md.WriteString(GenerateShellCommands(projectData))
	md.WriteString("```\n\n")

	md.WriteString("## File Contents\n\n")
	for _, file := range projectData.Files {
		language := getLanguageFromExtension(file.Path)
		md.WriteString(fmt.Sprintf("### %s\n\n```%s\n%s\n```\n\n", file.Path, language, file.Content))
	}

	return md.String()
}

func generateProjectTree(projectData ProjectData) string {
	var tree strings.Builder
	tree.WriteString(".\n")

	allPaths := append(projectData.Directories, getFilePaths(projectData.Files)...)
	sort.Strings(allPaths)

	for i, path := range allPaths {
		parts := strings.Split(path, string(os.PathSeparator))
		for j, part := range parts {
			isLast := i == len(allPaths)-1 && j == len(parts)-1
			prefix := strings.Repeat("│   ", j)
			if isLast {
				tree.WriteString(fmt.Sprintf("%s└── %s\n", prefix, part))
			} else {
				tree.WriteString(fmt.Sprintf("%s├── %s\n", prefix, part))
			}
		}
	}

	return tree.String()
}

func getFilePaths(files []FileData) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}
	return paths
}

func GenerateShellCommands(projectData ProjectData) string {
	var commands strings.Builder

	for _, dir := range projectData.Directories {
		commands.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", dir))
	}

	for _, file := range projectData.Files {
		dir := filepath.Dir(file.Path)
		if dir != "." {
			commands.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", dir))
		}
		commands.WriteString(fmt.Sprintf("touch \"%s\"\n", file.Path))
	}

	return commands.String()
}

func ParseGitHubURL(url string) (owner string, repo string, path string, err error) {
	parts := strings.Split(url, "/")
	if len(parts) < 5 || parts[2] != "github.com" {
		return "", "", "", fmt.Errorf("invalid GitHub URL: %s", url)
	}
	owner = parts[3]
	repo = parts[4]
	path = ""
	if len(parts) > 5 {
		if parts[5] == "tree" && len(parts) > 7 {
			path = strings.Join(parts[7:], "/")
		} else if parts[5] != "tree" {
			path = strings.Join(parts[5:], "/")
		}
	}
	return owner, repo, path, nil
}

func FetchGithubRepo(owner, repo, path string, includePatterns, excludePatterns []string, useGit bool, githubToken string, includeGit, includeNonText bool) (ProjectData, error) {
	if useGit {
		return fetchWithGit(owner, repo, path, includePatterns, excludePatterns, includeGit, includeNonText)
	}
	return fetchWithAPI(owner, repo, path, includePatterns, excludePatterns, githubToken, includeGit, includeNonText)
}

func fetchWithGit(owner, repo, path string, includePatterns, excludePatterns []string, includeGit, includeNonText bool) (ProjectData, error) {
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
	return DumpProject(projectPath, includePatterns, excludePatterns, includeGit, includeNonText)
}

func fetchWithAPI(owner, repo, path string, includePatterns, excludePatterns []string, githubToken string, includeGit, includeNonText bool) (ProjectData, error) {
	var projectData ProjectData
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)

	bar := progressbar.Default(-1, "Fetching repository")

	client := &http.Client{}
	err := fetchContents(apiURL, path, &projectData, includePatterns, excludePatterns, bar, client, githubToken, includeGit, includeNonText)
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

func fetchContents(url, path string, projectData *ProjectData, includePatterns, excludePatterns []string, bar *progressbar.ProgressBar, client *http.Client, githubToken string, includeGit, includeNonText bool) error {
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
			err = fetchContents(content.URL, content.Path, projectData, includePatterns, excludePatterns, bar, client, githubToken, includeGit, includeNonText)
			if err != nil {
				return err
			}
		} else if content.Type == "file" && matchesPatterns(content.Path, includePatterns, excludePatterns, includeGit, includeNonText) {
			fileContent, err := fetchFileContent(content.DownloadURL, client, githubToken)
			if err != nil {
				return err
			}
			projectData.Files = append(projectData.Files, FileData{Path: content.Path, Content: fileContent})
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

func FetchPyPIPackage(packageName string) (ProjectData, error) {
	var projectData ProjectData

	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)
	resp, err := http.Get(url)
	if err != nil {
		return projectData, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return projectData, fmt.Errorf("PyPI API returned status code %d", resp.StatusCode)
	}

	var pypiData struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
		Urls []struct {
			Filename string `json:"filename"`
			URL      string `json:"url"`
		} `json:"urls"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pypiData)
	if err != nil {
		return projectData, err
	}

	if len(pypiData.Urls) == 0 {
		return projectData, fmt.Errorf("no download URL found for package %s", packageName)
	}

	// Prefer source distribution (.tar.gz) if available, otherwise use the first URL
	var packageURL string
	for _, url := range pypiData.Urls {
		if strings.HasSuffix(url.Filename, ".tar.gz") {
			packageURL = url.URL
			break
		}
	}
	if packageURL == "" {
		packageURL = pypiData.Urls[0].URL
	}

	resp, err = http.Get(packageURL)
	if err != nil {
		return projectData, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return projectData, fmt.Errorf("failed to download package from %s", packageURL)
	}

	// Create a temporary file to store the downloaded package
	tmpFile, err := ioutil.TempFile("", "pypi-package-*")
	if err != nil {
		return projectData, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy the downloaded content to the temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return projectData, err
	}

	// Rewind the file for reading
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return projectData, err
	}

	// Handle different package formats
	if strings.HasSuffix(packageURL, ".tar.gz") {
		return extractTarGz(tmpFile)
	} else if strings.HasSuffix(packageURL, ".whl") {
		return extractWheel(tmpFile)
	} else {
		return projectData, fmt.Errorf("unsupported package format: %s", packageURL)
	}
}

func extractTarGz(file *os.File) (ProjectData, error) {
	var projectData ProjectData

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return projectData, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return projectData, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			projectData.Directories = append(projectData.Directories, header.Name)
		case tar.TypeReg:
			content, err := ioutil.ReadAll(tr)
			if err != nil {
				return projectData, err
			}
			projectData.Files = append(projectData.Files, FileData{Path: header.Name, Content: string(content)})
		}
	}

	return projectData, nil
}

func extractWheel(file *os.File) (ProjectData, error) {
	var projectData ProjectData

	fileInfo, err := file.Stat()
	if err != nil {
		return projectData, err
	}

	r, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return projectData, err
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			projectData.Directories = append(projectData.Directories, f.Name)
		} else {
			rc, err := f.Open()
			if err != nil {
				return projectData, err
			}
			content, err := ioutil.ReadAll(rc)
			rc.Close()
			if err != nil {
				return projectData, err
			}
			projectData.Files = append(projectData.Files, FileData{Path: f.Name, Content: string(content)})
		}
	}

	return projectData, nil
}
