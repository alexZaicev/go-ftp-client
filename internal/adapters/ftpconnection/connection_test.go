package ftpconnection_test

import (
	"bytes"
	"crypto/tls"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	ftpErrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

func Test_NewConnection_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)

	options := []ftpconnection.Option{
		ftpconnection.WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS13}),
		ftpconnection.WithVerboseWriter(bytes.NewBufferString("")),
		ftpconnection.WithDisabledEPSV(),
		ftpconnection.WithDisabledUTF8(),
	}

	serverConn, err := ftpconnection.NewConnection(
		host,
		dialer,
		tcpConn,
		connMock,
		options...,
	)
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
}

func Test_NewConnection_Errors(t *testing.T) {
	testCases := []struct {
		name           string
		host           string
		conn           ftpconnection.Conn
		textConn       ftpconnection.TextConnection
		dialer         ftpconnection.Dialer
		expectedErrMsg string
	}{
		{
			name:           "blank host",
			conn:           ftpConnectionMocks.NewConn(t),
			dialer:         ftpConnectionMocks.NewDialer(t),
			textConn:       ftpConnectionMocks.NewTextConnection(t),
			expectedErrMsg: "an invalid argument error occurred: argument host cannot be blank",
		},
		{
			name:           "nil dialer",
			host:           host,
			conn:           ftpConnectionMocks.NewConn(t),
			textConn:       ftpConnectionMocks.NewTextConnection(t),
			expectedErrMsg: "an invalid argument error occurred: argument dialer cannot be nil",
		},
		{
			name:           "nil conn",
			host:           host,
			dialer:         ftpConnectionMocks.NewDialer(t),
			textConn:       ftpConnectionMocks.NewTextConnection(t),
			expectedErrMsg: "an invalid argument error occurred: argument conn cannot be nil",
		},
		{
			name:           "nil textConn",
			host:           host,
			conn:           ftpConnectionMocks.NewConn(t),
			dialer:         ftpConnectionMocks.NewDialer(t),
			expectedErrMsg: "an invalid argument error occurred: argument textConn cannot be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serverConn, err := ftpconnection.NewConnection(
				tc.host,
				tc.dialer,
				tc.conn,
				tc.textConn,
			)
			assert.Nil(t, serverConn)
			require.EqualError(t, err, tc.expectedErrMsg)
			assert.IsType(t, ftpErrors.InvalidArgumentErrorType, err)
			assert.NoError(t, errors.Unwrap(err))
		})
	}
}

func Test_NewConnection_OptionErrors(t *testing.T) {
	testCases := []struct {
		name           string
		options        []ftpconnection.Option
		expectedErrMsg string
	}{
		{
			name: "invalid tls option",
			options: []ftpconnection.Option{
				ftpconnection.WithTLSConfig(nil),
				ftpconnection.WithVerboseWriter(bytes.NewBufferString("")),
			},
			expectedErrMsg: "an invalid argument error occurred: argument tlsConfig cannot be nil",
		},
		{
			name: "invalid verbose writer option",
			options: []ftpconnection.Option{
				ftpconnection.WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS13}),
				ftpconnection.WithVerboseWriter(nil),
			},
			expectedErrMsg: "an invalid argument error occurred: argument writer cannot be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tcpConn := ftpConnectionMocks.NewConn(t)
			dialer := ftpConnectionMocks.NewDialer(t)
			connMock := ftpConnectionMocks.NewTextConnection(t)

			serverConn, err := ftpconnection.NewConnection(
				host,
				dialer,
				tcpConn,
				connMock,
				tc.options...,
			)
			assert.Nil(t, serverConn)
			require.EqualError(t, err, tc.expectedErrMsg)
			assert.IsType(t, ftpErrors.InvalidArgumentErrorType, err)
			assert.NoError(t, errors.Unwrap(err))
		})
	}
}
