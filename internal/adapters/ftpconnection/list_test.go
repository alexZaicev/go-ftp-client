package ftpconnection_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

//nolint:funlen // test case can get a bit large
func Test_ServerConnection_List_Success(t *testing.T) {
	testCases := []struct {
		name    string
		options *connection.ListOptions
		command string
	}{
		{
			name: "list files",
			options: &connection.ListOptions{
				Path: remotePath,
			},
			command: models.CommandList,
		},
		{
			name: "list all files",
			options: &connection.ListOptions{
				Path:    remotePath,
				ShowAll: true,
			},
			command: models.CommandListHidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			expectedEntry := &entities.Entry{
				Type:                 entities.EntryTypeFile,
				Permissions:          "rw-r--r--",
				OwnerGroup:           "ftp",
				OwnerUser:            "ftp",
				Name:                 "file-1.txt",
				NumHardLinks:         1,
				SizeInBytes:          187,
				LastModificationDate: time.Date(0, 9, 16, 14, 34, 0, 0, time.UTC),
			}

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
					copy(bytes, entryFileMessage)
				}).
				Return(len(entryFileMessage), nil).
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
				On("Cmd", fmt.Sprintf(models.CommandPreTransfer, tc.command), remotePath).
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
				On("Cmd", tc.command, remotePath).
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

			entries, err := serverConn.List(ctx, tc.options)
			assert.NoError(t, err)
			if assert.Len(t, entries, 1) {
				assert.Equal(t, expectedEntry, entries[0])
			}
		})
	}
}

//nolint:funlen // test case can get a bit large
func Test_ServerConnection_List_WithTLS_Success(t *testing.T) {
	ctx := context.Background()

	expectedEntry := &entities.Entry{
		Type:                 entities.EntryTypeFile,
		Permissions:          "rw-r--r--",
		OwnerGroup:           "ftp",
		OwnerUser:            "ftp",
		Name:                 "file-1.txt",
		NumHardLinks:         1,
		SizeInBytes:          187,
		LastModificationDate: time.Date(0, 9, 16, 14, 34, 0, 0, time.UTC),
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

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
			copy(bytes, entryFileMessage)
		}).
		Return(len(entryFileMessage), nil).
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
		On("DialContextTLS", ctx, "tcp", fmt.Sprintf("%s:21103", host), tlsConfig).
		Return(dataConnMock, nil).
		Once()

	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, true)
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
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

	serverConn, err := ftpconnection.NewConnection(
		host,
		dialer,
		tcpConn,
		connMock,
		ftpconnection.WithTLSConfig(tlsConfig),
	)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.NoError(t, err)
	if assert.Len(t, entries, 1) {
		assert.Equal(t, expectedEntry, entries[0])
	}
}

func Test_ServerConnection_List_WithDisabledEPSV_Success(t *testing.T) {
	ctx := context.Background()

	expectedEntry := &entities.Entry{
		Type:                 entities.EntryTypeFile,
		Permissions:          "rw-r--r--",
		OwnerGroup:           "ftp",
		OwnerUser:            "ftp",
		Name:                 "file-1.txt",
		NumHardLinks:         1,
		SizeInBytes:          187,
		LastModificationDate: time.Date(0, 9, 16, 14, 34, 0, 0, time.UTC),
	}

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
			copy(bytes, entryFileMessage)
		}).
		Return(len(entryFileMessage), nil).
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
		On("DialContext", ctx, "tcp", "10.0.0.1:21103").
		Return(dataConnMock, nil).
		Once()

	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPassive).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusPassiveMode).
		Return(models.StatusPassiveMode, passiveModeMessage, nil).
		Once()
	connMock.
		On("Cmd", models.CommandList, remotePath).
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

	serverConn, err := ftpconnection.NewConnection(
		host,
		dialer,
		tcpConn,
		connMock,
		ftpconnection.WithDisabledEPSV(),
	)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.NoError(t, err)
	if assert.Len(t, entries, 1) {
		assert.Equal(t, expectedEntry, entries[0])
	}
}

