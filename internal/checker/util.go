package checker

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func extractFilesFromPatterns(patterns []string, extension string) ([]string, error) {
	var (
		alreadyAddedFiles = make(map[string]struct{}, len(patterns))
		result            = make([]string, 0, len(patterns))
	)

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if _, ok := alreadyAddedFiles[file]; ok {
				continue
			}

			alreadyAddedFiles[file] = struct{}{}

			fi, _ := os.Stat(file)
			if fi.IsDir() {
				continue
			}

			if extension != "" &&
				!strings.EqualFold(filepath.Ext(file), extension) {
				continue
			}

			result = append(result, file)
		}
	}

	return result, nil
}

func startsWithCapitalLetter(s string) bool {
	if len(s) == 0 {
		return false
	}

	return unicode.IsUpper(rune(s[0]))
}
