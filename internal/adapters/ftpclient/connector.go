package ftpclient

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"os"
	"strings"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

type ConnectorConfig struct {
	Address         string
	User            string
	Password        string
	Verbose         bool
	Timeout         time.Duration
	TLSCertFilePath string
	TLSKeyFilePath  string
	TLSInsecure     bool
}

func (c *ConnectorConfig) ServerName() string {
	const tokenSize = 2
	tokens := strings.SplitN(c.Address, ":", tokenSize)
	if len(tokens) != tokenSize {
		return c.Address
	}
	return tokens[0]
}

type Connector interface {
	Connect(ctx context.Context, config ConnectorConfig) (connection.Connection, error)
}

type connector struct {
}

//nolint:revive // connector is intended to be created with a constructor
func NewConnector() *connector {
	return &connector{}
}

func (c *connector) Connect(ctx context.Context, config ConnectorConfig) (conn connection.Connection, err error) {
	opts := make([]ftpconnection.Option, 0)
	if config.Verbose {
		opts = append(opts, ftpconnection.WithVerboseWriter(os.Stdout))
	}

	// if both TLS certificate and key are provided, dial FTP server with TLS configuration.
	if config.TLSCertFilePath != "" && config.TLSKeyFilePath != "" {
		tlsCfg, tlsCfgErr := getTLSConfig(config)
		if tlsCfgErr != nil {
			return nil, tlsCfgErr
		}

		opts = append(opts, ftpconnection.WithTLSConfig(tlsCfg))

		conn, err = ftpconnection.DialContextExplicitTLS(
			ctx,
			config.Address,
			tlsCfg,
			opts...,
		)
	} else {
		conn, err = ftpconnection.DialContext(
			ctx,
			config.Address,
			opts...,
		)
	}

	if err != nil {
		return nil, ftperrors.NewInternalError("failed to establish connection", err)
	}

	if loginErr := conn.Login(config.User, config.Password); loginErr != nil {
		defer func(conn connection.Connection) {
			if stopErr := conn.Stop(); stopErr != nil {
				err = stopErr
			}
		}(conn)
		return nil, ftperrors.NewInternalError("failed to authenticate with provided user account", loginErr)
	}
	return conn, nil
}

func getTLSConfig(config ConnectorConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(config.TLSCertFilePath, config.TLSCertFilePath)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to load X509 key pair", err)
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		Certificates: []tls.Certificate{
			cert,
		},
		Rand:               rand.Reader,
		Time:               time.Now,
		ServerName:         config.ServerName(),
		InsecureSkipVerify: config.TLSInsecure, //nolint:gosec // insecure skip verify is set with a flag
	}

	return cfg, nil
}
