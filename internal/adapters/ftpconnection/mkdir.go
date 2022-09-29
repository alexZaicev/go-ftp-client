package ftpconnection

import (
	"errors"
	"path/filepath"
	"strings"

	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Mkdir(path string) error {
	if path == "" {
		return ftperrors.NewInvalidArgumentError("path", ftperrors.ErrMsgCannotBeBlank)
	}

	if strings.HasPrefix(path, string(filepath.Separator)) {
		path = path[1:]
	}
	pathTokens := strings.Split(path, string(filepath.Separator))

	var builder strings.Builder
	builder.WriteString("/")
	for idx, pathToken := range pathTokens {
		if idx > 0 {
			builder.WriteRune(filepath.Separator)
		}
		builder.WriteString(pathToken)

		pathToCreate := builder.String()

		if err := c.Cd(pathToCreate); err != nil {
			var notFoundErr *ftperrors.NotFoundError
			if !errors.As(err, &notFoundErr) {
				return err
			}

			_, _, mkdErr := c.cmd(StatusPathCreated, CommandMakeDir, pathToCreate)
			if mkdErr != nil {
				return ftperrors.NewInternalError("failed to create directory", mkdErr)
			}
		}
	}

	if err := c.Cd("/"); err != nil {
		return err
	}

	return nil
}
