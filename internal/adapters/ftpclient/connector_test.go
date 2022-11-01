package ftpclient_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	certFilePath = "./testdata/vsftpd.pem"
	keyFilePath  = "./testdata/vsftpd.pem"
)

func Test_ConnectorConfig_ServerName_Success(t *testing.T) {
	testCases := []struct {
		name         string
		address      string
		expectedName string
	}{
		{
			name:         "domain name wit port",
			address:      "ftp.example.com:21",
			expectedName: "ftp.example.com",
		},
		{
			name:         "no port",
			address:      "ftp.example.com",
			expectedName: "ftp.example.com",
		},
		{
			name:         "only port",
			address:      ":21",
			expectedName: "",
		},
		{
			name:         "IPv4",
			address:      "172.0.20.12:21",
			expectedName: "172.0.20.12",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			cfg := ftpclient.ConnectorConfig{
				Address: tc.address,
			}

			// act
			name := cfg.ServerName()

			// assert
			assert.Equal(t, tc.expectedName, name)
		})
	}
}

func Test_NewConnector_Success(t *testing.T) {
	// arrange
	// act
	conn := ftpclient.NewConnector()

	// assert
	assert.NotNil(t, conn)
}

func Test_Connector_Connect_Success(t *testing.T) {
	testCases := []struct {
		name            string
		tlsCertFilePath string
		tlsKeyFilePath  string
	}{
		{
			name: "connection with no TLS",
		},
		{
			name:            "connection with TLS",
			tlsCertFilePath: certFilePath,
			tlsKeyFilePath:  keyFilePath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			ctx := context.Background()

			serverMock, err := newFtpMock(t)
			require.NoError(t, err, "error starting FTP mock")
			defer serverMock.Close()

			config := ftpclient.ConnectorConfig{
				Address:         serverMock.address,
				User:            anonymous,
				Password:        anonymous,
				Verbose:         true,
				TLSCertFilePath: tc.tlsCertFilePath,
				TLSKeyFilePath:  tc.tlsKeyFilePath,
				TLSInsecure:     true,
			}

			connector := ftpclient.NewConnector()

			// act
			conn, err := connector.Connect(ctx, config)

			// assert
			assert.NoError(t, err)
			assert.NotNil(t, conn)
		})
	}
}

func Test_Connector_Connect_LoadX509Error(t *testing.T) {
	// arrange
	ctx := context.Background()

	config := ftpclient.ConnectorConfig{
		Address:         "localhost:12345",
		User:            anonymous,
		Password:        anonymous,
		Verbose:         true,
		TLSCertFilePath: "not-valid-path",
		TLSKeyFilePath:  "not-valid-apth",
		TLSInsecure:     true,
	}

	connector := ftpclient.NewConnector()

	// act
	conn, err := connector.Connect(ctx, config)

	// assert
	assert.Nil(t, conn)

	require.EqualError(t, err, "an internal error occurred: failed to load X509 key pair")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "open not-valid-path: no such file or directory")
}

func Test_Connector_Connect_DialError(t *testing.T) {
	// arrange
	ctx := context.Background()

	config := ftpclient.ConnectorConfig{
		Address:  "not-valid-host",
		User:     anonymous,
		Password: anonymous,
		Verbose:  true,
	}

	connector := ftpclient.NewConnector()

	// act
	conn, err := connector.Connect(ctx, config)

	// assert
	assert.Nil(t, conn)

	require.EqualError(t, err, "an internal error occurred: failed to establish connection")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "an internal error occurred: failed dial server on [not-valid-host] address")
}

func Test_Connector_Connect_LoginError(t *testing.T) {
	// arrange
	ctx := context.Background()

	serverMock, err := newFtpMock(t)
	require.NoError(t, err, "error starting FTP mock")
	defer serverMock.Close()

	config := ftpclient.ConnectorConfig{
		Address:  serverMock.address,
		User:     "not-valid-user",
		Password: anonymous,
		Verbose:  true,
	}

	connector := ftpclient.NewConnector()

	// act
	conn, err := connector.Connect(ctx, config)

	// assert
	assert.Nil(t, conn)

	require.EqualError(t, err, "an internal error occurred: failed to authenticate with provided user account")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "an internal error occurred: This FTP server is anonymous only")
}
