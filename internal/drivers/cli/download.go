package cli

import (
	"context"
	"path/filepath"

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
	removeCMD := &cobra.Command{
		Use:   "download",
		Short: "Download file(s) from the server.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseDownloadFlags(cmd.Flags(), args)
			if err != nil {
				return err
			}

			logger, err := logging.NewZapJSONLogger(
				getLogLevel(input.Verbose),
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

	removeCMD.Flags().StringP(ArgAddress, ArgAddressShort, "", "Connection address for the FTP server (e.g. ftp.example.com:21)")
	if err := removeCMD.MarkFlagRequired(ArgAddress); err != nil {
		return err
	}

	removeCMD.Flags().StringP(ArgUser, ArgUserShort, defaultUserAccount, "Username for the FTP server user")
	removeCMD.Flags().StringP(ArgPassword, ArgPasswordShort, defaultUserPassword, "Password for the FTP server user")

	removeCMD.Flags().BoolP(ArgVerbose, ArgVerboseShort, false, "Verbose output")

	rootCMD.AddCommand(removeCMD)
	return nil
}

func parseDownloadFlags(flagSet *pflag.FlagSet, args []string) (*download.CmdDownloadInput, error) {
	address, err := flagSet.GetString(ArgAddress)
	if err != nil {
		return nil, err
	}

	user, err := flagSet.GetString(ArgUser)
	if err != nil {
		return nil, err
	}

	pwd, err := flagSet.GetString(ArgPassword)
	if err != nil {
		return nil, err
	}

	verbose, err := flagSet.GetBool(ArgVerbose)
	if err != nil {
		return nil, err
	}

	// nolint:gomnd // expecting 2 args for command
	if len(args) != 2 {
		return nil, ftperrors.NewInvalidArgumentError("args", "should contain valid remote file and download paths")
	}

	filePath, err := filepath.Abs(args[1])
	if err != nil {
		return nil, ftperrors.NewInvalidArgumentError("args", err.Error())
	}

	return &download.CmdDownloadInput{
		Address:    address,
		User:       user,
		Password:   pwd,
		Verbose:    verbose,
		Timeout:    defaultConnectionTimeout,
		RemotePath: args[0],
		Path:       filePath,
	}, nil
}
