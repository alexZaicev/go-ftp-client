package ftpconnection

import (
	"strconv"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Size(path string) (uint64, error) {
	_, msg, err := c.cmd(models.StatusFile, models.CommandSize, path)
	if err != nil {
		return 0, ftperrors.NewInternalError("failed to fetch file size", err)
	}
	sizeInBytes, err := strconv.ParseUint(msg, decimalBase, bitSize)
	if err != nil {
		return 0, ftperrors.NewInternalError("failed to parse file size to a non-zero integer", err)
	}
	return sizeInBytes, err
}
