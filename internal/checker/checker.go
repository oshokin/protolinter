// Package checker provides functionality for checking and validating
// Protocol Buffer (protobuf) files and their contents against predefined rules and conventions.
package checker

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/oshokin/protolinter/internal/config"
	"github.com/oshokin/protolinter/internal/parser"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// MethodHasVersion checks whether a method specifies a version.
	MethodHasVersion = "method_has_version"
	// MethodHasCorrectInputName checks if the method input is named correctly.
	MethodHasCorrectInputName = "method_has_correct_input_name"
	// MethodHasHTTPPath checks if an HTTP path is specified for the method.
	MethodHasHTTPPath = "method_has_http_path"
	// MethodHasBodyTag checks if methods with a required body have the correct body tag.
	MethodHasBodyTag = "method_has_body_tag"
	// MethodHasSwaggerTags checks if a method has appropriate Swagger tags.
	MethodHasSwaggerTags = "method_has_swagger_tags"
	// MethodHasSwaggerSummary checks if a method has a valid Swagger summary.
	MethodHasSwaggerSummary = "method_has_swagger_summary"
	// MethodHasSwaggerDescription checks if a method has a valid Swagger description.
	MethodHasSwaggerDescription = "method_has_swagger_description"
	// MethodHasDefaultErrorResponse checks if a default error response is defined in the method.
	MethodHasDefaultErrorResponse = "method_has_default_error_response"
	// FieldNameIsSnakeCase checks if a field's name is in snake_case.
	FieldNameIsSnakeCase = "field_name_is_snake_case"
	// FieldHasCorrectJSONName checks if a field's JSON name tag is correct.
	FieldHasCorrectJSONName = "field_has_correct_json_name"
	// FieldHasNoDescription checks if a field has no description.
	FieldHasNoDescription = "field_has_no_description"
	// FieldDescriptionStartsWithCapital checks if a field's description starts with a capital letter.
	FieldDescriptionStartsWithCapital = "field_description_starts_with_capital"
	// FieldDescriptionEndsWithDotOrQuestionMark checks if a field's description ends with a dot or a question mark.
	FieldDescriptionEndsWithDotOrQuestionMark = "field_description_ends_with_dot_or_question_mark"
	// EnumValueHasComments checks if an enum value has leading comments.
	EnumValueHasComments = "enum_value_has_comments"
)

const (
	snakeCaseNamePattern   = "^[a-z][a-z0-9_]*$"
	validMethodNamePattern = `^[A-Z][A-Za-z0-9]*V\d+$`
)

var (
	snakeCaseNameRegexp   = regexp.MustCompile(snakeCaseNamePattern)
	validMethodNameRegexp = regexp.MustCompile(validMethodNamePattern)
)

// NewProtoChecker creates a new ProtoChecker instance.
func NewProtoChecker(ctx context.Context, cfg *config.Config, filesCount int) *ProtoChecker {
	result := &ProtoChecker{
		config:     cfg,
		cache:      make(map[string][]byte, filesCount),
		cacheMutex: &sync.Mutex{},
	}

	result.compiler = &protocompile.Compiler{
		Resolver:       protocompile.WithStandardImports(result.getSourceResolver(ctx, cfg)),
		SourceInfoMode: protocompile.SourceInfoExtraComments | protocompile.SourceInfoExtraOptionLocations,
	}

	return result
}

// CheckFiles performs check operation on the provided protobuf files and returns
// a list of OperationResult instances, each containing the results for a single file.
func (c *ProtoChecker) CheckFiles(ctx context.Context, files []string) ([]*OperationResult, error) {
	return c.processFiles(ctx, files, c.checkFile)
}

func (c *ProtoChecker) processFiles(
	ctx context.Context,
	files []string,
	processor func(linker.File) *OperationResult,
) ([]*OperationResult, error) {
	result := make([]*OperationResult, 0, len(files))

	for _, file := range files {
		parsedFiles, err := c.compiler.Compile(ctx, file)
		if err != nil {
			return nil, fmt.Errorf("failed to compile file %s: %w", file, err)
		}

		parsedFile := parsedFiles.FindFileByPath(file)
		if parsedFile == nil {
			continue
		}

		result = append(result, processor(parsedFile))
	}

	return result, nil
}

func (c *ProtoChecker) checkFile(parsedFile linker.File) *OperationResult {
	result := NewOperationResult(parsedFile, c.config)
	packageName := string(parsedFile.Package().Name())
	parsedFileFullName := string(parsedFile.FullName())

	if c.config.GetPrintAllDescriptors() {
		result.descriptors = append(result.descriptors, parsedFileFullName)
	}

	if c.shouldDescriptorBeSkipped(parsedFileFullName) {
		if c.config.GetVerboseMode() {
			result.AddMessagef("Package %s is skipped", packageName)
		}

		return result
	}

	c.checkServices(parsedFile.Services(), result, parsedFileFullName)
	c.checkMessages(parsedFile.Messages(), result, parsedFile)
	c.checkEnums(parsedFile.Enums(), result, parsedFile)

	return result
}

