package ftpconnection_test

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

func Test_ServerConnection_Ready_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.On("ReadResponse", ftpconnection.StatusReady).Return(0, "", nil)

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Ready()
	assert.NoError(t, err)
}

func Test_ServerConnection_Ready_ReadResponseError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.On("ReadResponse", ftpconnection.StatusReady).Return(0, "", errors.New("mock error"))
	connMock.On("Cmd", ftpconnection.CommandQuit).Return(uid, nil)
	connMock.On("Close").Return(nil)

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Ready()
	require.EqualError(t, err, "an internal error occurred: failed to check if server is ready")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Ready_StopError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.On("ReadResponse", ftpconnection.StatusReady).Return(0, "", errors.New("mock error"))
	connMock.On("Cmd", ftpconnection.CommandQuit).Return(uid, nil)
	connMock.On("Close").Return(errors.New("mock error"))

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Ready()
	require.EqualError(t, err, "an internal error occurred: failed to close connection")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_EnableExplicitTLSMode_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.On("Cmd", ftpconnection.CommandAuthTLS).Return(uid, nil)
	connMock.On("ReadResponse", ftpconnection.StatusAuthOK).Return(0, "", nil)

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.EnableExplicitTLSMode()
	assert.NoError(t, err)
}

func Test_ServerConnection_EnableExplicitTLSMode_CmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

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

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.EnableExplicitTLSMode()
	require.EqualError(t, err, "an internal error occurred: failed to enable explicit TLS mode")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_EnableExplicitTLSMode_StopError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

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

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.EnableExplicitTLSMode()
	require.EqualError(t, err, "an internal error occurred: failed to close connection")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Stop_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandQuit).
		Return(uid, nil).
		Once()
	connMock.
		On("Close").
		Return(nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Stop()
	assert.NoError(t, err)
}

func Test_ServerConnection_Stop_CmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandQuit).
		Return(uid, errors.New("mock error")).
		Once()
	connMock.
		On("Close").
		Return(nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Stop()
	require.EqualError(t, err, "an internal error occurred: failed to disconnect from the server")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Stop_CloseError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandQuit).
		Return(uid, errors.New("mock error")).
		Once()
	connMock.
		On("Close").
		Return(errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Stop()
	require.EqualError(t, err, "an internal error occurred: failed to close connection")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_Success(t *testing.T) {
	testCases := []struct {
		name          string
		utfStatusCode int
	}{
		{
			name:          "status command ok",
			utfStatusCode: ftpconnection.StatusCommandOK,
		},
		{
			name:          "status bad arguments",
			utfStatusCode: ftpconnection.StatusBadArguments,
		},
		{
			name:          "status not implemented parameter",
			utfStatusCode: ftpconnection.StatusNotImplementedParameter,
		},
		{
			name:          "status command not implemented",
			utfStatusCode: ftpconnection.StatusCommandNotImplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dialOptions := ftpconnection.NewDialOptions()

			tcpConn := &net.TCPConn{}

			connMock := ftpConnectionMocks.NewTextConnection(t)
			connMock.
				On("Cmd", ftpconnection.CommandUser, user).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", ftpconnection.StatusNoCheck).
				Return(ftpconnection.StatusUserOK, "", nil).
				Once()
			connMock.
				On("Cmd", ftpconnection.CommandPass, password).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", ftpconnection.StatusLoggedIn).
				Return(ftpconnection.StatusLoggedIn, "", nil).
				Once()
			connMock.
				On("Cmd", ftpconnection.CommandFeat).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", ftpconnection.StatusNoCheck).
				Return(ftpconnection.StatusSystem, featureMsg, nil).
				Once()
			connMock.
				On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", ftpconnection.StatusCommandOK).
				Return(ftpconnection.StatusCommandOK, "", nil).
				Once()
			connMock.
				On("Cmd", ftpconnection.CommandOptions, ftpconnection.FeatureUTF8, ftpconnection.On).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", ftpconnection.StatusNoCheck).
				Return(tc.utfStatusCode, "", nil).
				Once()

			serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
			require.NoError(t, err)

			err = serverConn.Login(user, password)
			assert.NoError(t, err)
		})
	}
}

