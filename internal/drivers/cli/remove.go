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

// nolint:dupl // similar to AddStatusCommand
func AddRemoveCommand(rootCMD *cobra.Command) error {
	removeCMD := &cobra.Command{
		Use:   "rm",
		Short: "Remove file or directory",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseRemoveFlags(cmd.Flags(), args)
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

			dependencies := &remove.Dependencies{
				Connector: ftpclient.NewConnector(),
				UseCase:   &ftp.Remove{},
				OutWriter: cmd.OutOrStdout(),
			}

			err = remove.PerformRemove(ctx, logger, dependencies, input)
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

func parseRemoveFlags(flagSet *pflag.FlagSet, args []string) (*remove.CmdRemoveInput, error) {
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

	recursive, err := flagSet.GetBool(ArgRecursive)
	if err != nil {
		return nil, err
	}

	if len(args) > 1 {
		return nil, ftperrors.NewInvalidArgumentError("args", "cannot contain more than one path")
	}
	// if no args provided, set path to list current working directory
	if len(args) == 0 {
		args = append(args, "./")
	}

	return &remove.CmdRemoveInput{
		Address:   address,
		User:      user,
		Password:  pwd,
		Verbose:   verbose,
		Timeout:   defaultConnectionTimeout,
		Path:      args[0],
		Recursive: recursive,
	}, nil
}
