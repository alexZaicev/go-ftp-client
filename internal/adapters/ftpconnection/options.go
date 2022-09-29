package ftpconnection

import (
	"crypto/tls"
	"io"
	"net/textproto"

	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

type Option func(conn *ServerConnection) error

func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(conn *ServerConnection) error {
		if tlsConfig == nil {
			return errors.NewInvalidArgumentError("tlsConfig", errors.ErrMsgCannotBeNil)
		}
		conn.tlsConfig = tlsConfig
		return nil
	}
}

func WithVerboseWriter(writer io.Writer) Option {
	return func(conn *ServerConnection) error {
		if writer == nil {
			return errors.NewInvalidArgumentError("writer", errors.ErrMsgCannotBeNil)
		}
		conn.verboseWriter = writer
		// wrap existing connection
		conn.conn = textproto.NewConn(conn.wrapConnection(conn.tcpConn))
		return nil
	}
}

func WithDisabledUTF8() Option {
	return func(conn *ServerConnection) error {
		conn.disableUTF8 = true
		return nil
	}
}

func WithDisabledEPSV() Option {
	return func(conn *ServerConnection) error {
		conn.disableEPSV = true
		return nil
	}
}
