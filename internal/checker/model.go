package checker

import (
	"sync"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/oshokin/protolinter/internal/config"
)

type (
	// ProtoChecker represents a structure that wraps the compiler and parser for protobuf files.
	ProtoChecker struct {
		compiler   *protocompile.Compiler // Compiler for protobuf files.
		config     *config.Config         // Configuration for the checker.
		cache      map[string][]byte      // Cache for downloaded files.
		cacheMutex *sync.Mutex            // Mutex to synchronize cache access.
	}

	// OperationResult holds the results of an operation (check or list) on a single protobuf file.
	OperationResult struct {
		File     linker.File // The protobuf file that was operated on.
		Messages []string    // List of informational messages related to the operation.
		Errors   []string    // List of errors. If empty, the operation is considered successful.

		config      *config.Config // Configuration used for the operation.
		descriptors []string       // List of full names of descriptors processed during the operation.
	}

	// ArtifactoryErrorResponse represents the structure of an error response from artifactory.big-freaking-company.com.
	ArtifactoryErrorResponse struct {
		Errors []*ArtifactoryError `json:"errors"` // List of errors
	}

	// ArtifactoryError represents an individual error received from artifactory.big-freaking-company.com.
	ArtifactoryError struct {
		Status  int    `json:"status"`  // HTTP status code of the error.
		Message string `json:"message"` // Message attached to the error.
	}

	subcommandOptions struct {
		patterns            []string
		configPath          string
		githubURL           string
		isMimirFile         bool
		printAllDescriptors bool
	}
)
