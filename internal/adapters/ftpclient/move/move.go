package move

import (
	"context"
	"io"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	useCase "github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

type CmdMoveInput struct {
	Address  string
	User     string
	Password string
	Verbose  bool
	Timeout  time.Duration

	OldPath string
	NewPath string
}

type Dependencies struct {
	Connector ftpclient.Connector
	UseCase   useCase.MoveUseCase
	OutWriter io.Writer
}

func PerformMove(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdMoveInput) (err error) {
	options := &ftpclient.ConnectorOptions{
		Address:  input.Address,
		User:     input.User,
		Password: input.Password,
		Verbose:  input.Verbose,
	}
	conn, err := deps.Connector.Connect(
		ctx,
		options,
	)
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

	useCaseRepos := &useCase.MoveRepos{
		Logger:     logger,
		Connection: conn,
	}

	useCaseInput := &useCase.MoveInput{
		OldPath: input.OldPath,
		NewPath: input.NewPath,
	}

	if useCaseErr := deps.UseCase.Execute(ctx, useCaseRepos, useCaseInput); useCaseErr != nil {
		return useCaseErr
	}

	logger.Info("OK!")

	return nil
}
