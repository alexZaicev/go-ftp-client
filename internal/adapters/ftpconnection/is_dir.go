package ftpconnection

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) IsDir(ctx context.Context, path string) (bool, error) {
	parentDir, fileToRemove := filepath.Split(path)
	entries, err := c.List(ctx, &connection.ListOptions{
		Path:    parentDir,
		ShowAll: true,
	})
	if err != nil {
		return false, err
	}

	var entry *entities.Entry
	for _, e := range entries {
		if e.Name == fileToRemove {
			entry = e
			break
		}
	}

	if entry == nil {
		return false, ftperrors.NewNotFoundError(fmt.Sprintf("entry not found under %q path", path), nil)
	}

	isDir := entry.Type == entities.EntryTypeDir

	return isDir, nil
}
