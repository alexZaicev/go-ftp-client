package status

import (
	"context"
	"io"

	"github.com/olekukonko/tablewriter"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	useCase "github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

type CmdStatusInput struct {
	Config ftpclient.ConnectorConfig
}

type Dependencies struct {
	Connector ftpclient.Connector
	UseCase   useCase.StatusUseCase
	OutWriter io.Writer
}

func PerformStatus(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdStatusInput) (err error) {
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

	useCaseRepos := &useCase.StatusRepos{
		Logger:     logger,
		Connection: conn,
	}

	useCaseInput := &useCase.StatusInput{}

	status, err := deps.UseCase.Execute(ctx, useCaseRepos, useCaseInput)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(deps.OutWriter)
	table.SetHeader([]string{
		"status",
		"system",
		"remote address",
		"logged in user",
		"tls enabled",
	})

	tlsEnabled := "NO"
	if status.TLSEnabled {
		tlsEnabled = "YES"
	}

	table.Append([]string{
		"OK",
		status.System,
		status.RemoteAddress,
		status.LoggedInUser,
		tlsEnabled,
	})
	table.Render()

	return nil
}
