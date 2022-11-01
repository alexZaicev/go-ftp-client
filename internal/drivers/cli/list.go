package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/list"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/cli/models"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

func AddListCommand(rootCMD *cobra.Command) error {
	//nolint:dupl // single use case command are very similar
	listCMD := &cobra.Command{
		Use:   "ls",
		Short: "List files in directory.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseListFlags(cmd.Flags(), args)
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

			dependencies := &list.Dependencies{
				Connector: ftpclient.NewConnector(),
				UseCase:   &ftp.ListFiles{},
				OutWriter: cmd.OutOrStdout(),
			}

			err = list.PerformListFiles(ctx, logger, dependencies, input)
			return
		},
	}

	if err := setConnectionFlags(listCMD); err != nil {
		return err
	}

	listCMD.Flags().String(ArgSort, string(models.SortTypeName), "Sort returned entries by NAME/SIZE/DATE")

	listCMD.Flags().Bool(ArgAll, false, "Do not ignore entries starting with '.'")

	rootCMD.AddCommand(listCMD)
	return nil
}

func parseListFlags(flagSet *pflag.FlagSet, args []string) (*list.CmdListInput, error) {
	config, err := parseConnectionFlags(flagSet)
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
		return nil, ftperrors.NewInvalidArgumentError(
			"args",
			"should be empty or contain exactly one valid path",
		)
	}

	// if no args provided, set path to list current working directory
	if len(args) == 0 {
		args = append(args, "./")
	}

	return &list.CmdListInput{
		Config:   config,
		ShowAll:  showAll,
		Path:     args[0],
		SortType: models.SortType(sortTypeStr),
	}, nil
}