func Test_ServerConnection_Login_WithTLSConfig_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()
	err := ftpconnection.WithTLSConfig(&tls.Config{
		MinVersion: tls.VersionTLS13,
	})(dialOptions)
	require.NoError(t, err)

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusCommandOK).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Times(3)
	connMock.
		On("Cmd", ftpconnection.CommandOptions, ftpconnection.FeatureUTF8, ftpconnection.On).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandProtectionBufferSize).
		Return(uid, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandProtocol).
		Return(uid, nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_AlreadyLoggerInUser_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsgWithoutMLST, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusCommandOK).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandOptions, ftpconnection.FeatureUTF8, ftpconnection.On).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_FeatureCmdNotSupported_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusBadCommand, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_NoUTF8Feature_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsgWithoutUTF8, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusCommandOK).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_UTF8Disabled_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()
	err := ftpconnection.WithDisabledUTF8()(dialOptions)
	require.NoError(t, err)

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusCommandOK).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_UserCmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to start username authentication")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_InvalidStatusFromUserCmd(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusBadArguments, "mock error", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: mock error")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_ServerConnection_Login_PasswordError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to authenticate user")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_FeatureCmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to list supported features")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_TypeCmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to set binary transfer mode")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_UTF8CmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusCommandOK).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandOptions, ftpconnection.FeatureUTF8, ftpconnection.On).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to turn UTF-8 option on")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_InvalidStatusFromUTF8Cmd(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusCommandOK).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandOptions, ftpconnection.FeatureUTF8, ftpconnection.On).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusBadCommand, "mock error", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to turn UTF-8 option on")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_BufferSizeCmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()
	err := ftpconnection.WithTLSConfig(&tls.Config{
		MinVersion: tls.VersionTLS13,
	})(dialOptions)
	require.NoError(t, err)

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusCommandOK).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandOptions, ftpconnection.FeatureUTF8, ftpconnection.On).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandProtectionBufferSize).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to set protocol buffer size")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_ProtocolCmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()
	err := ftpconnection.WithTLSConfig(&tls.Config{
		MinVersion: tls.VersionTLS13,
	})(dialOptions)
	require.NoError(t, err)

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusLoggedIn).
		Return(ftpconnection.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandType, ftpconnection.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusCommandOK).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Times(2)
	connMock.
		On("Cmd", ftpconnection.CommandOptions, ftpconnection.FeatureUTF8, ftpconnection.On).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandProtectionBufferSize).
		Return(uid, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandProtocol).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to enable TLS protocol")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Cd_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusRequestedFileActionOK, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Cd(remotePath)
	assert.NoError(t, err)
}

func Test_ServerConnection_Cd_NotFoundError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusFileUnavailable, "", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Cd(remotePath)
	require.EqualError(t, err, fmt.Sprintf("not found error: path %s does not exist", remotePath))
	assert.IsType(t, ftperrors.NotFoundErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_ServerConnection_Cd_CmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Cd(remotePath)
	require.EqualError(t, err, "an internal error occurred: failed to change working directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Cd_InvalidStatus(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusCommandNotImplemented, "mock error", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Cd(remotePath)
	require.EqualError(t, err, "an internal error occurred: mock error")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_ServerConnection_Size_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandSize, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusFile).
		Return(ftpconnection.StatusFile, "1024", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	sizeInBytes, err := serverConn.Size(remotePath)
	assert.Equal(t, uint64(1024), sizeInBytes)
	assert.NoError(t, err)
}

func Test_ServerConnection_Size_CmdError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandSize, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	sizeInBytes, err := serverConn.Size(remotePath)
	assert.Equal(t, uint64(0), sizeInBytes)
	require.EqualError(t, err, "an internal error occurred: failed to fetch file size")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Size_ParseError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandSize, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusFile).
		Return(ftpconnection.StatusFile, "no-a-number", nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	sizeInBytes, err := serverConn.Size(remotePath)
	assert.Equal(t, uint64(0), sizeInBytes)
	require.EqualError(t, err, "an internal error occurred: failed to parse file size to a non-zero integer")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), `strconv.ParseUint: parsing "no-a-number": invalid syntax`)
}

func Test_ServerConnection_Status_Success(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandStatus).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusSystem).
		Return(ftpconnection.StatusSystem, statusMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandSystem).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusName).
		Return(ftpconnection.StatusName, systemMsg, nil).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
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
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandStatus).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	status, err := serverConn.Status()
	assert.Nil(t, status)
	require.EqualError(t, err, "an internal error occurred: failed to fetch server status")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Status_SystemCmdErr(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandStatus).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusSystem).
		Return(ftpconnection.StatusSystem, statusMsg, nil).
		Once()
	connMock.
		On("Cmd", ftpconnection.CommandSystem).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	status, err := serverConn.Status()
	assert.Nil(t, status)
	require.EqualError(t, err, "an internal error occurred: failed to fetch system type")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Mkdir_InvalidArgumentError(t *testing.T) {
	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

	connMock := ftpConnectionMocks.NewTextConnection(t)
	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir("")
	require.EqualError(t, err, "an invalid argument error: argument path cannot be blank")
	assert.IsType(t, ftperrors.InvalidArgumentErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_ServerConnection_Mkdir_1dPath_Success(t *testing.T) {
	const path = "/foo"

	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

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

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	assert.NoError(t, err)
}

func Test_ServerConnection_Mkdir_2dPath_Success(t *testing.T) {
	const path = "foo/bar"

	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

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

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	assert.NoError(t, err)
}

func Test_ServerConnection_Mkdir_CheckDirExistsError(t *testing.T) {
	const path = "/foo"

	tcpConn := &net.TCPConn{}

	dialOptions := ftpconnection.NewDialOptions()

	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandChangeWorkDir, path).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusBadCommand, "mock error", errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	require.EqualError(t, err, "an internal error occurred: failed to change working directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Mkdir_MakeDirError(t *testing.T) {
	const path = "/foo"

	tcpConn := &net.TCPConn{}

	dialOptions := ftpconnection.NewDialOptions()

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

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	require.EqualError(t, err, "an internal error occurred: failed to create directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Mkdir_CdError(t *testing.T) {
	const path = "/foo"

	dialOptions := ftpconnection.NewDialOptions()

	tcpConn := &net.TCPConn{}

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

	serverConn, err := ftpconnection.NewServerConnection(host, tcpConn, connMock, dialOptions)
	require.NoError(t, err)

	err = serverConn.Mkdir(path)
	require.EqualError(t, err, "an internal error occurred: failed to change working directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
