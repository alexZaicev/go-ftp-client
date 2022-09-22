package ftpconnection

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/textproto"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	defaultConnectionTimeout = 5 * time.Second
)

type dialer struct {
	dialer *net.Dialer
}

func newDialer() *dialer {
	return &dialer{
		dialer: new(net.Dialer),
	}
}

func (d *dialer) Dial(network, address string) (net.Conn, error) {
	return d.dialer.Dial(network, address)
}

func (d *dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.dialer.DialContext(ctx, network, address)
}

func (d *dialer) DialContextTLS(ctx context.Context, network, address string, tlsConfig *tls.Config) (net.Conn, error) {
	tlsDialer := &tls.Dialer{
		NetDialer: d.dialer,
		Config:    tlsConfig,
	}
	return tlsDialer.DialContext(ctx, network, address)
}

func DialContext(ctx context.Context, address string, options ...Option) (connection.Connection, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancel()

	d := newDialer()
	tcpConn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, ftperrors.NewInternalError(fmt.Sprintf("failed dial server on [%s] address", address), err)
	}
	remoteAddr := tcpConn.RemoteAddr().(*net.TCPAddr)
	host := remoteAddr.IP.String()

	textConn := textproto.NewConn(tcpConn)

	sc, err := NewConnection(host, d, tcpConn, textConn, options...)
	if err != nil {
		return nil, err
	}

	if readyErr := sc.Ready(); readyErr != nil {
		return nil, readyErr
	}

	return sc, nil
}

func DialContextTLS(ctx context.Context, address string, tlsConfig *tls.Config, options ...Option) (connection.Connection, error) {
	return dialContextTLS(ctx, address, tlsConfig, false, options...)
}

func DialContextExplicitTLS(ctx context.Context, address string, tlsConfig *tls.Config, options ...Option) (connection.Connection, error) {
	return dialContextTLS(ctx, address, tlsConfig, true, options...)
}

func dialContextTLS(
	ctx context.Context,
	address string,
	tlsConfig *tls.Config,
	explicitTLS bool,
	options ...Option,
) (connection.Connection, error) {
	if tlsConfig == nil {
		return nil, ftperrors.NewInvalidArgumentError("tlsConfig", ftperrors.ErrMsgCannotBeNil)
	}

	ctx, cancel := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancel()

	var tcpConn net.Conn
	var err error

	d := newDialer()
	if explicitTLS {
		tcpConn, err = d.DialContext(ctx, "tcp", address)
	} else {
		tcpConn, err = d.DialContextTLS(ctx, "tcp", address, tlsConfig)
	}

	if err != nil {
		return nil, ftperrors.NewInternalError(fmt.Sprintf("failed dial server on [%s] address", address), err)
	}

	remoteAddr := tcpConn.RemoteAddr().(*net.TCPAddr)
	host := remoteAddr.IP.String()

	textConn := textproto.NewConn(tcpConn)

	sc, err := NewConnection(host, d, tcpConn, textConn, options...)
	if err != nil {
		return nil, err
	}

	if readyErr := sc.Ready(); readyErr != nil {
		return nil, readyErr
	}

	if tlsErr := sc.EnableExplicitTLSMode(); tlsErr != nil {
		return nil, tlsErr
	}

	return sc, nil
}
