// Package errors provides facilities to create a custom error string format
// allowing to provide additional contextual information.
package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/atdiar/flag"
)

var (
	DEBUG = flag.NewCC()
)

// Error is a type implementing the error interface that can be customized with
// header or trailer values.
// Calling its Error() method returns a string that corresponds to its
// json-serialization.
type Error struct {
	ErrorInfo  map[string]interface{} `json:",omitempty"`
	ErrorCode  string                 `json:"-"`
	ErrorCause string
	Underlying *Error `json:"ErrorSource,omitempty"`
	codec      Codec
}

// Code sets an error code.
func (e *Error) Code(c int) *Error {
	e.ErrorCode = strconv.Itoa(c)
	e.AddInfo("Code", c)
	return e
}

// As tests whether the object implementing the error interface is of type Error.
func As(e error) *Error {
	err, ok := e.(*Error)
	if !ok {
		return nil
	}
	return err
}

// Is compares errors by the
func (e *Error) Is(code int) bool {
	if e == nil {
		return false
	}
	return e.ErrorCode == strconv.Itoa(code)
}

// AddInfo allows to prepend information to an error string.
func (e *Error) AddInfo(key string, value interface{}) *Error {
	if e.ErrorInfo == nil {
		e.ErrorInfo = make(map[string]interface{})
	}
	e.ErrorInfo[key] = value
	return e
}

// Retrieve will extract an Error object from an error interface.
func (e *Error) Retrieve(E error) *Error {
	if E == nil {
		return nil
	}
	if val, ok := E.(*Error); ok {
		return val
	}
	return e.codec.Decode([]byte(E.Error()))
}

func (e *Error) Wraps(E error) *Error {
	ne := *e
	err := ne.Retrieve(E)
	if e == err {
		return e
	}
	e.Underlying = err
	return e
}

// Error is the method allowing the Error type to implement the standard error
// interface.
func (e *Error) Error() string {
	var strErr string
	res, err := e.codec.Encode(e)
	if err != nil {
		strErr = err.Error()
		if DEBUG.IsTrue() {
			// create stacktrace and append it
			buf := make([]byte, 1024)
			runtime.Stack(buf, true)
			strErr = strErr + "\n\n" + fmt.Sprint(string(buf))
		}
		return strErr
	}
	strErr = string(res)
	if DEBUG.IsTrue() {
		// create stacktrace and append it
		buf := make([]byte, 1024)
		runtime.Stack(buf, true)
		strErr = strErr + "\n\nTRACE===========================================\n" + fmt.Sprint(string(buf)) + "\n\n"
	}
	return strErr
}

func (e *Error) String() string {
	return e.ErrorCause
}

// Constructor is a function that allows to create an Error creating function.
// A set of functions that return information key/value pairs can be specified.
// Any Error created will subsequently be decorated with information.
func Constructor(codec Codec, infoHeaderFuncs ...func() (key string, value interface{})) func(string) *Error {
	return func(message string) *Error {
		e := Error{nil, "", message, nil, codec}
		if len(infoHeaderFuncs) == 0 {
			return &e
		}
		e.ErrorInfo = make(map[string]interface{})
		for _, f := range infoHeaderFuncs {
			name, value := f()
			e.ErrorInfo[name] = value
		}
		return &e
	}
}

// Codec defines a pair of functions used to marshall/unmarshall an object of
// type Error.
type Codec struct {
	Encode func(interface{}) ([]byte, error)
	Decode func([]byte) *Error
}

// NewCodec allows the specification of a new codec.
func NewCodec(Enc func(interface{}) ([]byte, error), Dec func([]byte) *Error) Codec {
	return Codec{Enc, Dec}
}

// toJSON will enable the encoding of the bare error string and the additional
// information as a JSON string.
func toJSON(i interface{}) ([]byte, error) {
	return json.MarshalIndent(i, "", " ")
}

// fromJSON enables the decoding of an error string into an Error object.
func fromJSON(b []byte) *Error {
	var e Error
	err := json.Unmarshal(b, &e)
	if err != nil {
		e.ErrorCause = string(b)
		return &e
	}
	return &e
}

/* JSONCodec is an Error Encoder/Decoder object.
var JSONCodec Codec

// New is the default function that returns an Error object.
// It is initialized in an init block.
var New func(message string) *Error

// Defaults */
var (
	JSONCodec = NewCodec(toJSON, fromJSON)
	New       = Constructor(JSONCodec, PrintFile, PrintFunc, PrintLine)
)

// PrintDate returns the Unix formatted Date (UTC) at which an error occured.
func PrintDate() (fieldName string, date interface{}) {
	return "date", time.Now().UTC().Format(time.UnixDate)
}

// PrintLine returns the line number on which the error occured.
func PrintLine() (fieldName string, line interface{}) {
	_, _, line, _ = runtime.Caller(1)
	return "line", line
}

// PrintFile returns the name of the package file in which the error occured.
func PrintFile() (fieldName string, file interface{}) {
	_, file, _, _ = runtime.Caller(0)
	return "file", file
}

// PrintFunc returns the name of the function in which the error occured.
func PrintFunc() (fieldname string, fn interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	fn = f.Name()
	return "fn", fn
}

// PrintTrace returns the name of the function in which the error occured.
func PrintTrace() (fieldname string, funcs interface{}) {
	/*pc := make([]uintptr, 20)

	result := make(map[string]struct {
		File string `json:"file"`
		Line int    `json:"line"`
	}, runtime.Callers(0, pc))

	for _, counter := range pc {
		f := runtime.FuncForPC(counter)
		if f != nil {
			file, line := f.FileLine(0)
			result[f.Name()] = struct {
				File string `json:"file"`
				Line int    `json:"line"`
			}{file, line}
		}
	}
	return "trace", result
	*/
	buf := make([]byte, 1024)
	runtime.Stack(buf, true)
	return "trace", fmt.Sprint(string(buf))
}

// List  defines a datatype holding a list of error values.
type List struct {
	Values []error
}

// NewList returns a new, emptyn container for a list of errors.
func NewList() *List {
	l := new(List)
	l.Values = make([]error, 0)
	return l
}

// Add allows to append an error value to an error list.
func (l *List) Add(e ...error) {
	if l.Values == nil {
		l.Values = make([]error, 0)
	}
	l.Values = append(l.Values, e...)
}

func (l *List) Error() string {
	var s string
	for _, v := range l.Values {
		s = s + v.Error() + "\n"
	}
	return s
}

func (l *List) Nil() bool {
	return len(l.Values) == 0
}

// NOTE While this package defines an error type, the header is entirely customizable.
// People will have to generate their own specification specifying what can be found in
// the header and communicate that spec to a receiving endpoint/service that wants to
// inspect those headers.

// NOTE If one were to send Error values to a service over the wire, one would
// need to audit the error message to make sure that no sensitive information
// is leaked which could lead for instance to a security breach.