func Test_ServerConnection_List_InvalidArgumentError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)

	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	entries, err := serverConn.List(ctx, nil)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an invalid argument error occurred: argument options cannot be nil")
	assert.IsType(t, ftperrors.InvalidArgumentErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_ServerConnection_List_PreTransferCmdError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)

	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "an internal error occurred: failed to issue pre-transfer configuration")
}

func Test_ServerConnection_List_ExtendedPassiveModeError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)

	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandExtendedPassiveMode).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "an internal error occurred: failed to set extended passive mode")
}

func Test_ServerConnection_List_ExtendedPassiveModeMessageError(t *testing.T) {
	testCases := []struct {
		name          string
		message       string
		wrappedErrMsg string
	}{
		{
			name:          "invalid message format",
			message:       "not a valid message",
			wrappedErrMsg: "an internal error occurred: invalid format of extended passive mode message",
		},
		{
			name:          "no port definition",
			message:       "Entering Extended Passive Mode",
			wrappedErrMsg: "an internal error occurred: failed to extract port from extended passive mode message",
		},
		{
			name:          "port not an integer",
			message:       "Entering Extended Passive Mode (|||this is my port|)",
			wrappedErrMsg: "an internal error occurred: failed to parse connection port",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			tcpConn := ftpConnectionMocks.NewConn(t)

			dialer := ftpConnectionMocks.NewDialer(t)
			connMock := ftpConnectionMocks.NewTextConnection(t)
			// mock setup for login
			setMocksForLogin(connMock, false)
			// mock setup for list
			connMock.
				On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
				Return(models.StatusExtendedPassiveMode, tc.message, nil).
				Once()

			serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
			require.NoError(t, err)

			// this is required to feed the feature map
			err = serverConn.Login(user, password)
			require.NoError(t, err)

			options := &connection.ListOptions{
				Path: remotePath,
			}

			entries, err := serverConn.List(ctx, options)
			assert.Nil(t, entries)
			require.EqualError(t, err, "an internal error occurred: failed to list files")
			assert.IsType(t, ftperrors.InternalErrorType, err)
			assert.EqualError(t, errors.Unwrap(err), tc.wrappedErrMsg)
		})
	}
}

func Test_ServerConnection_List_WithDisabledEPSV_PassiveModeCmdError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)

	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	// mock setup for login
	setMocksForLogin(connMock, false)
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandPassive).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(
		host,
		dialer,
		tcpConn,
		connMock,
		ftpconnection.WithDisabledEPSV(),
	)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "an internal error occurred: failed to set passive mode")
}

func Test_ServerConnection_List_WithDisabledEPSV_InvalidPassiveModeMessageError(t *testing.T) {
	testCases := []struct {
		name          string
		message       string
		wrappedErrMsg string
	}{
		{
			name:          "invalid message format",
			message:       "not a valid message",
			wrappedErrMsg: "an internal error occurred: invalid format of passive mode message",
		},
		{
			name:          "failed to extract host and port",
			message:       "Entering Passive Mode",
			wrappedErrMsg: "an internal error occurred: failed to extract host and port from passive mode message",
		},
		{
			name:          "first part of port not an integer",
			message:       "Entering Passive Mode (10,0,0,1,not-an-int,111)",
			wrappedErrMsg: "an internal error occurred: failed to parse first part of connection port",
		},
		{
			name:          "second part of port not an integer",
			message:       "Entering Passive Mode (10,0,0,1,82,not-an-int)",
			wrappedErrMsg: "an internal error occurred: failed to parse second part of connection port",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			tcpConn := ftpConnectionMocks.NewConn(t)

			dialer := ftpConnectionMocks.NewDialer(t)
			connMock := ftpConnectionMocks.NewTextConnection(t)
			// mock setup for login
			setMocksForLogin(connMock, false)
			// mock setup for list
			connMock.
				On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", models.StatusCommandOK).
				Return(models.StatusCommandOK, "", nil).
				Once()
			connMock.
				On("Cmd", models.CommandPassive).
				Return(uid, nil).
				Once()
			connMock.
				On("ReadResponse", models.StatusPassiveMode).
				Return(models.StatusPassiveMode, tc.message, nil).
				Once()

			serverConn, err := ftpconnection.NewConnection(
				host,
				dialer,
				tcpConn,
				connMock,
				ftpconnection.WithDisabledEPSV(),
			)
			require.NoError(t, err)

			// this is required to feed the feature map
			err = serverConn.Login(user, password)
			require.NoError(t, err)

			options := &connection.ListOptions{
				Path: remotePath,
			}

			entries, err := serverConn.List(ctx, options)
			assert.Nil(t, entries)
			require.EqualError(t, err, "an internal error occurred: failed to list files")
			assert.IsType(t, ftperrors.InternalErrorType, err)
			assert.EqualError(t, errors.Unwrap(err), tc.wrappedErrMsg)
		})
	}
}

