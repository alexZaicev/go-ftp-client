package ftp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

type RemoveUseCase interface {
	Execute(context.Context, *RemoveRepos, *RemoveInput) error
}

type RemoveInput struct {
	Path      string
	Recursive bool
}

type RemoveRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
}

type Remove struct {
}

func (u *Remove) Execute(ctx context.Context, repos *RemoveRepos, input *RemoveInput) error {
	logger := repos.Logger.WithField("path", input.Path)

	isDir, err := u.isDir(ctx, repos, input.Path)
	if err != nil {
		return err
	}

	if !isDir {
		if removeErr := repos.Connection.RemoveFile(input.Path); removeErr != nil {
			logger.WithError(removeErr).Error("failed to remove file")
			return ftperrors.NewInternalError("failed to remove file", nil)
		}

		return nil
	}

	if input.Recursive {
		// recursively remove contents of the provided directory; the directory itself
		// will be removed in the later connection call.
		if removeErr := u.removeRecursive(ctx, repos, input.Path); removeErr != nil {
			return removeErr
		}
	}

	if removeErr := repos.Connection.RemoveDir(input.Path); removeErr != nil {
		logger.WithError(removeErr).Error("failed to remove directory")
		return ftperrors.NewInternalError("failed to remove directory", nil)
	}

	return nil
}

func (u *Remove) isDir(ctx context.Context, repos *RemoveRepos, path string) (bool, error) {
	parentDir, fileToRemove := filepath.Split(path)
	entries, err := repos.Connection.List(ctx, &connection.ListOptions{
		Path:    parentDir,
		ShowAll: true,
	})
	if err != nil {
		if parentDir == "" {
			parentDir = "/"
		}
		repos.Logger.WithError(err).WithField("path", parentDir).Error("failed to list directory")
		return false, ftperrors.NewInternalError("failed to list directory", nil)
	}

	var entry *entities.Entry
	for _, e := range entries {
		if e.Name == fileToRemove {
			entry = e
			break
		}
	}

	if entry == nil {
		msg := fmt.Sprintf("entry not found under [%s] path", path)
		repos.Logger.WithField("path", path).Info(msg)
		return false, ftperrors.NewNotFoundError(msg, nil)
	}

	isDir := entry.Type == entities.EntryTypeDir

	return isDir, nil
}

func (u *Remove) removeRecursive(ctx context.Context, repos *RemoveRepos, path string) error {
	entries, listErr := repos.Connection.List(ctx, &connection.ListOptions{
		Path:    path,
		ShowAll: true,
	})
	if listErr != nil {
		repos.Logger.
			WithError(listErr).
			WithField("path", path).
			Error("failed to list directory")
		return ftperrors.NewInternalError("failed to list directory", nil)
	}

	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		entryPath := filepath.Join(path, entry.Name)
		logger := repos.Logger.WithField("path", entryPath)

		switch entry.Type {
		case entities.EntryTypeFile, entities.EntryTypeLink:
			if removeErr := repos.Connection.RemoveFile(entryPath); removeErr != nil {
				logger.WithError(removeErr).Error("failed to remove file")
				return ftperrors.NewInternalError("failed to remove file", nil)
			}
		case entities.EntryTypeDir:
			if removeErr := u.removeRecursive(ctx, repos, entryPath); removeErr != nil {
				logger.WithError(removeErr).Error("failed to recursively remove directory")
				return ftperrors.NewInternalError("failed to recursively remove directory", nil)
			}

			if removeErr := repos.Connection.RemoveDir(entryPath); removeErr != nil {
				logger.WithError(removeErr).Error("failed to remove directory")
				return ftperrors.NewInternalError("failed to remove directory", nil)
			}
		default:
			return ftperrors.NewUnknownError(
				fmt.Sprintf("unexpected entry type: %d", entry.Type),
				nil,
			)
		}
	}

	return nil
}
