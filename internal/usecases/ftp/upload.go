package ftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

type UploadFileUseCase interface {
	Execute(context.Context, *UploadFileRepos, *UploadFileInput) error
}

type UploadFileInput struct {
	FileReader  io.Reader
	RemotePath  string
	SizeInBytes uint64
}

type UploadFileRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
}

type UploadFile struct {
}

func (u *UploadFile) Execute(_ context.Context, repos *UploadFileRepos, input *UploadFileInput) error {
	dirPath, fileName := filepath.Split(input.RemotePath)
	if strings.HasSuffix(dirPath, string(filepath.Separator)) {
		dirPath = dirPath[:len(dirPath)-1]
	}

	if dirPath != "" {
		if err := repos.Connection.Cd(dirPath); err != nil {
			var notFoundErr *ftperrors.NotFoundError
			if errors.As(err, &notFoundErr) {
				repos.Logger.WithError(notFoundErr).Error(fmt.Sprintf("directory %s not found", dirPath))
				return notFoundErr
			}

			repos.Logger.WithError(err).Error("failed to change directory")
			return ftperrors.NewInternalError("failed to change directory", nil)
		}
	}

	options := &connection.UploadOptions{
		FileReader: input.FileReader,
		Path:       fileName,
	}

	if err := repos.Connection.Upload(options); err != nil {
		repos.Logger.WithError(err).Error("failed to upload file")
		return ftperrors.NewInternalError("failed to upload file", nil)
	}

	sizeInBytes, err := repos.Connection.Size(input.RemotePath)
	if err != nil {
		repos.Logger.WithError(err).Error("failed to check file size")
		return ftperrors.NewInternalError("failed to check file size", nil)
	}

	if sizeInBytes != input.SizeInBytes {
		msg := fmt.Sprintf("uploaded file size %d does not match the actual %d", sizeInBytes, input.SizeInBytes)
		repos.Logger.WithFields(
			logging.Fields{
				"actual-size-in-bytes":   input.SizeInBytes,
				"uploaded-size-in-bytes": sizeInBytes,
			},
		).Error(msg)
		return ftperrors.NewInternalError(msg, nil)
	}

	return nil
}
