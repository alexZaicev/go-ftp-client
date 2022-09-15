package ftp

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

type UploadFileUseCase interface {
	Execute(context.Context, *UploadFileRepos, *UploadFileInput) error
}

type UploadFileInput struct {
	FileReader    io.Reader
	RemotePath    string
	SizeInBytes   uint64
	CreateParents bool
	Recursive     bool
}

type UploadFileRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
}

type UploadFile struct {
}

func (u *UploadFile) Execute(ctx context.Context, repos *UploadFileRepos, input *UploadFileInput) error {
	// TODO recursive file upload
	dirPath, fileName := filepath.Split(input.RemotePath)
	if strings.HasSuffix(dirPath, string(filepath.Separator)) {
		dirPath = dirPath[:len(dirPath)-1]
	}

	if dirPath != "" {
		if input.CreateParents {
			options := &connection.MkdirOptions{
				Path:          dirPath,
				CreateParents: true,
			}

			if err := repos.Connection.Mkdir(options); err != nil {
				repos.Logger.WithError(err).Error("failed to create parent directories")
				return errors.NewInternalError("failed to create parent directories", nil)
			}
		}

		if err := repos.Connection.Cd(dirPath); err != nil {
			repos.Logger.WithError(err).Error("failed to change directory")
			return errors.NewInternalError("failed to change directory", nil)
		}
	}

	options := &connection.UploadOptions{
		FileReader: input.FileReader,
		Path:       fileName,
	}

	if err := repos.Connection.Upload(options); err != nil {
		repos.Logger.WithError(err).Error("failed to upload file")
		return errors.NewInternalError("failed to upload file", nil)
	}

	// TODO validate file size

	return nil
}
