package ftpconnection_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

func Test_ServerConnection_Cd_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusRequestedFileActionOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Cd(remotePath)
	assert.NoError(t, err)
}

func Test_ServerConnection_Cd_NotFoundError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Cd(remotePath)
	require.EqualError(t, err, fmt.Sprintf("not found error occurred: path %s does not exist", remotePath))
	assert.IsType(t, ftperrors.NotFoundErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_ServerConnection_Cd_CmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Cd(remotePath)
	require.EqualError(t, err, "an internal error occurred: failed to change working directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Cd_InvalidStatus(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusCommandNotImplemented, "mock error", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Cd(remotePath)
	require.EqualError(t, err, "an internal error occurred: mock error")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}