func Test_ServerConnection_List_ListCmdError(t *testing.T) {
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
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_List_ListCmdFailedWithCloseError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Close").
		Return(errors.New("mock close error")).
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock close error")
}

func Test_ServerConnection_List_InvalidStatusFromListCmd(t *testing.T) {
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
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusBadArguments, "mock error", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "an internal error occurred: mock error")
}

func Test_ServerConnection_List_InvalidStatusFromListCmdCloseError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Close").
		Return(errors.New("mock close error")).
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusBadArguments, "mock error", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock close error")
}

func Test_ServerConnection_List_ParserError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	const invalidEntryFileMessage = "not-valid-entry"

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.Anything).
		Run(func(args mock.Arguments) {
			bytes := args.Get(0).([]byte)
			copy(bytes, invalidEntryFileMessage)
		}).
		Return(len(invalidEntryFileMessage), nil).
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
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

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* an internal error occurred: unsupported entry format\n\n")
}

func Test_ServerConnection_List_ScannerError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.Anything).
		Return(0, bufio.ErrBadReadCount).
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
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

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* bufio.Scanner: Read returned impossible count\n\n")
}

func Test_ServerConnection_List_ConnCloseError(t *testing.T) {
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
			copy(bytes, entryFileMessage)
		}).
		Return(len(entryFileMessage), nil).
		Once()
	dataConnMock.
		On("Read", mock.Anything).
		Return(0, io.EOF).
		Once()
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
	// mock setup for list
	connMock.
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
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

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* mock error\n\n")
}

func Test_ServerConnection_List_SetDeadlineError(t *testing.T) {
	ctx := context.Background()

	tcpConn := ftpConnectionMocks.NewConn(t)
	tcpConn.
		On("SetDeadline", mock.AnythingOfType("time.Time")).
		Return(errors.New("mock error")).
		Once()

	dataConnMock := ftpConnectionMocks.NewConn(t)
	dataConnMock.
		On("Read", mock.Anything).
		Run(func(args mock.Arguments) {
			bytes := args.Get(0).([]byte)
			copy(bytes, entryFileMessage)
		}).
		Return(len(entryFileMessage), nil).
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusAboutToSend, listMessage, nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* mock error\n\n")
}

func Test_ServerConnection_List_CheckConnShutError(t *testing.T) {
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
			copy(bytes, entryFileMessage)
		}).
		Return(len(entryFileMessage), nil).
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
		On("Cmd", fmt.Sprintf(models.CommandPreTransfer, models.CommandList), remotePath).
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
		On("Cmd", models.CommandList, remotePath).
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

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	// this is required to feed the feature map
	err = serverConn.Login(user, password)
	require.NoError(t, err)

	options := &connection.ListOptions{
		Path: remotePath,
	}

	entries, err := serverConn.List(ctx, options)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "1 error occurred:\n\t* mock error\n\n")
}
