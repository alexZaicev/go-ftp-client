package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/upload"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
)

func AddUploadCommand(rootCMD *cobra.Command) error {
	uploadCMD := &cobra.Command{
		Use:   "upload",
		Short: "Upload file(s) to the server",
		RunE:  doUpload,
	}

	uploadCMD.Flags().StringP(ArgAddress, ArgAddressShort, "", "Connection address for the FTP server (e.g. ftp.example.com:21)")
	if err := uploadCMD.MarkFlagRequired(ArgAddress); err != nil {
		return err
	}

	uploadCMD.Flags().StringP(ArgUser, ArgUserShort, defaultUserAccount, "Username for the FTP server user")
	uploadCMD.Flags().StringP(ArgPassword, ArgPasswordShort, defaultUserPassword, "Password for the FTP server user")

	uploadCMD.Flags().BoolP(ArgVerbose, ArgVerboseShort, false, "Verbose output")

	uploadCMD.Flags().StringP(ArgFile, ArgFileShort, "", "Path to file for upload")
	if err := uploadCMD.MarkFlagRequired(ArgFile); err != nil {
		return err
	}

	uploadCMD.Flags().BoolP(ArgRecursive, ArgRecursiveShort, false, "Recursively upload directory tree")

	rootCMD.AddCommand(uploadCMD)
	return nil
}

func parseUploadFlags(flagSet *pflag.FlagSet, args []string) (*upload.CmdUploadInput, error) {
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

	filePath, err := flagSet.GetString(ArgFile)
	if err != nil {
		return nil, err
	}
	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	recursive, err := flagSet.GetBool(ArgRecursive)
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

	return &upload.CmdUploadInput{
		Address:        address,
		User:           user,
		Password:       pwd,
		Verbose:        verbose,
		Timeout:        defaultConnectionTimeout,
		FilePath:       filePath,
		Recursive:      recursive,
		RemoteFilePath: args[0],
	}, nil
}

func doUpload(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	input, err := parseUploadFlags(cmd.Flags(), args)
	if err != nil {
		return err
	}

	logger, err := logging.NewZapJSONLogger(GetLogLevel(input.Verbose))
	if err != nil {
		return errors.NewInternalError("failed to setup logger", err)
	}

	filesystem := os.DirFS("/")

	dependencies := &upload.Dependencies{
		MkdirUseCase:  &ftp.Mkdir{},
		UploadUseCase: &ftp.UploadFile{},
		Filesystem:    filesystem,
	}

	if err := upload.PerformUploadFile(ctx, logger, dependencies, input); err != nil {
		return err
	}

	return nil
}
