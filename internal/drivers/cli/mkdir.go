package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/mkdir"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

//nolint:dupl // similar to AddStatusCommand
func AddMkdirCommand(rootCMD *cobra.Command) error {
	//nolint:dupl // single use case command are very similar
	mkdirCMD := &cobra.Command{
		Use:   "mkdir",
		Short: "Create directory(ies).",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseMkdirFlags(cmd.Flags(), args)
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

			dependencies := &mkdir.Dependencies{
				Connector: ftpclient.NewConnector(),
				UseCase:   &ftp.Mkdir{},
				OutWriter: cmd.OutOrStdout(),
			}

			err = mkdir.PerformMkdir(ctx, logger, dependencies, input)
			return
		},
	}

	if err := setConnectionFlags(mkdirCMD); err != nil {
		return err
	}

	rootCMD.AddCommand(mkdirCMD)
	return nil
}

func parseMkdirFlags(flagSet *pflag.FlagSet, args []string) (*mkdir.CmdMkdirInput, error) {
	config, err := parseConnectionFlags(flagSet)
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, ftperrors.NewInvalidArgumentError("args", "should contain exactly one valid path")
	}

	return &mkdir.CmdMkdirInput{
		Config: config,
		Path:   args[0],
	}, nil
}
