package checker

import (
	"context"
	"os"
	"sort"

	"github.com/oshokin/protolinter/internal/config"
	"github.com/oshokin/protolinter/internal/logger"
	"go.uber.org/zap/zapcore"
)

// ExecuteCheck runs the "check" subcommand.
func ExecuteCheck(patterns []string, configPath, githubURL string, isMimirFile bool) {
	ctx := context.Background()
	options := &subcommandOptions{
		patterns:            patterns,
		configPath:          configPath,
		githubURL:           githubURL,
		isMimirFile:         isMimirFile,
		printAllDescriptors: false,
	}

	_, results := runSubcommandOnFiles(ctx, options)

	processResults(ctx, results)
}

// ExecutePrintConfig runs the "print-config" subcommand.
func ExecutePrintConfig(patterns []string, configPath, githubURL string, isMimirFile, printAllDescriptors bool) {
	ctx := context.Background()
	options := &subcommandOptions{
		patterns:            patterns,
		configPath:          configPath,
		githubURL:           githubURL,
		isMimirFile:         isMimirFile,
		printAllDescriptors: printAllDescriptors,
	}

	checker, results := runSubcommandOnFiles(ctx, options)

	configuration, err := generateConfig(checker.config, results)
	if err != nil {
		logger.Fatalf(ctx, "Failed to generate configuration: %s", err.Error())
	}

	logger.Warn(ctx, configuration)
}

func runSubcommandOnFiles(
	ctx context.Context,
	options *subcommandOptions,
) (*ProtoChecker, []*OperationResult) {
	cfg, err := config.LoadConfig(options.configPath, options.githubURL, options.printAllDescriptors)
	if err != nil {
		logger.Fatalf(ctx, "Failed to load configuration: %s", err.Error())
	}

	logLevel := zapcore.WarnLevel
	if cfg.GetVerboseMode() {
		logLevel = zapcore.InfoLevel
	}

	logger.SetLevel(logLevel)

	var files []string
	if options.isMimirFile {
		files, err = extractFilesFromMimir(options.patterns[0])
	} else {
		files, err = extractFilesFromPatterns(options.patterns, "")
	}

	if err != nil {
		logger.Fatalf(ctx, "Failed to locate files based on the provided patterns: %s", err.Error())
	}

	if len(files) == 0 {
		logger.Warn(ctx, "No protobuf files were found based on the provided patterns.")
		os.Exit(0)
	}

	checker := NewProtoChecker(ctx, cfg, len(files))

	results, err := checker.CheckFiles(ctx, files)
	if err != nil {
		logger.Fatalf(ctx, "Failed to perform checks on files: %s", err.Error())
	}

	return checker, results
}

func processResults(ctx context.Context, results []*OperationResult) {
	var isCheckFailed bool

	for _, cr := range results {
		if len(cr.Messages) == 0 && len(cr.Errors) == 0 {
			continue
		}

		if len(cr.Errors) > 0 {
			isCheckFailed = true
		}

		logger.Warnf(ctx, "Processing file %s:", cr.File.Path())

		for _, message := range cr.Messages {
			logger.Warn(ctx, message)
		}

		for _, message := range cr.Errors {
			logger.Error(ctx, message)
		}
	}

	if isCheckFailed {
		os.Exit(1)
	}
}

func generateConfig(cfg *config.Config, results []*OperationResult) (string, error) {
	sourceChecks := cfg.GetExcludedChecks()
	finalChecks := make([]string, len(sourceChecks))

	copy(finalChecks, sourceChecks)

	sort.Strings(finalChecks)

	sourceDescriptors := cfg.GetExcludedDescriptors()
	allDescriptors := make([]string, len(sourceDescriptors))

	copy(allDescriptors, sourceDescriptors)

	for _, result := range results {
		allDescriptors = append(allDescriptors, result.descriptors...) //nolint:makezero // false alarm
	}

	uniqueDescriptors := make(map[string]struct{})
	for _, name := range allDescriptors {
		uniqueDescriptors[name] = struct{}{}
	}

	finalDescriptors := make([]string, 0, len(uniqueDescriptors))

	for name := range uniqueDescriptors {
		finalDescriptors = append(finalDescriptors, name)
	}

	sort.Strings(finalDescriptors)

	// Copy configuration and update ExcludedDescriptors
	newCfg := &config.Config{
		VerboseMode:         cfg.GetVerboseMode(),
		ExcludedChecks:      finalChecks,
		ExcludedDescriptors: finalDescriptors,
	}

	result, err := config.MarshalConfig(newCfg)
	if err != nil {
		return "", err
	}

	return result, nil
}
