package checker

import (
	"context"
	"fmt"

	"github.com/bufbuild/protocompile/linker"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ListFullNamesFromFiles performs checks on the provided protobuf files and returns
// a list of CheckResult instances, each containing the checking results for a single file.
// It uses the compiler and parser associated with the ProtoChecker instance.
func (c *ProtoChecker) ListFullNamesFromFiles(ctx context.Context, files ...string) ([]*ListResult, error) {
	parsedFiles, err := c.compiler.Compile(ctx, files...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile files %s: %w", files, err)
	}

	result := make([]*ListResult, 0, len(parsedFiles))

	for _, parsedFile := range parsedFiles {
		result = append(result, c.listFullNamesFromFile(parsedFile))
	}

	return result, nil
}

func (c *ProtoChecker) listFullNamesFromFile(parsedFile linker.File) *ListResult {
	result := NewListResult(parsedFile, c.config)
	parsedFileFullName := string(parsedFile.FullName())

	result.AddMessagef("Package: %s", parsedFileFullName)

	services := parsedFile.Services()
	servicesCount := services.Len()

	for serviceIndex := 0; serviceIndex < servicesCount; serviceIndex++ {
		service := services.Get(serviceIndex)
		serviceFullName := string(service.FullName())

		result.AddMessagef("Service: %s", serviceFullName)

		methods := service.Methods()
		for methodIndex := 0; methodIndex < methods.Len(); methodIndex++ {
			method := methods.Get(methodIndex)
			methodFullName := string(method.FullName())

			result.AddMessagef("Method: %s", methodFullName)
		}
	}

	c.listMessagesFullNames(parsedFile.Messages(), result)
	c.listEnumsFullNames(parsedFile.Enums(), result)

	return result
}

func (c *ProtoChecker) listMessagesFullNames(
	messages protoreflect.MessageDescriptors,
	result *ListResult,
) {
	for messageIndex := 0; messageIndex < messages.Len(); messageIndex++ {
		message := messages.Get(messageIndex)
		messageFullName := string(message.FullName())

		result.AddMessagef("Message: %s", messageFullName)

		fields := message.Fields()
		for fieldIndex := 0; fieldIndex < fields.Len(); fieldIndex++ {
			field := fields.Get(fieldIndex)
			fieldFullName := string(field.FullName())

			result.AddMessagef("Field: %s", fieldFullName)
		}

		c.listMessagesFullNames(message.Messages(), result)
		c.listEnumsFullNames(message.Enums(), result)
	}
}

func (c *ProtoChecker) listEnumsFullNames(
	enums protoreflect.EnumDescriptors,
	result *ListResult,
) {
	for enumIndex := 0; enumIndex < enums.Len(); enumIndex++ {
		enum := enums.Get(enumIndex)
		enumFullName := string(enum.FullName())

		result.AddMessagef("Enum: %s", enumFullName)

		enumValues := enum.Values()

		for enumValueIndex := 0; enumValueIndex < enumValues.Len(); enumValueIndex++ {
			enumValue := enumValues.Get(enumValueIndex)
			enumValueFullName := string(enumValue.FullName())

			result.AddMessagef("Enum value: %s", enumValueFullName)
		}
	}
}
