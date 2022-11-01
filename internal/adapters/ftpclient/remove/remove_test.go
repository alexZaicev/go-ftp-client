package remove_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/remove"
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
	path     = "/foo/bar/baz"
)

func Test_PerformRemove_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectInfo("OK!")

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: path,
	}

	useCaseMock := useCaseMocks.NewRemoveUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &remove.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &remove.CmdRemoveInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		Path: path,
	}

	// act
	err := remove.PerformRemove(ctx, logger, deps, input)

	// assert
	assert.NoError(t, err)
}

func Test_PerformRemove_ConnectError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to connect to server").
		WithError(assertlogging.EqualError("mock error"))

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(nil, errors.New("mock error")).
		Once()

	useCaseMock := useCaseMocks.NewRemoveUseCase(t)

	buffer := bytes.NewBufferString("")

	deps := &remove.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &remove.CmdRemoveInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		Path: path,
	}

	// act
	err := remove.PerformRemove(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformRemove_ConnectStopError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectInfo("OK!")
	logger.
		ExpectError("failed to stop server connection").
		WithError(assertlogging.EqualError("mock error"))

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(errors.New("mock error")).Once()

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.RemoveInput{
		Path: path,
	}

	useCaseMock := useCaseMocks.NewRemoveUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &remove.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &remove.CmdRemoveInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		Path: path,
	}

	// act
	err := remove.PerformRemove(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformRemove_UseCaseError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.RemoveInput{
		Path: path,
	}

	useCaseMock := useCaseMocks.NewRemoveUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(errors.New("mock error")).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &remove.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &remove.CmdRemoveInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		Path: path,
	}

	// act
	err := remove.PerformRemove(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}
