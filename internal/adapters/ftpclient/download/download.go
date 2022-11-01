package download

import (
	"context"
	"io"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/repositories"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

type CmdDownloadInput struct {
	Config     ftpclient.ConnectorConfig
	RemotePath string
	Path       string
}

type Dependencies struct {
	Connector ftpclient.Connector
	FileStore repositories.FileStore
	UseCase   ftp.DownloadUseCase
	OutWriter io.Writer
}

func PerformDownload(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdDownloadInput) (err error) {
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

	downloadUseCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: conn,
		FileStore:  deps.FileStore,
	}

	downloadUseCaseInput := &ftp.DownloadInput{
		RemotePath: input.RemotePath,
		Path:       input.Path,
	}

	if downloadErr := deps.UseCase.Execute(ctx, downloadUseCaseRepos, downloadUseCaseInput); downloadErr != nil {
		return downloadErr
	}

	logger.Info("OK!")

	return nil
}
