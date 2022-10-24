package ftpconnection_test

import (
	"bytes"
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

func Test_ServerConnection_Download_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.AnythingOfType("[]uint8")).
		Return(buffer.Len(), io.EOF)
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
	// mock setup for upload
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandRetrieve), remotePath).
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
		On("Cmd", models.CommandRetrieve, remotePath).
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
	data, err := serverConn.Download(ctx, remotePath)

	// assert
	assert.NoError(t, err)
	assert.Len(t, data, buffer.Len())
}

func Test_ServerConnection_Download_InvalidArgumentError(t *testing.T) {
	// arrange
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// act
	data, err := serverConn.Download(ctx, "")

	// assert
	assert.Nil(t, data)

	require.EqualError(t, err, "an invalid argument error occurred: argument path cannot be blank")
	assert.IsType(t, ftperrors.InvalidArgumentErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_ServerConnection_Download_CmdError(t *testing.T) {
	// arrange
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	dataConnMock := ftpConnectionMocks.NewConn(t)
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
	// mock setup for upload
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandRetrieve), remotePath).
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
		On("Cmd", models.CommandRetrieve, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	// act
	data, err := serverConn.Download(ctx, remotePath)

	// assert
	assert.Nil(t, data)

	require.EqualError(t, err, "an internal error occurred: failed to open data transfer connection")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Download_ReadError(t *testing.T) {
	// arrange
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.AnythingOfType("[]uint8")).
		Return(buffer.Len(), errors.New("mock error"))
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
	// mock setup for upload
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandRetrieve), remotePath).
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
		On("Cmd", models.CommandRetrieve, remotePath).
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
	data, err := serverConn.Download(ctx, remotePath)

	// assert
	assert.Nil(t, data)

	require.EqualError(t, err, "an internal error occurred: failed to download file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* mock error\n\n")
}

func Test_ServerConnection_Download_CloseError(t *testing.T) {
	// arrange
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.AnythingOfType("[]uint8")).
		Return(buffer.Len(), io.EOF)
	dataConnMock.
		On("Close").
		Return(errors.New("mock error")).
		Once()

	dialer := ftpConnectionMocks.NewDialer(t)
	dialer.
		On("DialContext", ctx, "tcp", fmt.Sprintf("%s:21103", host)).
		Return(dataConnMock, nil).
		Once()

	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)
	// mock setup for upload
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandRetrieve), remotePath).
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
		On("Cmd", models.CommandRetrieve, remotePath).
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
	data, err := serverConn.Download(ctx, remotePath)

	// assert
	assert.Nil(t, data)

	require.EqualError(t, err, "an internal error occurred: failed to download file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* mock error\n\n")
}

func Test_ServerConnection_Download_CheckConnShutError(t *testing.T) {
	// arrange
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.AnythingOfType("[]uint8")).
		Return(buffer.Len(), io.EOF)
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
	// mock setup for upload
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandRetrieve), remotePath).
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
		On("Cmd", models.CommandRetrieve, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusAboutToSend, listMessage, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusClosingDataConnection).
		Return(models.StatusClosingDataConnection, "", errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	// act
	data, err := serverConn.Download(ctx, remotePath)

	// assert
	assert.Nil(t, data)

	require.EqualError(t, err, "an internal error occurred: failed to download file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* mock error\n\n")
}
