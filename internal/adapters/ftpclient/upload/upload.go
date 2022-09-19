package upload

import (
	"context"
	"fmt"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

type CmdUploadInput struct {
	Address  string
	User     string
	Password string
	Verbose  bool
	Timeout  time.Duration

	FilePath       string
	RemoteFilePath string
	CreateParents  bool
	Recursive      bool
}

type Dependencies struct {
	Filesystem fs.FS
	UseCase    ftp.UploadFileUseCase
}

type fileToUpload struct {
	reader      fs.File
	sizeInBytes int64
	name        string
	path        string
}

func PerformUploadFile(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdUploadInput) error {
	inputFile, err := deps.Filesystem.Open(trimLeadingSlash(input.FilePath))
	if err != nil {
		return errors.NewInternalError("failed to open file", err)
	}
	inputFileInfo, err := inputFile.Stat()
	if err != nil {
		return errors.NewInternalError("failed to get file information", err)
	}

	var filesToUpload []*fileToUpload
	if input.Recursive {
		if !inputFileInfo.Mode().IsDir() {
			return errors.NewInternalError("path is not a directory", nil)
		}
		filesToUploadSlice, getErr := getFilesToUpload(deps.Filesystem, input.FilePath)
		if getErr != nil {
			return getErr
		}
		filesToUpload = filesToUploadSlice
	} else {
		if !inputFileInfo.Mode().IsRegular() {
			return errors.NewInternalError("path is not a regular file", nil)
		}
		filesToUpload = append(filesToUpload, &fileToUpload{
			reader:      inputFile,
			sizeInBytes: inputFileInfo.Size(),
			name:        inputFileInfo.Name(),
			path:        input.FilePath,
		})
	}

	conn, err := ftpclient.Connect(
		ctx,
		input.Address,
		input.User,
		input.Password,
		input.Timeout,
		input.Verbose,
	)
	if err != nil {
		return err
	}
	defer conn.Stop()

	for _, ftu := range filesToUpload {
		p := mpb.New(mpb.WithWidth(64))
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

		useCaseRepos := &ftp.UploadFileRepos{
			Logger:     logger,
			Connection: conn,
		}

		remoteFilePath := filepath.Join(input.RemoteFilePath, ftu.path[len(input.FilePath):])

		useCaseInput := &ftp.UploadFileInput{
			FileReader:    io.TeeReader(ftu.reader, cw),
			RemotePath:    remoteFilePath,
			SizeInBytes:   uint64(ftu.sizeInBytes),
			CreateParents: input.CreateParents,
		}

		if uploadErr := deps.UseCase.Execute(ctx, useCaseRepos, useCaseInput); uploadErr != nil {
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

func getFilesToUpload(filesystem fs.FS, path string) ([]*fileToUpload, error) {
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
		return nil, err
	}
	return filePaths, nil
}

func trimLeadingSlash(path string) string {
	if strings.HasPrefix(path, "/") {
		return path[1:]
	}
	return path
}
