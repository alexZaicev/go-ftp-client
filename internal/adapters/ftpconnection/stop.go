package ftpconnection

import (
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

// Stop function sends a quit command to FTP server and closes the TCP connection.
func (c *ServerConnection) Stop() (err error) {
	defer func() {
		if closeErr := c.conn.Close(); closeErr != nil {
			err = ftperrors.NewInternalError("failed to close connection", closeErr)
		}
	}()
	if _, cmdErr := c.conn.Cmd(models.CommandQuit); cmdErr != nil {
		return ftperrors.NewInternalError("failed to disconnect from the server", cmdErr)
	}
	return nil
}
