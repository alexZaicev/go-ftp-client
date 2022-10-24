package ftpconnection

import (
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) RemoveFile(path string) error {
	if path == "" {
		return ftperrors.NewInvalidArgumentError("path", ftperrors.ErrMsgCannotBeBlank)
	}

	if _, _, err := c.cmd(models.StatusRequestedFileActionOK, models.CommandRemoveFile, path); err != nil {
		return ftperrors.NewInternalError("failed to remove file", err)
	}
	return nil
}

func (c *ServerConnection) RemoveDir(path string) error {
	if path == "" {
		return ftperrors.NewInvalidArgumentError("path", ftperrors.ErrMsgCannotBeBlank)
	}

	if _, _, err := c.cmd(models.StatusRequestedFileActionOK, models.CommandRemoveDir, path); err != nil {
		return ftperrors.NewInternalError("failed to remove directory", err)
	}
	return nil
}
