package ftpconnection

import (
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"net"
)

func NewServerConnection(
	host string,
	tcpConn net.Conn,
	textConn TextConnection,
	options *DialOptions,
) (connection.Connection, error) {
	return newConnection(host, tcpConn, textConn, options)
}
