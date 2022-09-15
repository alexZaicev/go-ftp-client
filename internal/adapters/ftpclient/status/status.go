package status

import (
	"context"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	useCase "github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

type CmdStatusInput struct {
	Address  string
	User     string
	Password string
	Verbose  bool
	Timeout  time.Duration
}

type Dependencies struct {
	UseCase useCase.StatusUseCase
}

func PerformStatus(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdStatusInput) error {
	conn, err := ftpclient.Connect(
		ctx,
		input.Address,
		input.User,
		input.Password,
		input.Timeout,
		input.Verbose,
	)
	if err != nil {
		logger.WithError(err).Error("failed to connect to FTP server")
		return err
	}
	defer conn.Stop()

	useCaseRepos := &useCase.StatusRepos{
		Logger:     logger,
		Connection: conn,
	}

	useCaseInput := &useCase.StatusInput{}

	status, err := deps.UseCase.Execute(ctx, useCaseRepos, useCaseInput)
	if err != nil {
		msg := "failed to get server status"
		logger.WithError(err).Error(msg)
		return errors.NewInternalError(msg, err)
	}

	table := tablewriter.NewWriter(os.Stdout)
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
