package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/upload"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/cli/models"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

func AddUploadCommand(rootCMD *cobra.Command) error {
	uploadCMD := &cobra.Command{
		Use:   "upload",
		Short: "Upload file(s) to the server.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := context.Background()

			input, err := parseUploadFlags(cmd.Flags(), args)
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

			filesystem := os.DirFS("/")

			dependencies := &upload.Dependencies{
				MkdirUseCase:  &ftp.Mkdir{},
				UploadUseCase: &ftp.UploadFile{},
				Connector:     ftpclient.NewConnector(),
				Filesystem:    filesystem,
			}

			err = upload.PerformUploadFile(ctx, logger, dependencies, input)
			return
		},
	}

	if err := setConnectionFlags(uploadCMD); err != nil {
		return err
	}

	uploadCMD.Flags().BoolP(
		models.ArgRecursive.Long,
		models.ArgRecursive.Short,
		false,
		"Recursively upload directory tree",
	)

	rootCMD.AddCommand(uploadCMD)
	return nil
}

func parseUploadFlags(flagSet *pflag.FlagSet, args []string) (*upload.CmdUploadInput, error) {
	config, err := parseConnectionFlags(flagSet)
	if err != nil {
		return nil, err
	}

	recursive, err := flagSet.GetBool(models.ArgRecursive.Long)
	if err != nil {
		return nil, err
	}

	//nolint:gomnd // expecting 2 args for command
	if len(args) != 2 {
		return nil, ftperrors.NewInvalidArgumentError("args", "should contain valid path to file and remote path")
	}

	filePath, err := getFileAbsPath(args[0])
	if err != nil {
		return nil, ftperrors.NewInvalidArgumentError("args", err.Error())
	}

	return &upload.CmdUploadInput{
		Config:         config,
		FilePath:       filePath,
		Recursive:      recursive,
		RemoteFilePath: args[1],
	}, nil
}
