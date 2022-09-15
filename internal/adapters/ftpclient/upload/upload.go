package upload

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
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

func PerformUploadFile(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdUploadInput) error {
	fileToUpload, err := deps.Filesystem.Open(input.FilePath)
	if err != nil {
		fmt.Println(err)
		return errors.NewInternalError("failed to open file", err)
	}
	defer fileToUpload.Close()
	fileInfo, err := fileToUpload.Stat()
	if err != nil {
		return errors.NewInternalError("failed to get file information", err)
	}

	// TODO: implement support for recursive upload of the whole directory structure
	if !fileInfo.Mode().IsRegular() {
		return errors.NewInternalError("path is not a regular file", nil)
	}

	fileSizeInBytes := fileInfo.Size()
	fileName := fileInfo.Name()

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

	p := mpb.New(mpb.WithWidth(64))
	bar := p.New(
		fileSizeInBytes,
		mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding("-").Rbound("]"),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(fmt.Sprintf("Uploading %s", fileName)),
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

	useCaseInput := &ftp.UploadFileInput{
		FileReader:    io.TeeReader(fileToUpload, cw),
		RemotePath:    input.RemoteFilePath,
		SizeInBytes:   uint64(fileSizeInBytes),
		CreateParents: input.CreateParents,
		Recursive:     input.Recursive,
	}

	if uploadErr := deps.UseCase.Execute(ctx, useCaseRepos, useCaseInput); uploadErr != nil {
		return uploadErr
	}
	p.Wait()
	logger.Info("OK!")

	return nil
}
