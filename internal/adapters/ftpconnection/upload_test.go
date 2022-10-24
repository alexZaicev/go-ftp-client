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
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

func Test_ServerConnection_Upload_Success(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Write", mock.AnythingOfType("[]uint8")).
		Return(buffer.Len(), nil)
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandStore), remotePath).
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
		On("Cmd", models.CommandStore, remotePath).
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

	options := &connection.UploadOptions{
		Path:       remotePath,
		FileReader: buffer,
	}

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	err = serverConn.Upload(ctx, options)
	assert.NoError(t, err)
}

func Test_ServerConnection_Upload_InvalidArgumentError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)

	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	err = serverConn.Upload(ctx, nil)
	require.EqualError(t, err, "an invalid argument error occurred: argument options cannot be nil")
	assert.IsType(t, ftperrors.InvalidArgumentErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_ServerConnection_Upload_CmdError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandStore), remotePath).
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
		On("Cmd", models.CommandStore, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	options := &connection.UploadOptions{
		Path:       remotePath,
		FileReader: buffer,
	}

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	err = serverConn.Upload(ctx, options)
	require.EqualError(t, err, "an internal error occurred: failed to open data transfer connection")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Upload_CopyError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Write", mock.AnythingOfType("[]uint8")).
		Return(0, io.ErrShortWrite)
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandStore), remotePath).
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
		On("Cmd", models.CommandStore, remotePath).
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

	options := &connection.UploadOptions{
		Path:       remotePath,
		FileReader: buffer,
	}

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	err = serverConn.Upload(ctx, options)
	require.EqualError(t, err, "an internal error occurred: failed to upload file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* short write\n\n")
}

func Test_ServerConnection_Upload_ConnCloseError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Write", mock.AnythingOfType("[]uint8")).
		Return(buffer.Len(), nil)
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandStore), remotePath).
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
		On("Cmd", models.CommandStore, remotePath).
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

	options := &connection.UploadOptions{
		Path:       remotePath,
		FileReader: buffer,
	}

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	err = serverConn.Upload(ctx, options)
	require.EqualError(t, err, "an internal error occurred: failed to upload file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* mock error\n\n")
}

func Test_ServerConnection_Upload_CheckConnShutError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Write", mock.AnythingOfType("[]uint8")).
		Return(buffer.Len(), nil)
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandStore), remotePath).
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
		On("Cmd", models.CommandStore, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusAboutToSend, listMessage, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusClosingDataConnection).
		Return(models.StatusBadCommand, "", errors.New("mock error")).
		Once()

	options := &connection.UploadOptions{
		Path:       remotePath,
		FileReader: buffer,
	}

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	err = serverConn.Upload(ctx, options)
	require.EqualError(t, err, "an internal error occurred: failed to upload file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* mock error\n\n")
}
