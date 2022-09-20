package ftpconnection

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/textproto"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	defaultConnectionTimeout = 20 * time.Second
)

func Dial(ctx context.Context, address string, options ...DialOption) (connection.Connection, error) {
	dialOpts := NewDialOptions()
	for _, option := range options {
		if err := option(dialOpts); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancel()

	var dialFunc func(network, address string) (net.Conn, error)

	if dialOpts.tlsConfig != nil && !dialOpts.explicitTLS {
		dialFunc = func(network, address string) (net.Conn, error) {
			tlsDialer := tls.Dialer{
				NetDialer: dialOpts.dialer,
				Config:    dialOpts.tlsConfig,
			}
			return tlsDialer.DialContext(ctx, network, address)
		}
	} else {
		dialFunc = func(network, address string) (net.Conn, error) {
			return dialOpts.dialer.DialContext(ctx, network, address)
		}
	}

	tcpConn, err := dialFunc("tcp", address)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed dial FTP server on [%s] address", address), err)
	}
	remoteAddr := tcpConn.RemoteAddr().(*net.TCPAddr)

	textConn := textproto.NewConn(dialOpts.wrapConnection(tcpConn))

	conn, err := newConnection(remoteAddr.IP.String(), tcpConn, textConn, dialOpts)
	if err != nil {
		return nil, err
	}
	if err := conn.Ready(); err != nil {
		return nil, err
	}

	if dialOpts.explicitTLS {
		if err := conn.EnableExplicitTLSMode(); err != nil {
			return nil, err
		}
	}

	return conn, nil
}
