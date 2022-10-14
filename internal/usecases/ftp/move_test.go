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

func Test_Move_Execute_Success(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	connMock := connectionMocks.NewConnection(t)
	connMock.On("Move", remoteDirPath, dirPath).Return(nil)

	useCaseRepos := &ftp.MoveRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.MoveInput{
		OldPath: remoteDirPath,
		NewPath: dirPath,
	}

	useCase := &ftp.Move{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	assert.NoError(t, err)
}

func Test_Move_Execute_Error(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectError("failed to move file").WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.On("Move", remoteDirPath, dirPath).Return(errors.New("mock error"))

	useCaseRepos := &ftp.MoveRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.MoveInput{
		OldPath: remoteDirPath,
		NewPath: dirPath,
	}

	useCase := &ftp.Move{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	require.EqualError(t, err, "an internal error occurred: failed to move file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}
