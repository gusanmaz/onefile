package internal

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
