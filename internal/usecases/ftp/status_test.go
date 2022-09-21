package ftp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
	connectionMocks "github.com/alexZaicev/go-ftp-client/mocks/domain/connection"
)

func Test_Status_Execute_Success(t *testing.T) {
	ctx := context.Background()

	expectedStatus := &entities.Status{
		RemoteAddress: "127.0.0.1",
		LoggedInUser:  "user01",
		TLSEnabled:    false,
		System:        "UNIX",
	}

	logger := assertlogging.NewLogger(t)
	connMock := connectionMocks.NewConnection(t)
	connMock.On("Status").Return(expectedStatus, nil)

	useCaseRepos := &ftp.StatusRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.StatusInput{}

	useCase := &ftp.Status{}
	status, err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
}

func Test_Status_Execute_Error(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectError("failed to get server status").WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.On("Status").Return(nil, errors.New("mock error"))

	useCaseRepos := &ftp.StatusRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.StatusInput{}

	useCase := &ftp.Status{}
	status, err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	assert.Nil(t, status)
	require.EqualError(t, err, "an internal error occurred: failed to get server status")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}
