package ftpconnection

import (
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	defaultShutTimeout = 500 * time.Millisecond
)

// DialOptions holds consumer defined dial-up options
type DialOptions struct {
	tlsConfig *tls.Config
	dialer    *net.Dialer
	location  *time.Location

	verboseWriter io.Writer

	shutTimeout time.Duration

	// TODO: create option func to set bellow flags
	explicitTLS bool
	disableMLST bool
	writingMDTM bool
	disableUTF8 bool
	disableEPSV bool
}

type DialOption func(options *DialOptions) error

// NewDialOptions create a DialOptions struct with default configurations
func NewDialOptions() *DialOptions {
	return &DialOptions{
		location:    time.UTC,
		dialer:      &net.Dialer{},
		shutTimeout: defaultShutTimeout,
	}
}

func (do *DialOptions) wrapConnection(conn net.Conn) io.ReadWriteCloser {
	if do.verboseWriter == nil {
		return conn
	}
	return newVerboseConnectionWrapper(conn, do.verboseWriter)
}

func WithTimeout(timeout time.Duration) DialOption {
	return func(options *DialOptions) error {
		options.dialer.Timeout = timeout
		return nil
	}
}

func WithTLSConfig(tlsConfig *tls.Config) DialOption {
	return func(options *DialOptions) error {
		if tlsConfig == nil {
			return errors.NewInvalidArgumentError("tlsConfig", errors.ErrMsgCannotBeNil)
		}
		options.tlsConfig = tlsConfig
		return nil
	}
}

func WithExplicitTLSConfig(tlsConfig *tls.Config) DialOption {
	return func(options *DialOptions) error {
		if tlsConfig == nil {
			return errors.NewInvalidArgumentError("tlsConfig", errors.ErrMsgCannotBeNil)
		}
		options.tlsConfig = tlsConfig
		options.explicitTLS = true
		return nil
	}
}

func WithLocation(location *time.Location) DialOption {
	return func(options *DialOptions) error {
		if location == nil {
			return errors.NewInvalidArgumentError("location", errors.ErrMsgCannotBeNil)
		}
		options.location = location
		return nil
	}
}

func WithDialer(dialer *net.Dialer) DialOption {
	return func(options *DialOptions) error {
		if dialer == nil {
			return errors.NewInvalidArgumentError("dialer", errors.ErrMsgCannotBeNil)
		}
		options.dialer = dialer
		return nil
	}
}

func WithVerboseWriter(verboseWriter io.Writer) DialOption {
	return func(options *DialOptions) error {
		options.verboseWriter = verboseWriter
		return nil
	}
}
