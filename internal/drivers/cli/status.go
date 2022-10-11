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

func AddStatusCommand(rootCMD *cobra.Command) error {
	// nolint:dupl // single use case command are very similar
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
				getLogLevel(input.Verbose),
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

	statusCMD.Flags().StringP(ArgAddress, ArgAddressShort, "", "Connection address for the FTP server (e.g. ftp.example.com:21)")
	if err := statusCMD.MarkFlagRequired(ArgAddress); err != nil {
		return err
	}

	statusCMD.Flags().StringP(ArgUser, ArgUserShort, defaultUserAccount, "Username for the FTP server user")
	statusCMD.Flags().StringP(ArgPassword, ArgPasswordShort, defaultUserPassword, "Password for the FTP server user")

	statusCMD.Flags().BoolP(ArgVerbose, ArgVerboseShort, false, "Verbose output")

	rootCMD.AddCommand(statusCMD)
	return nil
}

func parseStatusFlags(flagSet *pflag.FlagSet, _ []string) (*status.CmdStatusInput, error) {
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

	return &status.CmdStatusInput{
		Address:  address,
		User:     user,
		Password: pwd,
		Verbose:  verbose,
		Timeout:  defaultConnectionTimeout,
	}, nil
}
