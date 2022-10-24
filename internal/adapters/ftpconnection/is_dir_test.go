package ftpconnection_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

func Test_ServerConnection_IsDir_Success(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.Anything).
		Run(func(args mock.Arguments) {
			bytes := args.Get(0).([]byte)
			copy(bytes, entryDirMessage)
		}).
		Return(len(entryDirMessage), nil).
		Once()
	dataConnMock.
		On("Read", mock.Anything).
		Return(0, io.EOF).
		Once()
	dataConnMock.
		On("Close").
		Return(nil).
		Once()

	dialer := ftpConnectionMocks.NewDialer(t)
	dialer.
		On("DialContext", ctx, "tcp", fmt.Sprintf("%s:21103", host)).
		Return(dataConnMock, nil).
		Once()

	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandListHidden), remoteParentPath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandExtendedPassiveMode).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusExtendedPassiveMode).
		Return(models.StatusExtendedPassiveMode, extendedPassiveModeMessage, nil).
		Once()
	connMock.
		On("Cmd", models.CommandListHidden, remoteParentPath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusAboutToSend, listMessage, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusClosingDataConnection).
		Return(models.StatusClosingDataConnection, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	// act
	isDir, err := serverConn.IsDir(ctx, remotePath)

	// assert
	assert.NoError(t, err)
	assert.True(t, isDir)
}

func Test_ServerConnection_IsDir_ListError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)

	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandListHidden), remoteParentPath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	// act
	isDir, err := serverConn.IsDir(ctx, remotePath)

	// assert
	assert.False(t, isDir)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "an internal error occurred: failed to issue pre-transfer configuration")
}

func Test_ServerConnection_IsDir_NotFoundError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.Anything).
		Run(func(args mock.Arguments) {
			bytes := args.Get(0).([]byte)
			copy(bytes, "")
		}).
		Return(0, nil).
		Once()
	dataConnMock.
		On("Read", mock.Anything).
		Return(0, io.EOF).
		Once()
	dataConnMock.
		On("Close").
		Return(nil).
		Once()

	dialer := ftpConnectionMocks.NewDialer(t)
	dialer.
		On("DialContext", ctx, "tcp", fmt.Sprintf("%s:21103", host)).
		Return(dataConnMock, nil).
		Once()

	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandListHidden), remoteParentPath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandExtendedPassiveMode).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusExtendedPassiveMode).
		Return(models.StatusExtendedPassiveMode, extendedPassiveModeMessage, nil).
		Once()
	connMock.
		On("Cmd", models.CommandListHidden, remoteParentPath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusAboutToSend, listMessage, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusClosingDataConnection).
		Return(models.StatusClosingDataConnection, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	// act
	isDir, err := serverConn.IsDir(ctx, remotePath)

	// assert
	assert.False(t, isDir)
	require.EqualError(t, err, fmt.Sprintf(`not found error occurred: entry not found under %q path`, remotePath))
	assert.IsType(t, ftperrors.NotFoundErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}
