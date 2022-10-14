package ftpconnection

import ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"

func (c *ServerConnection) Move(oldPath string, newPath string) error {
	if oldPath == "" {
		return ftperrors.NewInvalidArgumentError("oldPath", ftperrors.ErrMsgCannotBeBlank)
	}
	if newPath == "" {
		return ftperrors.NewInvalidArgumentError("newPath", ftperrors.ErrMsgCannotBeBlank)
	}

	if _, _, err := c.cmd(StatusRequestFilePending, CommandRenameFrom, oldPath); err != nil {
		return ftperrors.NewInternalError("failed to prepare file", err)
	}

	if _, _, err := c.cmd(StatusRequestedFileActionOK, CommandRenameTo, newPath); err != nil {
		return ftperrors.NewInternalError("failed to move file", err)
	}

	return nil
}
