package internal

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
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

//go:embed extension_language_map.json
var languageMappingJSON []byte

var languageMapping map[string][]string

func init() {
	err := json.Unmarshal(languageMappingJSON, &languageMapping)
	if err != nil {
		fmt.Println("Error parsing embedded language mapping:", err)
	}
}

func getLanguagesFromFile(filename string) []string {
	if languages, ok := languageMapping[filename]; ok {
		return languages
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if languages, ok := languageMapping[ext]; ok {
		return languages
	}
	return nil
}

func getLanguageFromExtension(filename string) string {
	languages := getLanguagesFromFile(filename)
	if len(languages) > 0 {
		return languages[0]
	}
	return ""
}

func isTextFile(path string) bool {
	return len(getLanguagesFromFile(path)) > 0
}

func loadPatternsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if pattern != "" && !strings.HasPrefix(pattern, "#") {
			patterns = append(patterns, pattern)
		}
	}

	return patterns, scanner.Err()
}

func CreateGitIgnoreMatcher(patterns []string) *ignore.GitIgnore {
	return ignore.CompileIgnoreLines(patterns...)
}

func MatchesPatterns(path string, gitIgnore *ignore.GitIgnore, includeGit, includeNonText bool) bool {
	if !includeGit && (strings.HasPrefix(path, ".git"+string(os.PathSeparator)) || path == ".git") {
		return false
	}
	if gitIgnore.MatchesPath(path) {
		return false
	}
	return includeNonText || isTextFile(path)
}

func ParsePatterns(patterns []string) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "@") {
			filePatterns, err := loadPatternsFromFile(strings.TrimPrefix(pattern, "@"))
			if err != nil {
				return nil, err
			}
			result = append(result, filePatterns...)
		} else {
			result = append(result, pattern)
		}
	}
	return result, nil
}

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

	return projectData, nil
}

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

func getFilePaths(files []FileData) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}
	return paths
}

func SaveAsMarkdown(projectData ProjectData, outputPath string, includeGit, includeNonText bool) error {
	markdown := GenerateMarkdown(projectData, includeGit, includeNonText)
	return ioutil.WriteFile(outputPath, []byte(markdown), 0644)
}

func GenerateMarkdown(projectData ProjectData, includeGit, includeNonText bool) string {
	var md strings.Builder

	md.WriteString("# Project Structure\n\n")
	md.WriteString("```\n")
	md.WriteString(generateProjectTree(projectData, includeGit, includeNonText))
	md.WriteString("```\n\n")

	md.WriteString("## Shell Commands to Create Project Structure\n\n")
	md.WriteString("```bash\n")
	md.WriteString(GenerateShellCommands(projectData, includeGit, includeNonText))
	md.WriteString("```\n\n")

	md.WriteString("## File Contents\n\n")
	for _, file := range projectData.Files {
		if file.Content != "" && (includeGit || !strings.HasPrefix(file.Path, ".git/")) && (includeNonText || isTextFile(file.Path)) {
			language := getLanguageFromExtension(file.Path)
			md.WriteString(fmt.Sprintf("### %s\n\n```%s\n%s\n```\n\n", file.Path, language, file.Content))
		}
	}

	return md.String()
}

func generateProjectTree(projectData ProjectData, includeGit, includeNonText bool) string {
	var tree strings.Builder
	tree.WriteString(".\n")

	var allPaths []string
	for _, dir := range projectData.Directories {
		if includeGit || !strings.HasPrefix(dir, ".git") {
			allPaths = append(allPaths, dir)
		}
	}
	for _, file := range projectData.Files {
		if (includeGit || !strings.HasPrefix(file.Path, ".git")) && (includeNonText || isTextFile(file.Path)) {
			allPaths = append(allPaths, file.Path)
		}
	}
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
func GenerateShellCommands(projectData ProjectData, includeGit, includeNonText bool) string {
	var commands strings.Builder

	for _, dir := range projectData.Directories {
		if includeGit || !strings.HasPrefix(dir, ".git") {
			commands.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", dir))
		}
	}

	for _, file := range projectData.Files {
		if (includeGit || !strings.HasPrefix(file.Path, ".git/")) && (includeNonText || isTextFile(file.Path)) {
			dir := filepath.Dir(file.Path)
			if dir != "." {
				commands.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", dir))
			}
			commands.WriteString(fmt.Sprintf("touch \"%s\"\n", file.Path))
		}
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

	tmpFile, err := ioutil.TempFile("", "pypi-package-*")
	if err != nil {
		return projectData, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return projectData, err
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return projectData, err
	}

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
