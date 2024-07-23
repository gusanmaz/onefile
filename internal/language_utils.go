package internal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
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
	return len(getLanguagesFromFile(path)) > 0
}
