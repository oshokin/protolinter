package checker

import (
	"context"
	"os"
	"path/filepath"

	"github.com/oshokin/protolinter/internal/config"
	"github.com/oshokin/protolinter/internal/logger"
)

// ExecuteCheck runs the "check" subcommand.
func ExecuteCheck(patterns []string, configPath string, isMimirFile bool) {
	ctx := context.Background()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fatalf(ctx, "Failed to load configuration: %s", err.Error())
	}

	var files []string
	if isMimirFile {
		files, err = extractFilesFromMimir(patterns[0])
	} else {
		files, err = extractFilesFromPatterns(patterns, "")
	}

	if err != nil {
		logger.Fatalf(ctx, "Failed to locate files based on the provided patterns: %s", err.Error())
	}

	if len(files) == 0 {
		logger.Fatal(ctx, "List of files is empty")
	}

	checker := NewProtoChecker(ctx, cfg)

	results, err := checker.CheckFiles(ctx, files...)
	if err != nil {
		logger.Fatalf(ctx, "Failed to perform checks on files: %s", err.Error())
	}

	processCheckResults(ctx, results)
}

// ExecuteListProtoFullNames runs the "lint" subcommand.
func ExecuteListProtoFullNames(patterns []string) {
	ctx := context.Background()

	files, err := extractFilesFromPatterns(patterns, "")
	if err != nil {
		logger.Fatalf(ctx, "Failed to locate files based on the provided patterns: %s", err.Error())
	}

	if len(files) == 0 {
		logger.Fatal(ctx, "List of files is empty")
	}

	checker := NewProtoChecker(ctx, nil)

	results, err := checker.ListFullNamesFromFiles(ctx, files...)
	if err != nil {
		logger.Fatalf(ctx, "Failed to list full names: %s", err.Error())
	}

	processListResults(ctx, results)
}

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

			if extension != "" && filepath.Ext(file) != extension {
				continue
			}

			result = append(result, file)
		}
	}

	return result, nil
}

func processCheckResults(ctx context.Context, results []*CheckResult) {
	var isCheckFailed bool

	for _, cr := range results {
		if len(cr.Messages) == 0 && len(cr.Errors) == 0 {
			continue
		}

		if len(cr.Errors) > 0 {
			isCheckFailed = true
		}

		logger.Infof(ctx, "Checking file %s:", cr.File.Path())

		for _, message := range cr.Messages {
			logger.Info(ctx, message)
		}

		for _, message := range cr.Errors {
			logger.Error(ctx, message)
		}
	}

	if isCheckFailed {
		os.Exit(1)
	}
}

func processListResults(ctx context.Context, results []*ListResult) {
	for _, lr := range results {
		if len(lr.Messages) == 0 {
			continue
		}

		logger.Infof(ctx, "Listing protobuf elements names from file %s:", lr.File.Path())

		for _, message := range lr.Messages {
			logger.Info(ctx, message)
		}
	}
}
