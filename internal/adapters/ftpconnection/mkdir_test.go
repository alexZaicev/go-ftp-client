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

func Test_ServerConnection_Mkdir_InvalidArgumentError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Mkdir("")
	require.EqualError(t, err, "an invalid argument error occurred: argument path cannot be blank")
	assert.IsType(t, ftperrors.InvalidArgumentErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_ServerConnection_Mkdir_1dPath_Success(t *testing.T) {
	const path = "/foo"

	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, path).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandMakeDir, path).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "/").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusRequestedFileActionOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	assert.NoError(t, err)
}

func Test_ServerConnection_Mkdir_2dPath_Success(t *testing.T) {
	const path = "foo/bar"

	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "/foo").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Twice()
	connMock.
		On("Cmd", ftpconnection.CommandMakeDir, "/foo").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Twice()
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "/foo/bar").
		Return(uid, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandMakeDir, "/foo/bar").
		Return(uid, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "/").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusRequestedFileActionOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	assert.NoError(t, err)
}

func Test_ServerConnection_Mkdir_CheckDirExistsError(t *testing.T) {
	const path = "/foo"

	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, path).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusBadCommand, "mock error", errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	require.EqualError(t, err, "an internal error occurred: failed to change working directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Mkdir_MakeDirError(t *testing.T) {
	const path = "/foo"

	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, path).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandMakeDir, path).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusBadCommand, "mock error", errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	require.EqualError(t, err, "an internal error occurred: failed to create directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Mkdir_CdError(t *testing.T) {
	const path = "/foo"

	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, path).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandMakeDir, path).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "/").
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	require.EqualError(t, err, "an internal error occurred: failed to change working directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
