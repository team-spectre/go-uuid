package uuid

import (
	"fmt"
	"reflect"
)

// ParseError represents an error in the contents of the input while parsing.
type ParseError struct {
	TypeName   string
	MethodName string
	Input      string
	Detail     string
	ExactInput []byte
}

var _ error = ParseError{}

func makeParseError(typeName, methodName string, input []byte, isText bool) ParseError {
	var mangled string
	if isText {
		mangled = quoteString(input)
	} else if input != nil {
		mangled = formatBytes(input)
	}
	return ParseError{
		TypeName:   typeName,
		MethodName: methodName,
		Input:      mangled,
		ExactInput: copyBytes(input),
	}
}

func (err ParseError) detailf(formatString string, args ...interface{}) ParseError {
	err.Detail = fmt.Sprintf(formatString, args...)
	return err
}

func (err ParseError) Error() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()

	w.WriteString("uuid")
	if err.TypeName != "" {
		w.WriteString(".")
		w.WriteString(err.TypeName)
	}
	if err.MethodName != "" {
		w.WriteString(".")
		w.WriteString(err.MethodName)
	}
	w.WriteString(": failed to parse")
	if err.Input != "" {
		w.WriteString(" ")
		w.WriteString(err.Input)
	}
	if err.Detail != "" {
		w.WriteString(": ")
		w.WriteString(err.Detail)
	}
	return w.String()
}

// TypeError represents an error in the Go type of the input while scanning.
type TypeError struct {
	TypeName      string
	MethodName    string
	Input         interface{}
	InputType     reflect.Type
	ExpectedTypes []reflect.Type
}

var _ error = TypeError{}

func makeTypeError(typeName, methodName string, input interface{}, dummies ...interface{}) TypeError {
	inputType := reflect.TypeOf(input)
	expectedTypes := make([]reflect.Type, len(dummies))
	for i, v := range dummies {
		expectedTypes[i] = reflect.TypeOf(v)
	}
	return TypeError{
		TypeName:      typeName,
		MethodName:    methodName,
		Input:         input,
		InputType:     inputType,
		ExpectedTypes: expectedTypes,
	}
}

func (err TypeError) Error() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()

	w.WriteString("uuid")
	if err.TypeName != "" {
		w.WriteString(".")
		w.WriteString(err.TypeName)
	}
	if err.MethodName != "" {
		w.WriteString(".")
		w.WriteString(err.MethodName)
	}
	w.WriteString(": wrong type")
	if err.InputType != nil {
		w.WriteString(" ")
		w.WriteString(err.InputType.String())
	}
	if len(err.ExpectedTypes) != 0 {
		w.WriteString(": expected one of")
		for _, t := range err.ExpectedTypes {
			w.WriteString(" ")
			w.WriteString(t.String())
		}
	}
	return w.String()
}
