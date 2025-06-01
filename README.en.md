# Protolinter

Protolinter is a command-line tool to lint and analyze Protocol Buffer files for compliance with coding conventions, best practices, and standards.\
It empowers developers to ensure that their Protocol Buffer files are well-formed, consistent, and adhere to recommended guidelines.

## Features

- Verify Protocol Buffer files for compliance with coding conventions and standards.
- Customizable through `.protolinter.yaml` configuration file.
- Exclude specific checks and descriptors based on your needs.
- Generate a configuration with excluded protobuf elements.

## Installation

You can install Protocol Linter using Go's `go install` command:

```sh
go install github.com/oshokin/protolinter@latest
```

Alternatively, you can build the executable using the provided Makefile:

```sh
make build
```

The built executable will be located in the `bin` subdirectory.

## Usage

```sh
# Lint and analyze protobuf files
protolinter check [--config=<path>] [--mimir] <file.proto>

# Generate and print a configuration file
# with full names of protobuf elements where errors were found during the check
protolinter print-config [--config=<path>] <file.proto>
```

## Configuration

Protolinter supports configuration through a .protolinter.yaml file.\
If the configuration file is absent, Protolinter will work with default settings, performing all checks and not excluding any proto descriptors from analysis.\
You can define excluded checks and descriptors to customize the analysis according to your project's needs.\
An example configuration file can be found in `.protolinter.example.yaml`.
Additionally, you can generate a configuration file using the following command:
```sh
protolinter print-config -m mimir.yaml > .protolinter.yaml
```
There are two files at the root of this project for the convenience of the developer:\
`Makefile.example` is an example of a Makefile.\
`.gitlab-ci.example.yaml` is an example Gitlab CI/CD job for protolinter.\
You can copy these files to your project and edit them to suit your needs.

## Checks Performed

Protolinter performs various checks on your Protocol Buffer files to ensure their compliance.\
The following checks can be excluded from analysis in the configuration file:

- `enum_value_has_comments`: Checks if an enum value has leading comments.
- `field_description_ends_with_dot_or_question_mark`: Checks if a field's description ends with a dot or a question mark.
- `field_description_starts_with_capital`: Checks if a field's description starts with a capital letter.
- `field_has_correct_json_name`: Checks if a field's JSON name tag is correct.
- `field_has_no_description`: Checks if a field has no description.
- `field_name_is_snake_case`: Checks if a field's name is in snake_case.
- `method_has_body_tag`: Checks if methods with a required body have the correct body tag.
- `method_has_correct_input_name`: Checks if the method input is named correctly.
- `method_has_default_error_response`: Checks if a default error response is defined in the method.
- `method_has_http_path`: Checks if an HTTP path is specified for the method.
- `method_has_swagger_description`: Checks if a method has a valid Swagger description.
- `method_has_swagger_summary`: Checks if a method has a valid Swagger summary.
- `method_has_swagger_tags`: Checks if a method has appropriate Swagger tags.
- `method_has_version`: Checks whether a method specifies a version.

## Translations

[Документация на русском языке](README.md)