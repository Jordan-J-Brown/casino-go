package errorx

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	//Set to true on initialisation to output stack traces
	OutputStack = false
)

type Code uint16

const (
	Unknown     Code = iota
	ClientError Code = 400
	NotFound    Code = 404
	ServerError Code = 500
)

func (t Code) IsServerError() bool {
	return t == Unknown || t >= ServerError && t < 600
}

func (t Code) IsClientError() bool {
	return t >= ClientError && t < ServerError
}

func (t Code) String() string {
	switch t {
	case ClientError:
		return "Client_Error"
	case NotFound:
		return "Not_Found"
	case ServerError:
		return "Server_Error"
	default:
		return "Unknown"
	}
}

type Error struct {
	outputStack bool
	code        Code
	errType     string
	msg         string
	stack       string
	children    []error
}

func New(msg string, opts ...Option) *Error {
	err := &Error{
		code: Unknown,
		msg:  msg,
	}

	for _, opt := range opts {
		opt(err)
	}

	if OutputStack || err.outputStack {
		stack := debug.Stack()
		size := len(stack)
		lineCount := 0
		split := 0
		for i, char := range stack {
			if char == '\n' {
				lineCount++
			}
			if lineCount == 4 && i+1 < size {
				split = i + 1
			}
		}
		err.stack = string(stack[split:])
	}

	return err
}

func NewFromError(err error, opts ...Option) *Error {
	return New(err.Error(), opts...)
}

func NewFromType(errType string, opts ...Option) *Error {
	opts = append(opts, WithType(errType))
	return New("", opts...)
}

// New error with formatting. Args should be passed first, then options.
func Newf(msg string, argsOrOpts ...any) *Error {
	opts := make([]Option, 0, len(argsOrOpts))
	split := -1
	for i := range argsOrOpts {
		if option, isOption := argsOrOpts[i].(Option); isOption {
			if split == -1 {
				split = i
			}
			opts = append(opts, option)
		}
	}
	if split == -1 {
		split = len(argsOrOpts)
	}

	msg = fmt.Sprintf(msg, argsOrOpts[:split])

	return New(msg, opts...)
}

func (e *Error) Code() Code {
	return e.code
}

func (e *Error) Type() string {
	return e.errType
}

func (e *Error) Message() string {
	return e.msg
}

func (e *Error) Error() string {
	s := &strings.Builder{}

	s.WriteString("Error(")
	s.WriteString(e.code.String())
	if e.errType != "" {
		s.WriteByte(':')
		s.WriteString(e.errType)
	}
	s.WriteString("): ")
	s.WriteString(e.msg)
	s.WriteString("\n\n")
	for i := range e.children {
		fmt.Fprintf(s, "Child Error #%d:\n", i+1)
		s.WriteString(e.children[i].Error())
		s.WriteString("\n\n")
	}
	if OutputStack || e.outputStack {
		s.WriteString("\n\nStack Trace:\n\n")
		s.WriteString(e.stack)
		s.WriteByte('\n')
	}

	return s.String()
}
