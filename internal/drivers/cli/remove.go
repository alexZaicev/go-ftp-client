package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/remove"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

//nolint:dupl // similar to AddStatusCommand
func AddRemoveCommand(rootCMD *cobra.Command) error {
	removeCMD := &cobra.Command{
		Use:   "rm",
		Short: "Remove file or directory.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseRemoveFlags(cmd.Flags(), args)
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

			dependencies := &remove.Dependencies{
				Connector: ftpclient.NewConnector(),
				UseCase:   &ftp.Remove{},
				OutWriter: cmd.OutOrStdout(),
			}

			err = remove.PerformRemove(ctx, logger, dependencies, input)
			return
		},
	}

	if err := setConnectionFlags(removeCMD); err != nil {
		return err
	}

	rootCMD.AddCommand(removeCMD)
	return nil
}

func parseRemoveFlags(flagSet *pflag.FlagSet, args []string) (*remove.CmdRemoveInput, error) {
	config, err := parseConnectionFlags(flagSet)
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, ftperrors.NewInvalidArgumentError("args", "should contain exactly one valid path")
	}

	return &remove.CmdRemoveInput{
		Config: config,
		Path:   args[0],
	}, nil
}