func (c *ProtoChecker) checkServices(
	services protoreflect.ServiceDescriptors,
	result *OperationResult,
	parsedFileFullName string,
) {
	servicesCount := services.Len()
	for serviceIndex := 0; serviceIndex < servicesCount; serviceIndex++ {
		service := services.Get(serviceIndex)
		serviceName := string(service.Name())
		serviceFullName := string(service.FullName())

		if c.config.GetPrintAllDescriptors() {
			result.descriptors = append(result.descriptors, serviceFullName)
		}

		if c.shouldDescriptorBeSkipped(serviceFullName) {
			if c.config.GetVerboseMode() {
				result.AddMessagef("Service %s is skipped", serviceName)
			}

			continue
		}

		c.checkMethods(service.Methods(), result, serviceName, servicesCount, parsedFileFullName)
	}
}

func (c *ProtoChecker) checkMethods(methods protoreflect.MethodDescriptors,
	result *OperationResult,
	serviceName string,
	servicesCount int,
	parsedFileFullName string,
) {
	for methodIndex := 0; methodIndex < methods.Len(); methodIndex++ {
		method := methods.Get(methodIndex)
		methodName := string(method.Name())
		methodFullName := string(method.FullName())
		methodLogName := c.getNameForLogs(
			parsedFileFullName,
			serviceName,
			servicesCount,
			methodFullName)

		if c.config.GetPrintAllDescriptors() {
			result.descriptors = append(result.descriptors, methodFullName)
		}

		if c.shouldDescriptorBeSkipped(methodFullName) {
			if c.config.GetVerboseMode() {
				result.AddMessagef("Method %s is skipped", methodLogName)
			}

			continue
		}

		isMethodNameCorrect := len(validMethodNameRegexp.FindStringIndex(methodName)) > 0
		if !c.config.IsCheckExcluded(MethodHasVersion) &&
			!isMethodNameCorrect {
			result.AddErrorf(
				method,
				"Name of method %s doesn't match regular expression: %s",
				methodLogName,
				validMethodNamePattern)

			if !c.config.GetPrintAllDescriptors() {
				result.descriptors = append(result.descriptors, methodFullName)
			}
		}

		inputName := string(method.Input().Name())
		inputFullName := string(method.Input().FullName())

		if !c.config.IsCheckExcluded(MethodHasCorrectInputName) &&
			isMethodNameCorrect &&
			inputFullName != "google.protobuf.Empty" {
			expectedInputName := strings.Join([]string{methodName, "Request"}, "")

			if inputName != expectedInputName {
				result.AddErrorf(
					method,
					"Input of method %s should be named as %s",
					methodLogName,
					expectedInputName)

				if !c.config.GetPrintAllDescriptors() {
					result.descriptors = append(result.descriptors, methodFullName)
				}
			}
		}

		c.checkMethodOptions(method, result, methodFullName, methodLogName)
	}
}

func (c *ProtoChecker) checkMethodOptions(
	method protoreflect.Descriptor,
	result *OperationResult,
	methodFullName string,
	methodLogName string,
) {
	method.Options().ProtoReflect().Range(
		func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			optionFullName := string(fd.FullName())

			switch optionFullName {
			case "google.api.http":
				parsedOptions, err := parser.ParseProtoMessageValues(v.Message())
				if err != nil {
					result.AddMessagef(
						"Failed to parse option %s of method %s: %s",
						optionFullName,
						methodLogName,
						err.Error())

					return true
				}

				path := c.fillGoogleAPIHTTPPath(parsedOptions)
				if !c.config.IsCheckExcluded(MethodHasHTTPPath) &&
					path == "" {
					result.AddErrorf(
						method,
						"Path of method %s is not specified",
						methodLogName)

					if !c.config.GetPrintAllDescriptors() {
						result.descriptors = append(result.descriptors, methodFullName)
					}
				}

				if !c.config.IsCheckExcluded(MethodHasBodyTag) &&
					c.isMethodWithRequiredBody(parsedOptions) &&
					parsedOptions.Get("body") != "*" {
					result.AddErrorf(
						method,
						"Method %s doesn't have body tag or body is not equal to *",
						methodLogName)

					if !c.config.GetPrintAllDescriptors() {
						result.descriptors = append(result.descriptors, methodFullName)
					}
				}
			case "grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation":
				parsedOptions, err := parser.ParseProtoMessageValues(v.Message())
				if err != nil {
					result.AddMessagef(
						"Failed to parse option %s of method %s: %s",
						optionFullName,
						methodLogName,
						err.Error())

					return true
				}

				if !c.config.IsCheckExcluded(MethodHasSwaggerTags) &&
					parsedOptions.Get("tags") == "" {
					result.AddErrorf(
						method,
						"Method %s has no swagger tags",
						methodLogName)

					if !c.config.GetPrintAllDescriptors() {
						result.descriptors = append(result.descriptors, methodFullName)
					}
				}

				if !c.config.IsCheckExcluded(MethodHasSwaggerSummary) &&
					parsedOptions.Get("summary") == "" {
					result.AddErrorf(
						method,
						"Method %s has no swagger summary",
						methodLogName)

					if !c.config.GetPrintAllDescriptors() {
						result.descriptors = append(result.descriptors, methodFullName)
					}
				}

				if !c.config.IsCheckExcluded(MethodHasSwaggerDescription) &&
					parsedOptions.Get("description") == "" {
					result.AddErrorf(
						method,
						"Method %s has no swagger description",
						methodLogName)

					if !c.config.GetPrintAllDescriptors() {
						result.descriptors = append(result.descriptors, methodFullName)
					}
				}

				if !c.config.IsCheckExcluded(MethodHasDefaultErrorResponse) {
					if parsedOptions.Get("responses[default]") == "" {
						result.AddErrorf(
							method,
							"Method %s doesn't have a default error response",
							methodLogName)
					}
				}
			}

			return true
		})
}

