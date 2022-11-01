package download_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/download"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
	ftpclientMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpclient"
	connectionMocks "github.com/alexZaicev/go-ftp-client/mocks/domain/connection"
	useCaseMocks "github.com/alexZaicev/go-ftp-client/mocks/usecases/ftp"
)

const (
	address    = "10.0.0.1:21"
	user       = "user01"
	password   = "pwd01"
	timeout    = 5 * time.Second
	remotePath = "/baz/bar/foo"
	path       = "/foo/bar/baz"
)

func Test_PerformDownload_Success(t *testing.T) {
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

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remotePath,
		Path:       path,
	}

	useCaseMock := useCaseMocks.NewDownloadUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &download.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &download.CmdDownloadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		Path:       path,
		RemotePath: remotePath,
	}

	// act
	err := download.PerformDownload(ctx, logger, deps, input)

	// assert
	assert.NoError(t, err)
}

func Test_PerformDownload_ConnectError(t *testing.T) {
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

	useCaseMock := useCaseMocks.NewDownloadUseCase(t)

	buffer := bytes.NewBufferString("")

	deps := &download.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &download.CmdDownloadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		Path:       path,
		RemotePath: remotePath,
	}

	// act
	err := download.PerformDownload(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformDownload_ConnectStopError(t *testing.T) {
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

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.DownloadInput{
		Path:       path,
		RemotePath: remotePath,
	}

	useCaseMock := useCaseMocks.NewDownloadUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &download.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &download.CmdDownloadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		Path:       path,
		RemotePath: remotePath,
	}

	// act
	err := download.PerformDownload(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformDownload_UseCaseError(t *testing.T) {
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

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.DownloadInput{
		RemotePath: remotePath,
		Path:       path,
	}

	useCaseMock := useCaseMocks.NewDownloadUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(errors.New("mock error")).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &download.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &download.CmdDownloadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		Path:       path,
		RemotePath: remotePath,
	}

	// act
	err := download.PerformDownload(ctx, logger, deps, input)

	// assert
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}
