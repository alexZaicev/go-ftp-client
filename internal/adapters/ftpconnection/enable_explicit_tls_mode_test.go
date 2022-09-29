package ftpconnection_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

func Test_ServerConnection_EnableExplicitTLSMode_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandAuthTLS).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusAuthOK).
		Return(0, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.EnableExplicitTLSMode()
	assert.NoError(t, err)
}

func Test_ServerConnection_EnableExplicitTLSMode_CmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandAuthTLS).
		Return(uid, errors.New("mock error")).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandQuit).
		Return(uid, nil).
		Once()
	connMock.
		On("Close").
		Return(nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.EnableExplicitTLSMode()
	require.EqualError(t, err, "an internal error occurred: failed to enable explicit TLS mode")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_EnableExplicitTLSMode_StopError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandAuthTLS).
		Return(uid, errors.New("mock error")).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandQuit).
		Return(uid, nil).
		Once()
	connMock.
		On("Close").
		Return(errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.EnableExplicitTLSMode()
	require.EqualError(t, err, "an internal error occurred: failed to close connection")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
