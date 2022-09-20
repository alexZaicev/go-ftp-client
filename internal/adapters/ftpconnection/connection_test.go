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
	dialOptions := ftpconnection.NewDialOptions()

	tcpConnMock := ftpConnectionMocks.NewTextConnection(t)
	serverConn, err := ftpconnection.NewServerConnection("", nil, tcpConnMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir("")
	require.EqualError(t, err, "an invalid argument error: argument path cannot be blank")
	assert.IsType(t, ftperrors.InvalidArgumentErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_ServerConnection_Mkdir_1dPath_Success(t *testing.T) {
	const path = "foo"
	uid := uint(0)

	dialOptions := ftpconnection.NewDialOptions()

	tcpConnMock := ftpConnectionMocks.NewTextConnection(t)
	tcpConnMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, path).
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandMakeDir, path).
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection("", nil, tcpConnMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	assert.NoError(t, err)
}

func Test_ServerConnection_Mkdir_2dPath_Success(t *testing.T) {
	const path = "foo/bar"
	uid := uint(0)

	dialOptions := ftpconnection.NewDialOptions()

	tcpConnMock := ftpConnectionMocks.NewTextConnection(t)
	tcpConnMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "foo").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandMakeDir, "foo").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "foo/bar").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandMakeDir, "foo/bar").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection("", nil, tcpConnMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	assert.NoError(t, err)
}

func Test_ServerConnection_Mkdir_3dPath_Success(t *testing.T) {
	const path = "foo/bar/baz"
	uid := uint(0)

	dialOptions := ftpconnection.NewDialOptions()

	tcpConnMock := ftpConnectionMocks.NewTextConnection(t)
	tcpConnMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "foo").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandMakeDir, "foo").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "foo/bar").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandMakeDir, "foo/bar").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, "foo/bar/baz").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandMakeDir, "foo/bar/baz").
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusPathCreated, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection("", nil, tcpConnMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	assert.NoError(t, err)
}

func Test_ServerConnection_Mkdir_CheckDirError(t *testing.T) {
	const path = "foo"
	uid := uint(0)

	dialOptions := ftpconnection.NewDialOptions()

	tcpConnMock := ftpConnectionMocks.NewTextConnection(t)
	tcpConnMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, path).
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusBadCommand, "mock error", errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection("", nil, tcpConnMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	require.EqualError(t, err, "an internal error occurred: failed to change working directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Mkdir_MakeDirError(t *testing.T) {
	const path = "foo"
	uid := uint(0)

	dialOptions := ftpconnection.NewDialOptions()

	tcpConnMock := ftpConnectionMocks.NewTextConnection(t)
	tcpConnMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, path).
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()
	tcpConnMock.
		On("Cmd", ftpconnection.CommandMakeDir, path).
		Return(uid, nil).
		Once()
	tcpConnMock.
		On("ReadResponse", ftpconnection.StatusPathCreated).
		Return(ftpconnection.StatusBadCommand, "mock error", errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection("", nil, tcpConnMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	require.EqualError(t, err, "an internal error occurred: failed to create directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
