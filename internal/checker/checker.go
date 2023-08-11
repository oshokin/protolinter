package checker

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

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
	// FieldHasCorrectJSONName checks if a field's JSON name tag is correct.
	FieldHasCorrectJSONName = "field_has_correct_json_name"
	// FieldHasNoDescription checks if a field has no description.
	FieldHasNoDescription = "field_has_no_description"
	// FieldDescriptionStartsWithCapital checks if a field's description starts with a capital letter.
	FieldDescriptionStartsWithCapital = "field_description_starts_with_capital"
	// FieldDescriptionEndsWithDot checks if a field's description ends with a dot.
	FieldDescriptionEndsWithDot = "field_description_ends_with_dot"
	// EnumValueHasComments checks if an enum value has leading comments.
	EnumValueHasComments = "enum_value_has_comments"
)

const validMethodNamePattern = `^[A-Z][A-Za-z0-9]*V\d+$`

var validMethodNameRegexp = regexp.MustCompile(validMethodNamePattern)

// NewProtoChecker creates a new ProtoChecker instance.
func NewProtoChecker(ctx context.Context, cfg *config.Config) *ProtoChecker {
	result := &ProtoChecker{
		compiler: &protocompile.Compiler{
			Resolver:       protocompile.WithStandardImports(getSourceResolver(ctx, cfg)),
			SourceInfoMode: protocompile.SourceInfoExtraComments | protocompile.SourceInfoExtraOptionLocations,
		},
	}

	result.config = cfg

	return result
}

// CheckFiles performs checks on the provided protobuf files and returns
// a list of CheckResult instances, each containing the checking results for a single file.
// It uses the compiler and parser associated with the ProtoChecker instance.
func (c *ProtoChecker) CheckFiles(ctx context.Context, files ...string) ([]*CheckResult, error) {
	parsedFiles, err := c.compiler.Compile(ctx, files...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile files %s: %w", files, err)
	}

	result := make([]*CheckResult, 0, len(parsedFiles))

	for _, parsedFile := range parsedFiles {
		result = append(result, c.checkFile(parsedFile))
	}

	return result, nil
}

func (c *ProtoChecker) checkFile(parsedFile linker.File) *CheckResult {
	result := NewCheckResult(parsedFile, c.config)
	packageName := string(parsedFile.Package().Name())
	parsedFileFullName := string(parsedFile.FullName())

	if c.shouldDescriptorBeSkipped(parsedFileFullName) {
		result.AddMessagef("Package %s is skipped", packageName)

		return result
	}

	c.checkServices(parsedFile.Services(), result, parsedFileFullName)
	c.checkMessages(parsedFile.Messages(), result, parsedFile)
	c.checkEnums(parsedFile.Enums(), result, parsedFile)

	return result
}

func (c *ProtoChecker) checkServices(
	services protoreflect.ServiceDescriptors,
	result *CheckResult,
	parsedFileFullName string,
) {
	servicesCount := services.Len()
	for serviceIndex := 0; serviceIndex < servicesCount; serviceIndex++ {
		service := services.Get(serviceIndex)
		serviceName := string(service.Name())
		serviceFullName := string(service.FullName())

		if c.shouldDescriptorBeSkipped(serviceFullName) {
			result.AddMessagef("Service %s is skipped", serviceName)

			continue
		}

		c.checkMethods(service.Methods(), result, serviceName, servicesCount, parsedFileFullName)
	}
}

func (c *ProtoChecker) checkMethods(methods protoreflect.MethodDescriptors,
	result *CheckResult,
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

		if c.shouldDescriptorBeSkipped(methodFullName) {
			result.AddMessagef("Method %s is skipped", methodLogName)

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
			}
		}

		c.checkMethodOptions(method, result, methodLogName)
	}
}

func (c *ProtoChecker) checkMethodOptions(
	method protoreflect.Descriptor,
	result *CheckResult,
	methodLogName string,
) {
	method.Options().ProtoReflect().Range(
		func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			optionFullName := string(fd.FullName())
			optionMessage := v.Message()

			switch optionFullName {
			case "google.api.http":
				parsedOptions, err := parser.ParseProtoMessageValues(optionMessage)
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
				}

				if !c.config.IsCheckExcluded(MethodHasBodyTag) &&
					c.isMethodWithRequiredBody(parsedOptions) &&
					parsedOptions.Get("body") != "*" {
					result.AddErrorf(
						method,
						"Method %s doesn't have body tag or body is not equal to *",
						methodLogName)
				}
			case "grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation":
				parsedOptions, err := parser.ParseProtoMessageValues(optionMessage)
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
				}

				if !c.config.IsCheckExcluded(MethodHasSwaggerSummary) &&
					parsedOptions.Get("summary") == "" {
					result.AddErrorf(
						method,
						"Method %s has no swagger summary",
						methodLogName)
				}

				if !c.config.IsCheckExcluded(MethodHasSwaggerDescription) &&
					parsedOptions.Get("description") == "" {
					result.AddErrorf(
						method,
						"Method %s has no swagger description",
						methodLogName)
				}
			}

			return true
		})
}

func (c *ProtoChecker) checkMessages(
	messages protoreflect.MessageDescriptors,
	result *CheckResult,
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

		if c.shouldDescriptorBeSkipped(messageFullName) {
			result.AddMessagef("Message %s is skipped", messageLogName)

			continue
		}

		c.checkMessageFields(message.Fields(), result, parsedFileFullName)
		c.checkMessages(message.Messages(), result, parsedFile)
		c.checkEnums(message.Enums(), result, parsedFile)
	}
}

func (c *ProtoChecker) checkMessageFields(
	fields protoreflect.FieldDescriptors,
	result *CheckResult,
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

		if c.shouldDescriptorBeSkipped(fieldFullName) {
			result.AddMessagef("Field %s is skipped", fieldLogName)

			continue
		}

		fieldJSONName := field.JSONName()
		if !c.config.IsCheckExcluded(FieldHasCorrectJSONName) &&
			field.HasJSONName() &&
			fieldName != fieldJSONName {
			result.AddErrorf(
				field,
				"Field %s has incorrect json_name tag",
				fieldLogName)
		}

		c.checkFieldOptions(field, result, fieldLogName)
	}
}

func (c *ProtoChecker) checkFieldOptions(field protoreflect.FieldDescriptor,
	result *CheckResult,
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
			}

			if !c.config.IsCheckExcluded(FieldDescriptionStartsWithCapital) &&
				fieldDescription != "" &&
				!startsWithCapitalLetter(fieldDescription) {
				result.AddErrorf(
					field,
					"Description of field %s doesn't start with capital letter",
					fieldLogName)
			}

			if !c.config.IsCheckExcluded(FieldDescriptionEndsWithDot) &&
				fieldDescription != "" &&
				!strings.HasSuffix(fieldDescription, ".") {
				result.AddErrorf(
					field,
					"Description of field %s must end with dot",
					fieldLogName)
			}

			return true
		})
}

func (c *ProtoChecker) checkEnums(
	enums protoreflect.EnumDescriptors,
	result *CheckResult,
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

		if c.shouldDescriptorBeSkipped(enumFullName) {
			result.AddMessagef("Enum %s is skipped", enumLogName)

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

			if c.shouldDescriptorBeSkipped(enumValueFullName) {
				result.AddMessagef("Enum value %s is skipped", enumValueLogName)

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
