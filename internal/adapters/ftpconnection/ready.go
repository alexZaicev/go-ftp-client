package ftpconnection

import (
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

// Ready function validates that the FTP server is ready to proceed.
func (c *ServerConnection) Ready() (err error) {
	if _, _, readErr := c.conn.ReadResponse(models.StatusReady); readErr != nil {
		defer func() {
			if stopErr := c.Stop(); stopErr != nil {
				err = stopErr
			}
		}()
		return ftperrors.NewInternalError("failed to check if server is ready", readErr)
	}
	return nil
}
