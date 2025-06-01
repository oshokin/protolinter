// Package parser provides functions for parsing and encoding Protobuf messages into URL values.
package parser

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ParseProtoMessageValues translates values from a Protobuf message
// into a map of string slices (url.Values).
func ParseProtoMessageValues(msg protoreflect.Message) (url.Values, error) {
	u := make(url.Values)

	err := encodeByField(u, "", msg)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func parseProtoMessageFieldValue(
	fieldDescriptor protoreflect.FieldDescriptor,
	value protoreflect.Value,
) (string, error) {
	switch fieldDescriptor.Kind() {
	case protoreflect.BoolKind:
		return strconv.FormatBool(value.Bool()), nil
	case protoreflect.EnumKind:
		if fieldDescriptor.Enum().FullName() == "google.protobuf.NullValue" {
			return "null", nil
		}

		desc := fieldDescriptor.Enum().Values().ByNumber(value.Enum())

		return string(desc.Name()), nil
	case protoreflect.StringKind:
		return value.String(), nil
	case protoreflect.BytesKind:
		return base64.URLEncoding.EncodeToString(value.Bytes()), nil
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return encodeMessage(fieldDescriptor.Message(), value)
	default:
		return fmt.Sprint(value.Interface()), nil
	}
}

func encodeByField(u url.Values, path string, m protoreflect.Message) error {
	var finalErr error

	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		key := fd.JSONName()
		if !fd.HasJSONName() {
			key = fd.TextName()
		}

		newPath := key
		if path != "" {
			newPath = strings.Join([]string{path, key}, ".")
		}

		if of := fd.ContainingOneof(); of != nil {
			if f := m.WhichOneof(of); f != nil && f != fd {
				return true
			}
		}

		switch {
		case fd.IsList():
			if v.List().Len() == 0 {
				return true
			}

			list, err := encodeRepeatedField(fd, v.List())
			if err != nil {
				finalErr = err
				return false
			}

			for _, item := range list {
				u.Add(newPath, item)
			}
		case fd.IsMap():
			if v.Map().Len() == 0 {
				return true
			}

			m := encodeMapField(fd, v.Map())
			for k, value := range m {
				u.Set(fmt.Sprintf("%s[%s]", newPath, k), value)
			}
		case (fd.Kind() == protoreflect.MessageKind) || (fd.Kind() == protoreflect.GroupKind):
			value, err := encodeMessage(fd.Message(), v)
			if err == nil {
				u.Set(newPath, value)
				return true
			}

			if err = encodeByField(u, newPath, v.Message()); err != nil {
				finalErr = err
				return false
			}
		default:
			value, err := parseProtoMessageFieldValue(fd, v)
			if err != nil {
				finalErr = err
				return false
			}

			u.Set(newPath, value)
		}

		return true
	})

	return finalErr
}

func encodeRepeatedField(fieldDescriptor protoreflect.FieldDescriptor, list protoreflect.List) ([]string, error) {
	valuesCount := list.Len()

	if valuesCount == 0 {
		return nil, nil
	}

	values := make([]string, 0, valuesCount)

	for i := 0; i < valuesCount; i++ {
		value, err := parseProtoMessageFieldValue(fieldDescriptor, list.Get(i))
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

func encodeMapField(fieldDescriptor protoreflect.FieldDescriptor, mp protoreflect.Map) map[string]string {
	result := make(map[string]string, mp.Len())

	mp.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		key, err := parseProtoMessageFieldValue(fieldDescriptor.MapValue(), k.Value())
		if err != nil {
			return false
		}

		value, err := parseProtoMessageFieldValue(fieldDescriptor.MapValue(), v)
		if err != nil {
			return false
		}

		result[key] = value

		return true
	})

	return result
}

func encodeMessage(msgDescriptor protoreflect.MessageDescriptor, value protoreflect.Value) (string, error) {
	switch msgDescriptor.FullName() {
	case timestampMessageFullname:
		return marshalTimestamp(value.Message())
	case durationMessageFullname:
		return marshalDuration(value.Message())
	case bytesMessageFullname:
		return marshalBytes(value.Message())
	case "google.protobuf.DoubleValue",
		"google.protobuf.FloatValue",
		"google.protobuf.Int64Value",
		"google.protobuf.Int32Value",
		"google.protobuf.UInt64Value",
		"google.protobuf.UInt32Value",
		"google.protobuf.BoolValue",
		"google.protobuf.StringValue":
		fd := msgDescriptor.Fields()
		v := value.Message().Get(fd.ByName("value"))

		return fmt.Sprint(v.Interface()), nil
	case fieldMaskFullName:
		m, ok := value.Message().Interface().(*fieldmaskpb.FieldMask)
		if !ok || m == nil {
			return "", nil
		}

		for i, v := range m.Paths {
			m.Paths[i] = convertSnakeCaseToCamelCase(v)
		}

		return strings.Join(m.Paths, ","), nil
	case responseFieldFullName, responseEntryFullName:
		return value.String(), nil
	default:
		return "", fmt.Errorf("unsupported message type: %q", string(msgDescriptor.FullName()))
	}
}

// https://github.com/protocolbuffers/protobuf-go/blob/master/encoding/protojson/well_known_types.go#L842
func convertSnakeCaseToCamelCase(s string) string {
	var (
		b                 []byte
		isUnderscoreFound bool
	)

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c != '_' {
			isASCIILowerCaseLetter := 'a' <= c && c <= 'z'
			if isUnderscoreFound && isASCIILowerCaseLetter {
				c -= 'a' - 'A'
			}

			b = append(b, c)
		}

		isUnderscoreFound = c == '_'
	}

	return string(b)
}
