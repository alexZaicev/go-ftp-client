package ftpconnection_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

func Test_ServerConnection_Status_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandStatus).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusSystem).
		Return(models.StatusSystem, statusMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandSystem).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusName).
		Return(models.StatusName, systemMsg, nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	status, err := serverConn.Status()
	assert.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "172.22.0.2", status.RemoteAddress)
	assert.Equal(t, "ftpuser01", status.LoggedInUser)
	assert.Equal(t, "UNIX", status.System)
	assert.False(t, status.TLSEnabled)
}

func Test_ServerConnection_Status_StatusCmdErr(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandStatus).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	status, err := serverConn.Status()
	assert.Nil(t, status)
	require.EqualError(t, err, "an internal error occurred: failed to fetch server status")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Status_SystemCmdErr(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandStatus).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusSystem).
		Return(models.StatusSystem, statusMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandSystem).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	status, err := serverConn.Status()
	assert.Nil(t, status)
	require.EqualError(t, err, "an internal error occurred: failed to fetch system type")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