func (c *ProtoChecker) checkMessages(
	messages protoreflect.MessageDescriptors,
	result *OperationResult,
	parsedFile linker.File,
) {
	parsedFileFullName := string(parsedFile.FullName())

	for messageIndex := 0; messageIndex < messages.Len(); messageIndex++ {
		message := messages.Get(messageIndex)
		messageFullName := string(message.FullName())

		messageLogName := c.getNameForLogs(
			parsedFileFullName,
			"",
			0,
			messageFullName)

		if c.config.GetPrintAllDescriptors() {
			result.descriptors = append(result.descriptors, messageFullName)
		}

		if c.shouldDescriptorBeSkipped(messageFullName) {
			if c.config.GetVerboseMode() {
				result.AddMessagef("Message %s is skipped", messageLogName)
			}

			continue
		}

		c.checkMessageFields(message.Fields(), result, parsedFileFullName)
		c.checkMessages(message.Messages(), result, parsedFile)
		c.checkEnums(message.Enums(), result, parsedFile)
	}
}

func (c *ProtoChecker) checkMessageFields(
	fields protoreflect.FieldDescriptors,
	result *OperationResult,
	parsedFileFullName string,
) {
	for fieldIndex := 0; fieldIndex < fields.Len(); fieldIndex++ {
		field := fields.Get(fieldIndex)
		fieldName := string(field.Name())
		fieldFullName := string(field.FullName())

		fieldLogName := c.getNameForLogs(
			parsedFileFullName,
			"",
			0,
			fieldFullName)

		if c.config.GetPrintAllDescriptors() {
			result.descriptors = append(result.descriptors, fieldFullName)
		}

		if c.shouldDescriptorBeSkipped(fieldFullName) {
			if c.config.GetVerboseMode() {
				result.AddMessagef("Field %s is skipped", fieldLogName)
			}

			continue
		}

		isFieldNameSnakeCase := len(snakeCaseNameRegexp.FindStringIndex(fieldName)) > 0
		if !c.config.IsCheckExcluded(FieldNameIsSnakeCase) &&
			!isFieldNameSnakeCase {
			result.AddErrorf(
				field,
				"Name of field %s doesn't match regular expression: %s",
				fieldLogName,
				snakeCaseNamePattern)

			if !c.config.GetPrintAllDescriptors() {
				result.descriptors = append(result.descriptors, fieldFullName)
			}
		}

		fieldJSONName := field.JSONName()
		if !c.config.IsCheckExcluded(FieldHasCorrectJSONName) &&
			field.HasJSONName() &&
			fieldName != fieldJSONName {
			result.AddErrorf(
				field,
				"Field %s has incorrect json_name tag",
				fieldLogName)

			if !c.config.GetPrintAllDescriptors() {
				result.descriptors = append(result.descriptors, fieldFullName)
			}
		}

		c.checkFieldOptions(field, result, fieldFullName, fieldLogName)
	}
}

