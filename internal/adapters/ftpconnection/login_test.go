package ftpconnection_test

import (
	"crypto/tls"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
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
			utfStatusCode: models.StatusCommandOK,
		},
		{
			name:          "status bad arguments",
			utfStatusCode: models.StatusBadArguments,
		},
		{
			name:          "status not implemented parameter",
			utfStatusCode: models.StatusNotImplementedParameter,
		},
		{
			name:          "status command not implemented",
			utfStatusCode: models.StatusCommandNotImplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tcpConn := ftpConnectionMocks.NewConn(t)
			dialer := ftpConnectionMocks.NewDialer(t)
			connMock := ftpConnectionMocks.NewTextConnection(t)
			connMock.
				On("Cmd", models.CommandUser, user).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", models.StatusNoCheck).
				Return(models.StatusUserOK, "", nil).
				Once()
			connMock.
				On("Cmd", models.CommandPass, password).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", models.StatusLoggedIn).
				Return(models.StatusLoggedIn, "", nil).
				Once()
			connMock.
				On("Cmd", models.CommandFeat).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", models.StatusNoCheck).
				Return(models.StatusSystem, featureMsg, nil).
				Once()
			connMock.
				On("Cmd", models.CommandType, models.TransferTypeBinary).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", models.StatusCommandOK).
				Return(models.StatusCommandOK, "", nil).
				Once()
			connMock.
				On("Cmd", models.CommandOptions, models.FeatureUTF8, "ON").
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", models.StatusNoCheck).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Times(3)
	connMock.
		On("Cmd", models.CommandOptions, models.FeatureUTF8, "ON").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandProtectionBufferSize).
		Return(uid, nil).
		Once()
	connMock.
		On("Cmd", models.CommandProtocol).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsgWithoutMLST, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandOptions, models.FeatureUTF8, "ON").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusCommandOK, "", nil).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusBadCommand, "", nil).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsgWithoutUTF8, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
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

// nolint:dupl // similar to Test_ServerConnection_Move_PrepareCmdError
func Test_ServerConnection_Login_UserCmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandUser, user).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusBadArguments, "mock error", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	err = serverConn.Login(user, password)
	require.EqualError(t, err, "an internal error occurred: mock error")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

// nolint:dupl // similar to Test_ServerConnection_Move_MoveCmdError
func Test_ServerConnection_Login_PasswordError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandOptions, models.FeatureUTF8, "ON").
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandOptions, models.FeatureUTF8, "ON").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusBadCommand, "mock error", nil).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandOptions, models.FeatureUTF8, "ON").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandProtectionBufferSize).
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
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusUserOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPass, password).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusLoggedIn).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsg, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Times(2)
	connMock.
		On("Cmd", models.CommandOptions, models.FeatureUTF8, "ON").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandProtectionBufferSize).
		Return(uid, nil).
		Once()
	connMock.
		On("Cmd", models.CommandProtocol).
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
