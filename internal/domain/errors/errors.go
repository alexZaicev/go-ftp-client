package errors

import "fmt"

const (
	ErrMsgCannotBeNil   = "cannot be nil"
	ErrMsgCannotBeBlank = "cannot be blank"
)

var (
	InternalErrorType        = &InternalError{}
	InvalidArgumentErrorType = &InvalidArgumentError{}
	UnknownErrorType         = &UnknownError{}
	NotFoundErrorType        = &NotFoundError{}
)

type InternalError struct {
	baseError
}

func NewInternalError(msg string, err error) *InternalError {
	return &InternalError{
		baseError: newBaseError(
			fmt.Sprintf("an internal error occurred: %s", msg),
			err,
		),
	}
}

type InvalidArgumentError struct {
	baseError
}

func NewInvalidArgumentError(arg, msg string) *InvalidArgumentError {
	return &InvalidArgumentError{
		baseError: newBaseError(
			fmt.Sprintf("an invalid argument error occurred: argument %s %s", arg, msg),
			nil,
		),
	}
}

type UnknownError struct {
	baseError
}

func NewUnknownError(msg string, err error) *UnknownError {
	return &UnknownError{
		baseError: newBaseError(
			fmt.Sprintf("an unknown error occurred: %s", msg),
			err,
		),
	}
}

type NotFoundError struct {
	baseError
}

func NewNotFoundError(msg string, err error) *NotFoundError {
	return &NotFoundError{
		baseError: newBaseError(
			fmt.Sprintf("not found error occurred: %s", msg),
			err,
		),
	}
}
