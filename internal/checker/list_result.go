package checker

import (
	"fmt"

	"github.com/bufbuild/protocompile/linker"
	"github.com/oshokin/protolinter/internal/config"
)

// NewListResult creates a new ListResult based on the given parsed file and configuration.
func NewListResult(parsedFile linker.File, cfg *config.Config) *ListResult {
	return &ListResult{
		File:   parsedFile,
		config: cfg,
	}
}

// AddMessage appends an informational message to the ListResult's messages.
func (c *ListResult) AddMessage(v string) {
	c.Messages = append(c.Messages, v)
}

// AddMessagef appends a formatted informational message to the ListResult's messages.
func (c *ListResult) AddMessagef(format string, args ...interface{}) {
	c.Messages = append(c.Messages, fmt.Sprintf(format, args...))
}
