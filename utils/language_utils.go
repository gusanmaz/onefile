package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

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
	// First, check if it's a known text file type based on extension
	if len(getLanguagesFromFile(path)) > 0 {
		return true
	}

	// If not determined by extension, check the content
	content, err := ioutil.ReadFile(path)
	if err != nil {
		// If we can't read the file, assume it's not text
		return false
	}

	mime := mimetype.Detect(content)
	return strings.HasPrefix(mime.String(), "text/")
}

func IsTextContent(content []byte) bool {
	mime := mimetype.Detect(content)
	return strings.HasPrefix(mime.String(), "text/")
}
