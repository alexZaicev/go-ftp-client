package ftpconnection

import (
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Move(oldPath string, newPath string) error {
	if oldPath == "" {
		return ftperrors.NewInvalidArgumentError("oldPath", ftperrors.ErrMsgCannotBeBlank)
	}
	if newPath == "" {
		return ftperrors.NewInvalidArgumentError("newPath", ftperrors.ErrMsgCannotBeBlank)
	}

	if _, _, err := c.cmd(models.StatusRequestFilePending, models.CommandRenameFrom, oldPath); err != nil {
		return ftperrors.NewInternalError("failed to prepare file", err)
	}

	if _, _, err := c.cmd(models.StatusRequestedFileActionOK, models.CommandRenameTo, newPath); err != nil {
		return ftperrors.NewInternalError("failed to move file", err)
	}

	return nil
}
