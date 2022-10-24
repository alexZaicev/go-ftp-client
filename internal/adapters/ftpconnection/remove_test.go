package ftpconnection_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftpErrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

// nolint:dupl // similar to Test_ServerConnection_RemoveDir_Success
func Test_ServerConnection_RemoveFile_Success(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandRemoveFile, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusRequestedFileActionOK).
		Return(models.StatusRequestedFileActionOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.RemoveFile(remotePath)

	// assert
	assert.NoError(t, err)
}

func Test_ServerConnection_RemoveFile_InvalidArgument(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.RemoveFile("")

	// assert
	require.EqualError(t, err, "an invalid argument error occurred: argument path cannot be blank")
	assert.IsType(t, ftpErrors.InvalidArgumentErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

// nolint:dupl // similar to Test_ServerConnection_Cd_CmdError
func Test_ServerConnection_RemoveFile_CmdError(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandRemoveFile, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.RemoveFile(remotePath)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to remove file")
	assert.IsType(t, ftpErrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

// nolint:dupl // similar to Test_ServerConnection_RemoveFile_Success
func Test_ServerConnection_RemoveDir_Success(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandRemoveDir, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusRequestedFileActionOK).
		Return(models.StatusRequestedFileActionOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.RemoveDir(remotePath)

	// assert
	assert.NoError(t, err)
}

func Test_ServerConnection_RemoveDir_InvalidArgument(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.RemoveDir("")

	// assert
	require.EqualError(t, err, "an invalid argument error occurred: argument path cannot be blank")
	assert.IsType(t, ftpErrors.InvalidArgumentErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

// nolint:dupl // similar to Test_ServerConnection_Cd_CmdError
func Test_ServerConnection_RemoveDir_CmdError(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandRemoveDir, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.RemoveDir(remotePath)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to remove directory")
	assert.IsType(t, ftpErrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
