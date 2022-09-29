package ftpclient

import (
	"context"
	"os"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

type ConnectorOptions struct {
	Address  string
	User     string
	Password string
	Verbose  bool
}

type Connector interface {
	Connect(ctx context.Context, options *ConnectorOptions) (connection.Connection, error)
}

type connector struct {
}

func NewConnector() *connector {
	return &connector{}
}

func (c *connector) Connect(ctx context.Context, options *ConnectorOptions) (conn connection.Connection, err error) {
	opts := make([]ftpconnection.Option, 0)
	if options.Verbose {
		opts = append(opts, ftpconnection.WithVerboseWriter(os.Stdout))
	}

	conn, err = ftpconnection.DialContext(
		ctx,
		options.Address,
		opts...,
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to establish connection", err)
	}

	if loginErr := conn.Login(options.User, options.Password); loginErr != nil {
		defer func() {
			if stopErr := conn.Stop(); stopErr != nil {
				err = stopErr
			}
		}()
		return nil, errors.NewInternalError("failed to authenticate with provided user account", loginErr)
	}
	return conn, nil
}
