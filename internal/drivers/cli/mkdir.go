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

// nolint:dupl // similar to AddStatusCommand
func AddMkdirCommand(rootCMD *cobra.Command) error {
	// nolint:dupl // single use case command are very similar
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
				getLogLevel(input.Verbose),
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

	mkdirCMD.Flags().StringP(ArgAddress, ArgAddressShort, "", "Connection address for the FTP server (e.g. ftp.example.com:21)")
	if err := mkdirCMD.MarkFlagRequired(ArgAddress); err != nil {
		return err
	}

	mkdirCMD.Flags().StringP(ArgUser, ArgUserShort, defaultUserAccount, "Username for the FTP server user")
	mkdirCMD.Flags().StringP(ArgPassword, ArgPasswordShort, defaultUserPassword, "Password for the FTP server user")

	mkdirCMD.Flags().BoolP(ArgVerbose, ArgVerboseShort, false, "Verbose output")

	rootCMD.AddCommand(mkdirCMD)
	return nil
}

func parseMkdirFlags(flagSet *pflag.FlagSet, args []string) (*mkdir.CmdMkdirInput, error) {
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

	if len(args) != 1 {
		return nil, ftperrors.NewInvalidArgumentError("args", "should contain exactly one valid path")
	}

	return &mkdir.CmdMkdirInput{
		Address:  address,
		User:     user,
		Password: pwd,
		Verbose:  verbose,
		Timeout:  defaultConnectionTimeout,
		Path:     args[0],
	}, nil
}
