package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/status"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

func AddStatusCommand(rootCMD *cobra.Command) error {
	statusCMD := &cobra.Command{
		Use:   "status",
		Short: "Returns information on the server status, including the status of the current connection.",
		RunE:  doStatus,
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

func doStatus(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	input, err := parseStatusFlags(cmd.Flags(), args)
	if err != nil {
		return err
	}

	logger, err := logging.NewZapJSONLogger(GetLogLevel(input.Verbose))
	if err != nil {
		return errors.NewInternalError("failed to setup logger", err)
	}

	dependencies := &status.Dependencies{
		UseCase: &ftp.Status{},
	}

	if err = status.PerformStatus(ctx, logger, dependencies, input); err != nil {
		return err
	}
	return nil
}
