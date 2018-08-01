package uuid

import (
	"testing"
)

func TestParseError(t *testing.T) {
	type testrow struct {
		testName   string
		typeName   string
		methodName string
		input      []byte
		isText     bool
		detailFmt  string
		detailArgs []interface{}
		expected   string
	}
	data := []testrow{
		{
			testName: "empty",
			expected: `uuid: failed to parse`,
		},
		{
			testName:   "top-level function",
			methodName: "Func",
			expected:   `uuid.Func: failed to parse`,
		},
		{
			testName:   "method",
			typeName:   "Class",
			methodName: "Func",
			expected:   `uuid.Class.Func: failed to parse`,
		},
		{
			testName:   "basic text",
			typeName:   "Class",
			methodName: "Func",
			input:      []byte(`some text`),
			isText:     true,
			expected:   `uuid.Class.Func: failed to parse "some text"`,
		},
		{
			testName:   "basic bytes",
			typeName:   "Class",
			methodName: "Func",
			input:      []byte{0xde, 0xad, 0xca, 0xfe},
			isText:     false,
			expected:   `uuid.Class.Func: failed to parse [de ad ca fe]`,
		},
		{
			testName:   "detail",
			typeName:   "Class",
			methodName: "Func",
			input:      []byte{0xde, 0xad, 0xca, 0xfe},
			isText:     false,
			detailFmt:  "%q %#02x",
			detailArgs: []interface{}{"string", 64},
			expected:   `uuid.Class.Func: failed to parse [de ad ca fe]: "string" 0x40`,
		},
		{
			testName:   "detail no input",
			typeName:   "Class",
			methodName: "Func",
			detailFmt:  "%q %#02x",
			detailArgs: []interface{}{"string", 64},
			expected:   `uuid.Class.Func: failed to parse: "string" 0x40`,
		},
	}

	for _, row := range data {
		t.Run(row.testName, func(t *testing.T) {
			err := makeParseError(row.typeName, row.methodName, row.input, row.isText)
			if row.detailFmt != "" {
				err = err.detailf(row.detailFmt, row.detailArgs...)
			}
			actual := err.Error()
			if row.expected != actual {
				t.Errorf("wrong error message: expected %q, got %q", row.expected, actual)
			}
		})
	}
}

func TestTypeError(t *testing.T) {
	type testrow struct {
		testName   string
		typeName   string
		methodName string
		input      interface{}
		dummies    []interface{}
		expected   string
	}
	data := []testrow{
		{
			testName: "empty",
			expected: `uuid: wrong type`,
		},
		{
			testName:   "top-level function",
			methodName: "Func",
			expected:   `uuid.Func: wrong type`,
		},
		{
			testName:   "method",
			typeName:   "Class",
			methodName: "Func",
			expected:   `uuid.Class.Func: wrong type`,
		},
		{
			testName:   "basic type",
			typeName:   "Class",
			methodName: "Func",
			input:      uint64(0),
			expected:   `uuid.Class.Func: wrong type uint64`,
		},
		{
			testName:   "custom type",
			typeName:   "Class",
			methodName: "Func",
			input:      &TypeError{},
			expected:   `uuid.Class.Func: wrong type *uuid.TypeError`,
		},
		{
			testName:   "dummy list",
			typeName:   "Class",
			methodName: "Func",
			input:      uint64(0),
			dummies:    []interface{}{new(bool), new(string), []byte(nil), &TypeError{}},
			expected:   `uuid.Class.Func: wrong type uint64: expected one of *bool *string []uint8 *uuid.TypeError`,
		},
	}

	for _, row := range data {
		t.Run(row.testName, func(t *testing.T) {
			err := makeTypeError(row.typeName, row.methodName, row.input, row.dummies...)
			actual := err.Error()
			if row.expected != actual {
				t.Errorf("wrong error message: expected %q, got %q", row.expected, actual)
			}
		})
	}
}
