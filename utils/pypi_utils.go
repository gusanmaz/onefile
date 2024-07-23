package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

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
