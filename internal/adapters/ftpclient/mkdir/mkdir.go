package mkdir

import (
	"context"
	"io"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	useCase "github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

type CmdMkdirInput struct {
	Config ftpclient.ConnectorConfig
	Path   string
}

type Dependencies struct {
	Connector ftpclient.Connector
	UseCase   useCase.MkdirUseCase
	OutWriter io.Writer
}

func PerformMkdir(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdMkdirInput) (err error) {
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

	useCaseRepos := &useCase.MkdirRepos{
		Logger:     logger,
		Connection: conn,
	}

	useCaseInput := &useCase.MkdirInput{
		Path: input.Path,
	}

	if useCaseErr := deps.UseCase.Execute(ctx, useCaseRepos, useCaseInput); useCaseErr != nil {
		return useCaseErr
	}

	logger.Info("OK!")

	return nil
}
