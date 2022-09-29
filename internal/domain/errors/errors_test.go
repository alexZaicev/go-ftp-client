package errors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func Test_NewInternalError_Success(t *testing.T) {
	err := ftperrors.NewInternalError("hello world", errors.New("mock error"))
	assert.EqualError(t, err, "an internal error occurred: hello world")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_NewInvalidArgumentError_Success(t *testing.T) {
	err := ftperrors.NewInvalidArgumentError("arg1", "hello world")
	assert.EqualError(t, err, "an invalid argument error occurred: argument arg1 hello world")
	assert.IsType(t, ftperrors.InvalidArgumentErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_NewUnknownError_Success(t *testing.T) {
	err := ftperrors.NewUnknownError("hello world", errors.New("mock error"))
	assert.EqualError(t, err, "an unknown error occurred: hello world")
	assert.IsType(t, ftperrors.UnknownErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_NewNotFoundError_Success(t *testing.T) {
	err := ftperrors.NewNotFoundError("hello world", errors.New("mock error"))
	assert.EqualError(t, err, "not found error occurred: hello world")
	assert.IsType(t, ftperrors.NotFoundErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
