# Protolinter

Protolinter is a command-line tool to lint and analyze Protocol Buffer files for compliance with coding conventions, best practices, and standards.\
It empowers developers to ensure that their Protocol Buffer files are well-formed, consistent, and adhere to recommended guidelines.

## Features

- Lint and analyze Protocol Buffer files.
- Verify compliance with coding conventions and standards.
- Customizable through `.protolinter.yaml` configuration file.
- Supports exclusion of specific checks and descriptors.
- Generate a list of full protobuf element names.

## Installation

You can install Protocol Linter using Go's `go install` command:

```sh
go install github.com/oshokin/protolinter/cmd/protolinter
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

# Generate a list of full protobuf element names
protolinter list <file.proto>
```

## Configuration

Protolinter supports configuration through a .protolinter.yaml file.\
If the configuration file is absent, Protolinter will work with default settings, performing all checks and not excluding any proto descriptors from analysis.\
You can define excluded checks and descriptors to customize the analysis according to your project's needs.\
An example configuration file can be found in `.protolinter.example.yaml`.

## Checks Performed

Protolinter performs various checks on your Protocol Buffer files to ensure their compliance.\
The following checks can be excluded from analysis in the configuration file:

- `method_has_version`: Checks whether a method specifies a version.
- `method_has_correct_input_name`: Checks if the method input is named correctly.
- `method_has_http_path`: Checks if an HTTP path is specified for the method.
- `method_has_body_tag`: Checks if methods with a required body have the correct body tag.
- `method_has_swagger_tags`: Checks if a method has appropriate Swagger tags.
- `method_has_swagger_summary`: Checks if a method has a valid Swagger summary.
- `method_has_swagger_description`: Checks if a method has a valid Swagger description.
- `field_has_correct_json_name`: Checks if a field's JSON name tag is correct.
- `field_has_no_description`: Checks if a field has no description.
- `field_description_starts_with_capital`: Checks if a field's description starts with a capital letter.
- `field_description_ends_with_dot`: Checks if a field's description ends with a dot.
- `enum_value_has_comments`: Checks if an enum value has leading comments.

## Translations

[Документация на русском языке](README.ru.md)