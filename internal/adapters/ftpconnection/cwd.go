package ftpconnection

import (
	"fmt"

	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Cd(path string) error {
	code, msg, err := c.cmd(StatusNoCheck, CommandChangeWorkDir, path)
	if err != nil {
		return ftperrors.NewInternalError("failed to change working directory", err)
	}
	if code == StatusRequestedFileActionOK {
		return nil
	}
	if code == StatusFileUnavailable {
		return ftperrors.NewNotFoundError(fmt.Sprintf("path %s does not exist", path), nil)
	}
	return ftperrors.NewInternalError(msg, nil)
}
