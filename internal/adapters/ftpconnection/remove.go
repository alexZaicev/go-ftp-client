package ftpconnection

import ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"

func (c *ServerConnection) RemoveFile(path string) error {
	if path == "" {
		return ftperrors.NewInvalidArgumentError("path", ftperrors.ErrMsgCannotBeBlank)
	}

	if _, _, err := c.cmd(StatusRequestedFileActionOK, CommandRemoveFile, path); err != nil {
		return ftperrors.NewInternalError("failed to remove file", err)
	}
	return nil
}

func (c *ServerConnection) RemoveDir(path string) error {
	if path == "" {
		return ftperrors.NewInvalidArgumentError("path", ftperrors.ErrMsgCannotBeBlank)
	}

	if _, _, err := c.cmd(StatusRequestedFileActionOK, CommandRemoveDir, path); err != nil {
		return ftperrors.NewInternalError("failed to remove directory", err)
	}
	return nil
}
