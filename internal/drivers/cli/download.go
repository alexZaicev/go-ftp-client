package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/filestore"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/download"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

func AddDownloadCommand(rootCMD *cobra.Command) error {
	downloadCMD := &cobra.Command{
		Use:   "download",
		Short: "Download file(s) from the server.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseDownloadFlags(cmd.Flags(), args)
			if err != nil {
				return err
			}

			logger, err := logging.NewZapJSONLogger(
				getLogLevel(input.Config.Verbose),
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
			if err != nil {
				return ftperrors.NewInternalError("failed to setup logger", err)
			}

			dependencies := &download.Dependencies{
				Connector: ftpclient.NewConnector(),
				UseCase:   &ftp.Download{},
				FileStore: &filestore.FileStore{},
				OutWriter: cmd.OutOrStdout(),
			}

			err = download.PerformDownload(ctx, logger, dependencies, input)
			return
		},
	}

	if err := setConnectionFlags(downloadCMD); err != nil {
		return err
	}

	rootCMD.AddCommand(downloadCMD)
	return nil
}

func parseDownloadFlags(flagSet *pflag.FlagSet, args []string) (*download.CmdDownloadInput, error) {
	config, err := parseConnectionFlags(flagSet)
	if err != nil {
		return nil, err
	}

	//nolint:gomnd // expecting 2 args for command
	if len(args) != 2 {
		return nil, ftperrors.NewInvalidArgumentError("args", "should contain valid remote file and download paths")
	}

	filePath, err := getFileAbsPath(args[1])
	if err != nil {
		return nil, ftperrors.NewInvalidArgumentError("args", err.Error())
	}

	return &download.CmdDownloadInput{
		Config:     config,
		RemotePath: args[0],
		Path:       filePath,
	}, nil
}
