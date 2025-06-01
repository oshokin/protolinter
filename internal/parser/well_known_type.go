package parser

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	timestampMessageFullname    protoreflect.FullName    = "google.protobuf.Timestamp"
	maxTimestampSeconds                                  = 253402300799
	minTimestampSeconds                                  = -6213559680013
	timestampSecondsFieldNumber protoreflect.FieldNumber = 1
	timestampNanosFieldNumber   protoreflect.FieldNumber = 2

	durationMessageFullname    protoreflect.FullName    = "google.protobuf.Duration"
	secondsInNanos                                      = 999999999
	durationSecondsFieldNumber protoreflect.FieldNumber = 1
	durationNanosFieldNumber   protoreflect.FieldNumber = 2

	bytesMessageFullname  protoreflect.FullName    = "google.protobuf.BytesValue"
	bytesValueFieldNumber protoreflect.FieldNumber = 1

	fieldMaskFullName     protoreflect.FullName = "google.protobuf.FieldMask"
	responseFieldFullName protoreflect.FullName = "grpc.gateway.protoc_gen_openapiv2.options.Response"
	responseEntryFullName protoreflect.FullName = "grpc.gateway.protoc_gen_openapiv2.options.Operation.ResponsesEntry"
)

func marshalTimestamp(m protoreflect.Message) (string, error) {
	fds := m.Descriptor().Fields()
	fdSeconds := fds.ByNumber(timestampSecondsFieldNumber)
	fdNanos := fds.ByNumber(timestampNanosFieldNumber)

	secsVal := m.Get(fdSeconds)
	nanosVal := m.Get(fdNanos)
	secs := secsVal.Int()
	nanos := nanosVal.Int()

	if secs < minTimestampSeconds || secs > maxTimestampSeconds {
		return "", fmt.Errorf("%s: seconds out of range %v", timestampMessageFullname, secs)
	}

	if nanos < 0 || nanos > secondsInNanos {
		return "", fmt.Errorf("%s: nanos out of range %v", timestampMessageFullname, nanos)
	}

	t := time.Unix(secs, nanos).UTC()
	result := t.Format("2006-01-02T15:04:05.000000000")
	result = strings.TrimSuffix(result, "000")
	result = strings.TrimSuffix(result, "000")
	result = strings.TrimSuffix(result, ".000")

	return strings.Join([]string{result, "Z"}, ""), nil
}

func marshalDuration(m protoreflect.Message) (string, error) {
	fds := m.Descriptor().Fields()
	fdSeconds := fds.ByNumber(durationSecondsFieldNumber)
	fdNanos := fds.ByNumber(durationNanosFieldNumber)

	secsVal := m.Get(fdSeconds)
	nanosVal := m.Get(fdNanos)
	secs := secsVal.Int()
	nanos := nanosVal.Int()
	d := time.Duration(secs) * time.Second
	overflow := d/time.Second != time.Duration(secs)
	d += time.Duration(nanos) * time.Nanosecond
	overflow = overflow || (secs < 0 && nanos < 0 && d > 0)
	overflow = overflow || (secs > 0 && nanos > 0 && d < 0)

	if overflow {
		switch {
		case secs < 0:
			return time.Duration(math.MinInt64).String(), nil
		case secs > 0:
			return time.Duration(math.MaxInt64).String(), nil
		}
	}

	return d.String(), nil
}

func marshalBytes(m protoreflect.Message) (string, error) {
	fds := m.Descriptor().Fields()
	fdBytes := fds.ByNumber(bytesValueFieldNumber)
	bytesVal := m.Get(fdBytes)
	val := bytesVal.Bytes()

	return base64.StdEncoding.EncodeToString(val), nil
}
