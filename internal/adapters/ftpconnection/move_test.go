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

func Test_ServerConnection_Move_Success(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandRenameFrom, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusRequestFilePending).
		Return(models.StatusRequestFilePending, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandRenameTo, newRemotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusRequestedFileActionOK).
		Return(models.StatusRequestedFileActionOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.Move(remotePath, newRemotePath)

	// assert
	assert.NoError(t, err)
}

func Test_ServerConnection_Move_InvalidArgument(t *testing.T) {
	testCases := []struct {
		name           string
		oldPath        string
		newPath        string
		expectedErrMsg string
	}{
		{
			name:           "oldPath blank",
			newPath:        newRemotePath,
			expectedErrMsg: "an invalid argument error occurred: argument oldPath cannot be blank",
		},
		{
			name:           "newPath blank",
			oldPath:        remotePath,
			expectedErrMsg: "an invalid argument error occurred: argument newPath cannot be blank",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			tcpConn := ftpConnectionMocks.NewConn(t)
			dialer := ftpConnectionMocks.NewDialer(t)
			connMock := ftpConnectionMocks.NewTextConnection(t)

			serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
			require.NoError(t, err)

			// act
			err = serverConn.Move(tc.oldPath, tc.newPath)

			// assert
			require.EqualError(t, err, tc.expectedErrMsg)
			assert.IsType(t, ftpErrors.InvalidArgumentErrorType, err)
			assert.NoError(t, errors.Unwrap(err))
		})
	}
}

// nolint:dupl // similar to Test_ServerConnection_Login_UserCmdError
func Test_ServerConnection_Move_PrepareCmdError(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandRenameFrom, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.Move(remotePath, newRemotePath)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to prepare file")
	assert.IsType(t, ftpErrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

// nolint:dupl // similar to Test_ServerConnection_Login_PasswordError
func Test_ServerConnection_Move_MoveCmdError(t *testing.T) {
	// arrange
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandRenameFrom, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusRequestFilePending).
		Return(models.StatusRequestFilePending, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandRenameTo, newRemotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	err = serverConn.Move(remotePath, newRemotePath)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to move file")
	assert.IsType(t, ftpErrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
