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
	Path string
}

type RemoveRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
}

type Remove struct {
}

func (u *Remove) Execute(ctx context.Context, repos *RemoveRepos, input *RemoveInput) error {
	isDir, err := repos.Connection.IsDir(ctx, input.Path)
	if err != nil {
		repos.Logger.
			WithError(err).
			WithField("remote-path", input.Path).
			Error("failed to check if entry is a directory")
		return ftperrors.NewInternalError("failed to check if entry is a directory", nil)
	}

	if !isDir {
		if removeErr := repos.Connection.RemoveFile(input.Path); removeErr != nil {
			repos.Logger.
				WithError(removeErr).
				WithField("remote-path", input.Path).
				Error("failed to remove file")
			return ftperrors.NewInternalError("failed to remove file", nil)
		}

		return nil
	}

	// recursively remove contents of the provided directory; the directory itself
	// will be removed in the later connection call.
	if removeErr := u.removeRecursive(ctx, repos, input.Path); removeErr != nil {
		return removeErr
	}

	return nil
}

func (u *Remove) removeRecursive(ctx context.Context, repos *RemoveRepos, path string) error {
	entries, listErr := repos.Connection.List(ctx, &connection.ListOptions{
		Path:    path,
		ShowAll: true,
	})
	if listErr != nil {
		repos.Logger.
			WithError(listErr).
			WithField("remote-path", path).
			Error("failed to list directory")
		return ftperrors.NewInternalError("failed to list directory", nil)
	}

	for _, entry := range entries {
		if isRootDir(entry.Name) {
			continue
		}

		entryPath := filepath.Join(path, entry.Name)
		logger := repos.Logger.WithField("remote-path", entryPath)

		switch entry.Type {
		case entities.EntryTypeFile, entities.EntryTypeLink:
			if removeErr := repos.Connection.RemoveFile(entryPath); removeErr != nil {
				logger.WithError(removeErr).Error("failed to remove file")
				return ftperrors.NewInternalError("failed to remove file", nil)
			}
		case entities.EntryTypeDir:
			if removeErr := u.removeRecursive(ctx, repos, entryPath); removeErr != nil {
				return removeErr
			}
		default:
			return ftperrors.NewUnknownError(
				fmt.Sprintf("unexpected entry type: %d", entry.Type),
				nil,
			)
		}
	}

	if removeErr := repos.Connection.RemoveDir(path); removeErr != nil {
		repos.Logger.
			WithError(removeErr).
			WithField("remote-path", path).
			Error("failed to remove directory")
		return ftperrors.NewInternalError("failed to remove directory", nil)
	}

	return nil
}
