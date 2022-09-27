package list

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	ftpErrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/cli/models"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	useCase "github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

type CmdListInput struct {
	Address   string
	User      string
	Password  string
	Verbose   bool
	Timeout   time.Duration
	OutWriter io.Writer

	ShowAll  bool
	Path     string
	SortType models.SortType
}

type Dependencies struct {
	Connector ftpclient.Connector
	UseCase   useCase.ListFilesUseCase
}

func PerformListFiles(ctx context.Context, logger logging.Logger, deps *Dependencies, input *CmdListInput) (err error) {
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
	defer func() {
		if stopErr := conn.Stop(); stopErr != nil {
			logger.WithError(stopErr).Error("failed to stop server connection")
			err = stopErr
		}
	}()

	useCaseRepos := &useCase.ListFilesRepos{
		Logger:     logger,
		Connection: conn,
	}

	sortType, err := models.SortTypeToDomain(input.SortType)
	if err != nil {
		return err
	}

	useCaseInput := &useCase.ListFilesInput{
		Path:     input.Path,
		ShowAll:  input.ShowAll,
		SortType: sortType,
	}

	entries, err := deps.UseCase.Execute(ctx, useCaseRepos, useCaseInput)
	if err != nil {
		var notFoundErr *ftpErrors.NotFoundError
		if errors.As(err, &notFoundErr) {
			logger.Info("no entries found under specified path")
			return nil
		}

		return err
	}

	table := tablewriter.NewWriter(input.OutWriter)
	table.SetHeader([]string{"type", "permissions", "owners", "name", "last modified", "size"})
	for _, entry := range entries {
		entryType, cnvErr := ftpclient.EntryTypeToStr(entry.Type)
		if cnvErr != nil {
			return cnvErr
		}

		table.Append([]string{
			entryType,
			entry.Permissions,
			fmt.Sprintf("%s:%s", entry.OwnerUser, entry.OwnerGroup),
			entry.Name,
			entry.LastModificationDate.Format(ftpclient.DateFormat),
			ftpclient.FormatSizeInBytes(entry.SizeInBytes),
		})
	}
	table.Render()

	return nil
}
