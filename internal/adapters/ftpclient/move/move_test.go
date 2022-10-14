package move_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/move"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
	ftpclientMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpclient"
	connectionMocks "github.com/alexZaicev/go-ftp-client/mocks/domain/connection"
	useCaseMocks "github.com/alexZaicev/go-ftp-client/mocks/usecases/ftp"
)

const (
	address  = "10.0.0.1:21"
	user     = "user01"
	password = "pwd01"
	timeout  = 5 * time.Second
	oldPath  = "/foo/bar/baz"
	newPath  = "/baz/bar/foo"
)

func Test_PerformMove_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectInfo("OK!")

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.MoveRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}

	useCaseInput := &ftp.MoveInput{
		OldPath: oldPath,
		NewPath: newPath,
	}

	useCaseMock := useCaseMocks.NewMoveUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &move.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &move.CmdMoveInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		OldPath:  oldPath,
		NewPath:  newPath,
	}

	// act
	err := move.PerformMove(ctx, logger, deps, input)

	// assert
	assert.NoError(t, err)
}

func Test_PerformMove_ConnectError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to connect to server").
		WithError(assertlogging.EqualError("mock error"))

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(nil, errors.New("mock error")).
		Once()

	useCaseMock := useCaseMocks.NewMoveUseCase(t)

	buffer := bytes.NewBufferString("")

	deps := &move.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &move.CmdMoveInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		OldPath:  oldPath,
		NewPath:  newPath,
	}

	// act
	err := move.PerformMove(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformMove_ConnectStopError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectInfo("OK!")
	logger.
		ExpectError("failed to stop server connection").
		WithError(assertlogging.EqualError("mock error"))

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(errors.New("mock error")).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.MoveRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.MoveInput{
		OldPath: oldPath,
		NewPath: newPath,
	}

	useCaseMock := useCaseMocks.NewMoveUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &move.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &move.CmdMoveInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		OldPath:  oldPath,
		NewPath:  newPath,
	}

	// act
	err := move.PerformMove(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformMove_UseCaseError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.MoveRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.MoveInput{
		OldPath: oldPath,
		NewPath: newPath,
	}

	useCaseMock := useCaseMocks.NewMoveUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(errors.New("mock error")).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &move.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &move.CmdMoveInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		OldPath:  oldPath,
		NewPath:  newPath,
	}

	// act
	err := move.PerformMove(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}
