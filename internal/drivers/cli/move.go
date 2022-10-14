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

// nolint:dupl // similar to AddStatusCommand
func AddMoveCommand(rootCMD *cobra.Command) error {
	removeCMD := &cobra.Command{
		Use:   "mv",
		Short: "Move file or directory",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseMoveFlags(cmd.Flags(), args)
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

			dependencies := &move.Dependencies{
				Connector: ftpclient.NewConnector(),
				UseCase:   &ftp.Move{},
				OutWriter: cmd.OutOrStdout(),
			}

			err = move.PerformMove(ctx, logger, dependencies, input)
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

	removeCMD.Flags().BoolP(ArgRecursive, ArgRecursiveShort, false, "Recursive")

	rootCMD.AddCommand(removeCMD)
	return nil
}

func parseMoveFlags(flagSet *pflag.FlagSet, args []string) (*move.CmdMoveInput, error) {
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
		return nil, ftperrors.NewInvalidArgumentError(
			"args",
			"should contain a valid path to file/directory and a new name (e.g. '/foo/bar baz')",
		)
	}

	return &move.CmdMoveInput{
		Address:  address,
		User:     user,
		Password: pwd,
		Verbose:  verbose,
		Timeout:  defaultConnectionTimeout,
		OldPath:  args[0],
		NewPath:  args[1],
	}, nil
}
