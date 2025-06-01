package checker

import (
	"fmt"

	"github.com/bufbuild/protocompile/linker"
	"github.com/oshokin/protolinter/internal/config"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewOperationResult creates a new OperationResult based on the given parsed file and configuration.
func NewOperationResult(parsedFile linker.File, cfg *config.Config) *OperationResult {
	return &OperationResult{
		File:   parsedFile,
		config: cfg,
	}
}

// AddMessage appends an informational message to the OperationResult's messages.
func (c *OperationResult) AddMessage(v string) {
	c.Messages = append(c.Messages, v)
}

// AddMessagef appends a formatted informational message to the OperationResult's messages.
func (c *OperationResult) AddMessagef(format string, args ...any) {
	c.Messages = append(c.Messages, fmt.Sprintf(format, args...))
}

// AddError appends an error message to the OperationResult's errors.
func (c *OperationResult) AddError(desc protoreflect.Descriptor, v string) {
	c.Errors = append(c.Errors, c.appendErrorLocation(desc, v))
}

// AddErrorf appends a formatted error message to the OperationResult's errors.
func (c *OperationResult) AddErrorf(desc protoreflect.Descriptor, format string, args ...any) {
	c.AddError(desc, fmt.Sprintf(format, args...))
}

// appendErrorLocation appends error location information to the error message if available.
func (c *OperationResult) appendErrorLocation(desc protoreflect.Descriptor, message string) string {
	var (
		fileSourceLocations = c.File.SourceLocations()
		sl                  = fileSourceLocations.ByDescriptor(desc)
		line                int
		column              int
	)

	if sl.Path != nil {
		line = sl.StartLine + 1
		column = sl.StartColumn + 1
	}

	if line > 0 && column > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", c.File.Path(), line, column, message)
	}

	return message
}
