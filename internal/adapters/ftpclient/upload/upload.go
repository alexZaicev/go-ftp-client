package upload

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

const (
	progressBarWidth = 64
)

type CmdUploadInput struct {
	Config         ftpclient.ConnectorConfig
	FilePath       string
	RemoteFilePath string
	Recursive      bool
}

type Dependencies struct {
	Connector     ftpclient.Connector
	Filesystem    fs.FS
	UploadUseCase ftp.UploadFileUseCase
	MkdirUseCase  ftp.MkdirUseCase
}

type fileToUpload struct {
	reader      fs.File
	sizeInBytes int64
	name        string
	path        string
}

func PerformUploadFile(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdUploadInput) (err error) {
	filesToUpload, err := getFilesToUpload(deps.Filesystem, input.FilePath, input.Recursive)
	if err != nil {
		return err
	}

	conn, err := deps.Connector.Connect(ctx, input.Config)
	if err != nil {
		logger.WithError(err).Error("failed to connect to server")
		return err
	}
	defer func(conn connection.Connection) {
		if stopErr := conn.Stop(); stopErr != nil {
			logger.WithError(stopErr).Error("failed to stop server connection")
			err = stopErr
		}
	}(conn)

	// FIXME: add a record of what directories have been created to avoid unnecessary calls
	for _, ftu := range filesToUpload {
		remoteFilePath := filepath.Join(input.RemoteFilePath, ftu.path[len(input.FilePath):])

		dirPath, _ := filepath.Split(remoteFilePath)
		if strings.HasSuffix(dirPath, string(filepath.Separator)) {
			dirPath = dirPath[:len(dirPath)-1]
		}

		if dirPath != "" {
			mkdirUseCaseInput := &ftp.MkdirInput{
				Path: dirPath,
			}
			mkdirUseCaseRepos := &ftp.MkdirRepos{
				Logger:     logger,
				Connection: conn,
			}

			if mkdirErr := deps.MkdirUseCase.Execute(ctx, mkdirUseCaseRepos, mkdirUseCaseInput); mkdirErr != nil {
				return mkdirErr
			}
		}

		// FIXME: add ability to write to progress bar writer, so that logs would be visible during the upload
		p := mpb.New(mpb.WithWidth(progressBarWidth))
		bar := p.New(
			ftu.sizeInBytes,
			mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding("-").Rbound("]"),
			mpb.PrependDecorators(
				// display our name with one space on the right
				decor.Name(fmt.Sprintf("Uploading %s", ftu.name)),
			),
			mpb.AppendDecorators(decor.Percentage()),
		)

		cw := &ftpclient.CallbackWriter{
			Callback: func(bytesRead int64) {
				bar.IncrInt64(bytesRead)
			},
		}

		uploadUseCaseRepos := &ftp.UploadFileRepos{
			Logger:     logger,
			Connection: conn,
		}

		uploadUseCaseInput := &ftp.UploadFileInput{
			FileReader:  io.TeeReader(ftu.reader, cw),
			RemotePath:  remoteFilePath,
			SizeInBytes: uint64(ftu.sizeInBytes),
		}

		if uploadErr := deps.UploadUseCase.Execute(ctx, uploadUseCaseRepos, uploadUseCaseInput); uploadErr != nil {
			return uploadErr
		}
		p.Wait()

		if closeErr := ftu.reader.Close(); closeErr != nil {
			logger.WithError(closeErr).Warn(fmt.Sprintf("failed to close file %s", ftu.path))
		}
	}

	logger.Info("OK!")

	return nil
}

func getFilesToUpload(filesystem fs.FS, filePath string, recursive bool) ([]*fileToUpload, error) {
	inputFile, err := filesystem.Open(trimLeadingSlash(filePath))
	if err != nil {
		return nil, errors.NewInternalError("failed to open file", err)
	}
	inputFileInfo, err := inputFile.Stat()
	if err != nil {
		return nil, errors.NewInternalError("failed to get file information", err)
	}

	var filesToUpload []*fileToUpload
	if recursive {
		if !inputFileInfo.Mode().IsDir() {
			return nil, errors.NewInternalError("path is not a directory", nil)
		}
		filesToUploadSlice, getErr := getFilesToUploadRecursively(filesystem, filePath)
		if getErr != nil {
			return nil, getErr
		}
		filesToUpload = filesToUploadSlice
	} else {
		if !inputFileInfo.Mode().IsRegular() {
			return nil, errors.NewInternalError("path is not a regular file", nil)
		}
		filesToUpload = append(filesToUpload, &fileToUpload{
			reader:      inputFile,
			sizeInBytes: inputFileInfo.Size(),
			name:        inputFileInfo.Name(),
			path:        filePath,
		})
	}
	return filesToUpload, nil
}

func getFilesToUploadRecursively(filesystem fs.FS, path string) ([]*fileToUpload, error) {
	var filePaths []*fileToUpload
	if err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		reader, err := filesystem.Open(trimLeadingSlash(path))
		if err != nil {
			return errors.NewInternalError("failed to open file", err)
		}

		filePaths = append(filePaths, &fileToUpload{
			reader:      reader,
			sizeInBytes: info.Size(),
			name:        info.Name(),
			path:        path,
		})
		return nil
	}); err != nil {
		return nil, errors.NewInternalError("failed to walk directory", err)
	}
	return filePaths, nil
}

func trimLeadingSlash(path string) string {
	if strings.HasPrefix(path, "/") {
		return path[1:]
	}
	return path
}
