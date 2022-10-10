package status_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/status"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
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
)

func Test_PerformStatus_Success(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	expectedStatus := &entities.Status{
		RemoteAddress: address[:len(address)-3],
		LoggedInUser:  user,
		TLSEnabled:    true,
		System:        "UNIX",
	}

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

	useCaseRepos := &ftp.StatusRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.StatusInput{}

	useCaseMock := useCaseMocks.NewStatusUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(expectedStatus, nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &status.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &status.CmdStatusInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}

	expectedStatusStr := `+--------+--------+----------------+----------------+-------------+
| STATUS | SYSTEM | REMOTE ADDRESS | LOGGED IN USER | TLS ENABLED |
+--------+--------+----------------+----------------+-------------+
| OK     | UNIX   | 10.0.0.1       | user01         | YES         |
+--------+--------+----------------+----------------+-------------+
`

	err := status.PerformStatus(ctx, logger, deps, input)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatusStr, buffer.String())
}

func Test_PerformStatus_ConnectError(t *testing.T) {
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

	useCaseMock := useCaseMocks.NewStatusUseCase(t)

	buffer := bytes.NewBufferString("")

	deps := &status.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &status.CmdStatusInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}

	err := status.PerformStatus(ctx, logger, deps, input)
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformStatus_ConnectStopError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to stop server connection").
		WithError(assertlogging.EqualError("mock error"))

	expectedStatus := &entities.Status{
		RemoteAddress: address[:len(address)-3],
		LoggedInUser:  user,
		TLSEnabled:    true,
		System:        "UNIX",
	}

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

	useCaseRepos := &ftp.StatusRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.StatusInput{}

	useCaseMock := useCaseMocks.NewStatusUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(expectedStatus, nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &status.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &status.CmdStatusInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}

	err := status.PerformStatus(ctx, logger, deps, input)
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformStatus_UseCaseError(t *testing.T) {
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

	useCaseRepos := &ftp.StatusRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.StatusInput{}

	useCaseMock := useCaseMocks.NewStatusUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil, errors.New("mock error")).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &status.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &status.CmdStatusInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}

	err := status.PerformStatus(ctx, logger, deps, input)
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}
