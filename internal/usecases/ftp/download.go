package ftp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/domain/repositories"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

type DownloadUseCase interface {
	Execute(context.Context, *DownloadRepos, *DownloadInput) error
}

type DownloadInput struct {
	RemotePath string
	Path       string
}

type DownloadRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
	FileStore  repositories.FileStore
}

type Download struct {
}

func (d *Download) Execute(ctx context.Context, repos *DownloadRepos, input *DownloadInput) error {
	isDir, err := repos.Connection.IsDir(ctx, input.RemotePath)
	if err != nil {
		repos.Logger.
			WithError(err).
			WithField("remote-path", input.RemotePath).
			Error("failed to check if entry is a directory")
		return ftperrors.NewInternalError("failed to check if entry is a directory", nil)
	}

	if !isDir {
		if downloadErr := d.downloadAndSaveFile(ctx, repos, input.RemotePath, input.Path); downloadErr != nil {
			return downloadErr
		}

		return nil
	}

	if downloadErr := d.downloadAndSaveFileRecursively(ctx, repos, input.RemotePath, input.Path); downloadErr != nil {
		return downloadErr
	}

	return nil
}

func (d *Download) downloadAndSaveFile(ctx context.Context, repos *DownloadRepos, remotePath, path string) error {
	logger := repos.Logger.WithField("remote-path", remotePath)

	sizeInBytes, err := repos.Connection.Size(remotePath)
	if err != nil {
		logger.WithError(err).Error("failed to retrieve file size")
		return ftperrors.NewInternalError("failed to retrieve file size", nil)
	}

	data, err := repos.Connection.Download(ctx, remotePath)
	if err != nil {
		logger.WithError(err).Error("failed to download file")
		return ftperrors.NewInternalError("failed to download file", nil)
	}

	downloadSizeInBytes := uint64(len(data))

	if sizeInBytes != downloadSizeInBytes {
		msg := fmt.Sprintf("downloaded file size %d does not match the actual %d", downloadSizeInBytes, sizeInBytes)
		logger.WithFields(
			logging.Fields{
				"actual-size-in-bytes":     sizeInBytes,
				"downloaded-size-in-bytes": downloadSizeInBytes,
			},
		).Error(msg)
		return ftperrors.NewInternalError(msg, nil)
	}

	if saveErr := repos.FileStore.SaveFile(path, data); saveErr != nil {
		logger.WithField("path", path).WithError(saveErr).Error("failed to save file")
		return ftperrors.NewInternalError("failed to save file", nil)
	}

	return nil
}

func (d *Download) downloadAndSaveFileRecursively(ctx context.Context, repos *DownloadRepos, remotePath, path string) error {
	if createDirErr := repos.FileStore.CreateDir(path); createDirErr != nil {
		repos.Logger.WithError(createDirErr).WithField("path", path).Error("failed to create directory")
		return ftperrors.NewInternalError("failed to create directory", nil)
	}

	entries, listErr := repos.Connection.List(ctx, &connection.ListOptions{
		Path:    remotePath,
		ShowAll: true,
	})
	if listErr != nil {
		repos.Logger.
			WithError(listErr).
			WithField("remote-path", remotePath).
			Error("failed to list directory")
		return ftperrors.NewInternalError("failed to list directory", nil)
	}

	for _, entry := range entries {
		if isRootDir(entry.Name) {
			continue
		}

		entryPath := filepath.Join(remotePath, entry.Name)
		localPath := filepath.Join(path, entry.Name)

		switch entry.Type {
		// ignore links as they not downloadable
		case entities.EntryTypeLink:
		case entities.EntryTypeFile:
			if downloadErr := d.downloadAndSaveFile(ctx, repos, entryPath, localPath); downloadErr != nil {
				return downloadErr
			}
		case entities.EntryTypeDir:
			if downloadErr := d.downloadAndSaveFileRecursively(ctx, repos, entryPath, localPath); downloadErr != nil {
				return downloadErr
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
