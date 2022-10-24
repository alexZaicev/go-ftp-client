package ftpconnection

import (
	"fmt"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Cd(path string) error {
	code, msg, err := c.cmd(models.StatusNoCheck, models.CommandChangeWorkDir, path)
	if err != nil {
		return ftperrors.NewInternalError("failed to change working directory", err)
	}
	if code == models.StatusRequestedFileActionOK {
		return nil
	}
	if code == models.StatusFileUnavailable {
		return ftperrors.NewNotFoundError(fmt.Sprintf("path %s does not exist", path), nil)
	}
	return ftperrors.NewInternalError(msg, nil)
}
