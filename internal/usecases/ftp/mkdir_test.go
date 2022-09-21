package ftp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
	connectionMocks "github.com/alexZaicev/go-ftp-client/mocks/domain/connection"
)

func Test_Mkdir_Execute_Success(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	connMock := connectionMocks.NewConnection(t)
	connMock.On("Mkdir", dirPath).Return(nil)

	useCaseRepos := &ftp.MkdirRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.MkdirInput{
		Path: dirPath,
	}

	useCase := &ftp.Mkdir{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	assert.NoError(t, err)
}

func Test_Mkdir_Execute_Error(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectError("failed to create directory").WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.On("Mkdir", dirPath).Return(errors.New("mock error"))

	useCaseRepos := &ftp.MkdirRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.MkdirInput{
		Path: dirPath,
	}

	useCase := &ftp.Mkdir{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	require.EqualError(t, err, "an internal error occurred: failed to create directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}
