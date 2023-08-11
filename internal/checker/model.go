package checker

import (
	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/oshokin/protolinter/internal/config"
)

type (
	// ProtoChecker represents a structure that
	// wraps the compiler and parser for protobuf files.
	ProtoChecker struct {
		compiler *protocompile.Compiler
		config   *config.Config
	}

	// CheckResult holds the results of checking a single protobuf file.
	CheckResult struct {
		File     linker.File // Checked file.
		Messages []string    // List of informational messages related to the file.
		Errors   []string    // List of errors. If empty, the check is considered successful.
		config   *config.Config
	}

	// ListResult holds the results of listing full protobuf element names.
	ListResult struct {
		File     linker.File // Analyzed file.
		Messages []string    // List of full protobuf element names found in the file.
		config   *config.Config
	}
)
