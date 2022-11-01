package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/status"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

//nolint:dupl // similar to AddMkdirCommand
func AddStatusCommand(rootCMD *cobra.Command) error {
	//nolint:dupl // single use case command are very similar
	statusCMD := &cobra.Command{
		Use:   "status",
		Short: "Returns information on the server status, including the status of the current connection.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseStatusFlags(cmd.Flags(), args)
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

			dependencies := &status.Dependencies{
				Connector: ftpclient.NewConnector(),
				UseCase:   &ftp.Status{},
				OutWriter: cmd.OutOrStdout(),
			}

			err = status.PerformStatus(ctx, logger, dependencies, input)
			return
		},
	}

	if err := setConnectionFlags(statusCMD); err != nil {
		return err
	}

	rootCMD.AddCommand(statusCMD)
	return nil
}

func parseStatusFlags(flagSet *pflag.FlagSet, _ []string) (*status.CmdStatusInput, error) {
	config, err := parseConnectionFlags(flagSet)
	if err != nil {
		return nil, err
	}

	return &status.CmdStatusInput{
		Config: config,
	}, nil
}
