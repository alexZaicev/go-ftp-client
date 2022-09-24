package ftpconnection_test

import (
	"crypto/tls"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

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
			tcpConn := ftpConnectionMocks.NewConn(t)
			dialer := ftpConnectionMocks.NewDialer(t)
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

			serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
			require.NoError(t, err)

			err = serverConn.Login(user, password)
			assert.NoError(t, err)
		})
	}
}

func Test_ServerConnection_Login_WithTLSConfig_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(
		host,
		dialer,
		tcpConn,
		connMock,
		ftpconnection.WithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS13,
		}),
	)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_AlreadyLoggerInUser_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_FeatureCmdNotSupported_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_NoUTF8Feature_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_UTF8Disabled_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(
		host,
		dialer,
		tcpConn,
		connMock,
		ftpconnection.WithDisabledUTF8(),
	)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	assert.NoError(t, err)
}

func Test_ServerConnection_Login_UserCmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to start username authentication")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_InvalidStatusFromUserCmd(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", ftpconnection.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", ftpconnection.StatusNoCheck).
		Return(ftpconnection.StatusBadArguments, "mock error", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: mock error")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_ServerConnection_Login_PasswordError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to authenticate user")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_FeatureCmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to list supported features")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_TypeCmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to set binary transfer mode")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_UTF8CmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to turn UTF-8 option on")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_InvalidStatusFromUTF8Cmd(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to turn UTF-8 option on")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_BufferSizeCmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(
		host,
		dialer,
		tcpConn,
		connMock,
		ftpconnection.WithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS13,
		}),
	)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to set protocol buffer size")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Login_ProtocolCmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
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

	serverConn, err := ftpconnection.NewConnection(
		host,
		dialer,
		tcpConn,
		connMock,
		ftpconnection.WithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS13,
		}),
	)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: failed to enable TLS protocol")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}
