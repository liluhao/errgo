package errGo

import (
	"fmt"
	"io"
	"net/http"
)

// ErrorType is the type of an error
type Type uint

const (
	// NoType error
	NoType Type = iota
	// BadRequest error
	BadRequest
	// NotFound error
	NotFound
	// default error type
	InternalServerError
)

// The stack printing in errors will print many source paths
// as well, but it is not necessary, I also want to format the
// method call stack, so I write a separate property to do that
// the function I want  —— Stack
type CustomError struct {
	ErrorType     Type
	OriginalError error
	Context       Context
}

type Context struct {
	Field 		string
	Message		string
}

func HandleError(err error) (status int) {
	errorType := GetType(err)
	switch errorType {
	case BadRequest:
		status = http.StatusBadRequest
	case NotFound:
		status = http.StatusNotFound
	default:
		status = http.StatusInternalServerError
	}
	return
}

// TODO Optimization: Print the same in-stack frame to an error structure.
//  The reason for the problem is that each time a new or wrap is used, the stack
//  frame information is refetched, and the data is repeated in each structure.
//  meanwhile
//  The Formatter interface is implemented recursively, causing the withStack
//  structure to be reprinted every time.
// New creates a new customError
func (errorType Type) New(msg string) error {
	return &CustomError{
		ErrorType:     errorType,
		OriginalError: new(msg),
	}
}

// New creates a new customError with formatted message
// The Errorf method in the FMT package is used to format the error message.
// You can pass in multiple strings to add placeholders to the error message.
func (errorType Type) Newf(msg string, args ...interface{}) error {
	return &CustomError{
		ErrorType: errorType,
		OriginalError: fmt.Errorf(msg, args...),
	}
}

// Wrap creates a new wrapped error
func (errorType Type) Wrap(err error, msg string) error {
	return errorType.Wrapf(err, msg)
}

// Wrap creates a new wrapped error with formatted message
func (errorType Type) Wrapf(err error, msg string, args ...interface{}) error {
	return &CustomError{
		ErrorType:     errorType,
		OriginalError: wrapf(err, msg, args...),
	}
}

// Error returns the message of a customError
func (error CustomError) Error() string  {
	return error.OriginalError.Error()
}

// New creates a no type error
func New(msg string) error {
	return &CustomError{
		ErrorType:     NoType,
		OriginalError: new(msg),
	}
}

// Newf creates a no type error with formatted message
func Newf(msg string, args ...interface{}) error {
	return &CustomError{
		ErrorType:     NoType,
		OriginalError: new(fmt.Sprintf(msg, args...)),
	}
}

// Wrap an error with a string
func Wrap(err error, msg string) error {
	return Wrapf(err, msg)
}

// Wrapf an error with format string
func Wrapf(err error, msg string, args ...interface{}) error {
	wrappedError := wrapf(err, msg, args...)
	if customError, ok := err.(*CustomError); ok {
		return &CustomError{
			ErrorType:     customError.ErrorType,
			OriginalError: wrappedError,
			Context:       customError.Context,
		}
	}
	return &CustomError{
		ErrorType:     NoType,
		OriginalError: wrappedError,
	}
}


// AddErrorContext adds a context to an error
func AddErrorContext(err error, field, message string) error {
	context := Context{
		Field:   field,
		Message: message,
	}
	if customError, ok := err.(*CustomError); ok {
		return &CustomError{
			ErrorType:     customError.ErrorType,
			OriginalError: customError.OriginalError,
			Context:       context,
		}
	}
	return &CustomError{
		ErrorType:     NoType,
		OriginalError: err,
		Context:       context,
	}
}

// GetErrorContext returns the error context
func GetErrorContext(err error) map[string]string {
	emptyContext := Context{}
	if customErr, ok := err.(*CustomError); ok || customErr.Context != emptyContext {
		return map[string]string{
			"field": customErr.Context.Field,
			"message": customErr.Context.Message,
		}
	}
	return nil
}

// GetType returns the error type
func GetType(err error) Type {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.ErrorType
	}
	return NoType
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type causer interface {
//            Cause() error
//     }
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
//func CauseIF(err error) error {
//	type causer interface {
//		Cause() error
//	}
//
//	for err != nil {
//		cause, ok := err.(causer)
//		if !ok {
//			break
//		}
//		err = cause.Cause()
//	}
//	return err
//}
// In one of the articles I referred to, this place is written this way to
// implement the(Cause interface of github.com/pkg/errors) interface.
// The methods in this interface abstract a function to recursively assert
// the lowest level of error message.
// However,the same interface is used here and it won`t work, so I need to
// rewrite a new interface.
// Cause gives the original error

func (error CustomError) Cause() error {
	return cause(error.OriginalError)
}

func Cause(err error) error {
	return cause(err)
}



// Inherit from the Formatter interface and use the fmt.Printf function
// to print the desired error message.
// %s,%v Function the same, output error message, does not contain the stack.
// %q The output error message is quoted and does not contain the stack.
// %+v Output error message and stack
// If there are any loop-wrapped error types, these errors will be recursively
// printed.
func (error CustomError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", error.OriginalError)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, error.Error())
	case 'q':
		fmt.Fprintf(s, "%q", error.Error())
	}
}