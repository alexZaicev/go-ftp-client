package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/move"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

//nolint:dupl // similar to AddStatusCommand
func AddMoveCommand(rootCMD *cobra.Command) error {
	moveCMD := &cobra.Command{
		Use:   "mv",
		Short: "Move file or directory.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseMoveFlags(cmd.Flags(), args)
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

			dependencies := &move.Dependencies{
				Connector: ftpclient.NewConnector(),
				UseCase:   &ftp.Move{},
				OutWriter: cmd.OutOrStdout(),
			}

			err = move.PerformMove(ctx, logger, dependencies, input)
			return
		},
	}

	if err := setConnectionFlags(moveCMD); err != nil {
		return err
	}

	rootCMD.AddCommand(moveCMD)
	return nil
}

func parseMoveFlags(flagSet *pflag.FlagSet, args []string) (*move.CmdMoveInput, error) {
	config, err := parseConnectionFlags(flagSet)
	if err != nil {
		return nil, err
	}

	//nolint:gomnd // expecting 2 args for command
	if len(args) != 2 {
		return nil, ftperrors.NewInvalidArgumentError("args", "should contain valid from and to paths")
	}

	return &move.CmdMoveInput{
		Config:  config,
		OldPath: args[0],
		NewPath: args[1],
	}, nil
}
