package checker

import (
	"fmt"

	"github.com/bufbuild/protocompile/linker"
	"github.com/oshokin/protolinter/internal/config"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewCheckResult creates a new CheckResult based on the given parsed file and configuration.
func NewCheckResult(parsedFile linker.File, cfg *config.Config) *CheckResult {
	return &CheckResult{
		File:   parsedFile,
		config: cfg,
	}
}

// AddMessage appends an informational message to the CheckResult's messages.
func (c *CheckResult) AddMessage(v string) {
	c.Messages = append(c.Messages, v)
}

// AddMessagef appends a formatted informational message to the CheckResult's messages.
func (c *CheckResult) AddMessagef(format string, args ...any) {
	c.Messages = append(c.Messages, fmt.Sprintf(format, args...))
}

// AddError appends an error message to the CheckResult's errors.
func (c *CheckResult) AddError(desc protoreflect.Descriptor, v string) {
	message := v
	if !c.config.GetOmitCoordinates() {
		message = c.appendErrorLocation(desc, v)
	}

	c.Errors = append(c.Errors, message)
}

// AddErrorf appends a formatted error message to the CheckResult's errors.
func (c *CheckResult) AddErrorf(desc protoreflect.Descriptor, format string, args ...any) {
	c.AddError(desc, fmt.Sprintf(format, args...))
}

// appendErrorLocation appends error location information to the error message if available.
func (c *CheckResult) appendErrorLocation(desc protoreflect.Descriptor, message string) string {
	var (
		fileSourceLocations = c.File.SourceLocations()
		sl                  = fileSourceLocations.ByDescriptor(desc)
		row                 int
		column              int
	)

	if sl.Path != nil {
		row = sl.StartLine
		column = sl.StartColumn
	}

	if row > 0 && column > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", c.File.Path(), row, column, message)
	}

	return message
}
