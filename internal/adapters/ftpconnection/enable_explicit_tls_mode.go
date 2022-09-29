package ftpconnection

import (
	"crypto/tls"
	"net/textproto"

	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

// EnableExplicitTLSMode function enables TLS modes on established TCP connection.
func (c *ServerConnection) EnableExplicitTLSMode() (err error) {
	if _, _, readErr := c.cmd(StatusAuthOK, CommandAuthTLS); readErr != nil {
		defer func() {
			if stopErr := c.Stop(); stopErr != nil {
				err = stopErr
			}
		}()
		return ftperrors.NewInternalError("failed to enable explicit TLS mode", readErr)
	}
	tlsConn := tls.Client(c.tcpConn, c.tlsConfig)
	c.tcpConn = tlsConn
	c.conn = textproto.NewConn(c.wrapConnection(tlsConn))
	return nil
}
