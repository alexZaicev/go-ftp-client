package filestore

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"

	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	dirMode os.FileMode = 0755
)

type FileStore struct{}

func (s *FileStore) SaveFile(path string, data []byte) error {
	parentDir, _ := filepath.Split(path)
	// check if parent dif exists, if not create directory tree
	if err := s.CreateDir(parentDir); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return ftperrors.NewInternalError("failed to create file", err)
	}

	var multiErr *multierror.Error

	if _, writeErr := file.Write(data); writeErr != nil {
		multiErr = multierror.Append(multiErr, writeErr)
	}

	if closeErr := file.Close(); closeErr != nil {
		multiErr = multierror.Append(multiErr, closeErr)
	}

	err = multiErr.ErrorOrNil()
	if err != nil {
		return ftperrors.NewInternalError("failed to create file", err)
	}

	return nil
}
func (s *FileStore) CreateDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return ftperrors.NewInternalError("failed to stat parent directory", err)
		}

		if mkdirErr := os.MkdirAll(path, dirMode); mkdirErr != nil {
			return ftperrors.NewInternalError("failed to create parent directories", err)
		}
	}

	return nil
}