func (c *ProtoChecker) checkFieldOptions(field protoreflect.FieldDescriptor,
	result *OperationResult,
	fieldFullName string,
	fieldLogName string,
) {
	field.Options().ProtoReflect().Range(
		func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			optionFullName := string(fd.FullName())

			if optionFullName != "grpc.gateway.protoc_gen_openapiv2.options.openapiv2_field" {
				return true
			}

			parsedOptions, err := parser.ParseProtoMessageValues(v.Message())
			if err != nil {
				result.AddMessagef(
					"Failed to parse option %s of field %s: %s",
					optionFullName,
					fieldLogName,
					err.Error())

				return true
			}

			fieldDescription := parsedOptions.Get("description")
			if !c.config.IsCheckExcluded(FieldHasNoDescription) &&
				fieldDescription == "" {
				result.AddErrorf(
					field,
					"Field %s in doesn't have description",
					fieldLogName)

				if !c.config.GetPrintAllDescriptors() {
					result.descriptors = append(result.descriptors, fieldFullName)
				}
			}

			if !c.config.IsCheckExcluded(FieldDescriptionStartsWithCapital) &&
				fieldDescription != "" &&
				!startsWithCapitalLetter(fieldDescription) {
				result.AddErrorf(
					field,
					"Description of field %s doesn't start with capital letter",
					fieldLogName)

				if !c.config.GetPrintAllDescriptors() {
					result.descriptors = append(result.descriptors, fieldFullName)
				}
			}

			if !c.config.IsCheckExcluded(FieldDescriptionEndsWithDotOrQuestionMark) &&
				fieldDescription != "" &&
				!strings.HasSuffix(fieldDescription, ".") &&
				!strings.HasSuffix(fieldDescription, "?") {
				result.AddErrorf(
					field,
					"Description of field %s must end with dot or question mark",
					fieldLogName)

				if !c.config.GetPrintAllDescriptors() {
					result.descriptors = append(result.descriptors, fieldFullName)
				}
			}

			return true
		})
}

func (c *ProtoChecker) checkEnums(
	enums protoreflect.EnumDescriptors,
	result *OperationResult,
	parsedFile linker.File,
) {
	parsedFileFullName := string(parsedFile.FullName())

	for enumIndex := 0; enumIndex < enums.Len(); enumIndex++ {
		enum := enums.Get(enumIndex)
		enumFullName := string(enum.FullName())
		enumLogName := c.getNameForLogs(
			parsedFileFullName,
			"",
			0,
			enumFullName)

		if c.config.GetPrintAllDescriptors() {
			result.descriptors = append(result.descriptors, enumFullName)
		}

		if c.shouldDescriptorBeSkipped(enumFullName) {
			if c.config.GetVerboseMode() {
				result.AddMessagef("Enum %s is skipped", enumLogName)
			}

			continue
		}

		enumValues := enum.Values()

		for enumValueIndex := 0; enumValueIndex < enumValues.Len(); enumValueIndex++ {
			var (
				enumValue         = enumValues.Get(enumValueIndex)
				enumValueName     = string(enumValue.Name())
				enumValueFullName = string(enumValue.FullName())
				enumValueLogName  = strings.Join([]string{enumLogName, enumValueName}, ".")
			)

			if c.config.GetPrintAllDescriptors() {
				result.descriptors = append(result.descriptors, enumValueFullName)
			}

			if c.shouldDescriptorBeSkipped(enumValueFullName) {
				if c.config.GetVerboseMode() {
					result.AddMessagef("Enum value %s is skipped", enumValueLogName)
				}

				continue
			}

			if !c.config.IsCheckExcluded(EnumValueHasComments) {
				var (
					enumValueSL              = parsedFile.SourceLocations().ByDescriptor(enumValue)
					noEnumValueCommentsFound bool
				)

				if enumValueSL.Path == nil || strings.TrimSpace(enumValueSL.LeadingComments) == "" {
					noEnumValueCommentsFound = true
				}

				if noEnumValueCommentsFound {
					result.AddErrorf(
						enumValue,
						"Enum value %s has no leading comments",
						enumValueLogName)

					if !c.config.GetPrintAllDescriptors() {
						result.descriptors = append(result.descriptors, enumValueFullName)
					}
				}
			}
		}
	}
}

func (c *ProtoChecker) getNameForLogs(
	packageName,
	serviceName string,
	servicesCount int,
	fullName string,
) string {
	if packageName != "" {
		fullName = strings.TrimPrefix(
			fullName,
			strings.Join([]string{packageName, "."}, ""))
	}

	if servicesCount == 1 && serviceName != "" {
		fullName = strings.TrimPrefix(
			fullName,
			strings.Join([]string{serviceName, "."}, ""))
	}

	return fullName
}

func (c *ProtoChecker) shouldDescriptorBeSkipped(name string) bool {
	for _, exception := range c.config.GetExcludedDescriptors() {
		if strings.HasPrefix(name, exception) {
			return true
		}
	}

	return false
}

func (c *ProtoChecker) fillGoogleAPIHTTPPath(params url.Values) string {
	for k, v := range params {
		switch k {
		case "get", "put", "post", "delete", "patch":
			if len(v) > 0 {
				return v[0]
			}

			return ""
		}
	}

	return ""
}

func (c *ProtoChecker) isMethodWithRequiredBody(values url.Values) bool {
	return values.Has("post") || values.Has("put")
}
