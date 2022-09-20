package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/list"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/cli/models"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

func AddListCommand(rootCMD *cobra.Command) error {
	listCMD := &cobra.Command{
		Use:   "ls",
		Short: "List files in directory.",
		RunE:  doList,
	}

	listCMD.Flags().StringP(ArgAddress, ArgAddressShort, "", "Connection address for the FTP server (e.g. ftp.example.com:21)")
	if err := listCMD.MarkFlagRequired(ArgAddress); err != nil {
		return err
	}

	listCMD.Flags().StringP(ArgUser, ArgUserShort, defaultUserAccount, "Username for the FTP server user")
	listCMD.Flags().StringP(ArgPassword, ArgPasswordShort, defaultUserPassword, "Password for the FTP server user")

	listCMD.Flags().BoolP(ArgVerbose, ArgVerboseShort, false, "Verbose output")

	listCMD.Flags().String(ArgSort, string(models.SortTypeName), "Sort returned entries by NAME/SIZE/DATE")

	listCMD.Flags().Bool(ArgAll, false, "Do not ignore entries starting with '.'")

	rootCMD.AddCommand(listCMD)
	return nil
}

func parseListFlags(flagSet *pflag.FlagSet, args []string) (*list.CmdListInput, error) {
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

	showAll, err := flagSet.GetBool(ArgAll)
	if err != nil {
		return nil, err
	}

	sortTypeStr, err := flagSet.GetString(ArgSort)
	if err != nil {
		return nil, err
	}

	if len(args) > 1 {
		return nil, errors.NewInvalidArgumentError("args", "cannot contain more than one path")
	}
	// if no args provided, set path to list current working directory
	if len(args) == 0 {
		args = append(args, "./")
	}

	return &list.CmdListInput{
		Address:  address,
		User:     user,
		Password: pwd,
		Verbose:  verbose,
		Timeout:  defaultConnectionTimeout,
		ShowAll:  showAll,
		Path:     args[0],
		SortType: models.SortType(sortTypeStr),
	}, nil
}

func doList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	input, err := parseListFlags(cmd.Flags(), args)
	if err != nil {
		return err
	}

	logger, err := logging.NewZapJSONLogger(GetLogLevel(input.Verbose))
	if err != nil {
		return errors.NewInternalError("failed to setup logger", err)
	}

	dependencies := &list.Dependencies{
		UseCase: &ftp.ListFiles{},
	}

	if err := list.PerformListFiles(ctx, logger, dependencies, input); err != nil {
		return err
	}

	return nil
}
