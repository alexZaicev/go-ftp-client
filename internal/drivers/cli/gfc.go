package cli

import (
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/cli/models"
)

const (
	ArgAll  = "all"
	ArgSort = "sort"

	defaultConnectionTimeout = 5 * time.Second
	defaultUserAccount       = "anonymous"
	defaultUserPassword      = "anonymous"
)

func NewGfcCommand() (*cobra.Command, error) {
	rootCMD := &cobra.Command{
		Use:   "gfc",
		Short: "A CLI utility to manage files on FTP server",
	}

	if err := AddStatusCommand(rootCMD); err != nil {
		return nil, ftperrors.NewInternalError("failed to setup status command", err)
	}
	if err := AddListCommand(rootCMD); err != nil {
		return nil, ftperrors.NewInternalError("failed to setup list command", err)
	}
	if err := AddUploadCommand(rootCMD); err != nil {
		return nil, ftperrors.NewInternalError("failed to setup upload command", err)
	}
	if err := AddMkdirCommand(rootCMD); err != nil {
		return nil, ftperrors.NewInternalError("failed to setup mkdir command", err)
	}
	if err := AddRemoveCommand(rootCMD); err != nil {
		return nil, ftperrors.NewInternalError("failed to setup remove command", err)
	}
	if err := AddMoveCommand(rootCMD); err != nil {
		return nil, ftperrors.NewInternalError("failed to setup move command", err)
	}
	if err := AddDownloadCommand(rootCMD); err != nil {
		return nil, ftperrors.NewInternalError("failed to setup download command", err)
	}

	return rootCMD, nil
}

func getLogLevel(verbose bool) string {
	if verbose {
		return "debug"
	}
	return "info"
}

func getFileAbsPath(filePath string) (string, error) {
	if filePath == "" {
		return "", nil
	}
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	return absFilePath, nil
}

func setConnectionFlags(cmd *cobra.Command) error {
	cmd.Flags().StringP(models.ArgAddress.Long, models.ArgAddress.Short, "", models.ArgAddress.Help)
	if err := cmd.MarkFlagRequired(models.ArgAddress.Long); err != nil {
		return err
	}

	cmd.Flags().StringP(models.ArgUser.Long, models.ArgUser.Short, defaultUserAccount, models.ArgUser.Help)
	cmd.Flags().StringP(models.ArgPassword.Long, models.ArgPassword.Short, defaultUserPassword, models.ArgPassword.Help)

	cmd.Flags().BoolP(models.ArgVerbose.Long, models.ArgVerbose.Short, false, models.ArgVerbose.Help)

	cmd.Flags().String(models.ArgTLSCertFilePath.Long, "", models.ArgTLSCertFilePath.Help)
	cmd.Flags().String(models.ArgTLSKeyFilePath.Long, "", models.ArgTLSKeyFilePath.Help)
	cmd.Flags().Bool(models.ArgTLSInsecure.Long, false, models.ArgTLSInsecure.Help)

	return nil
}

func parseConnectionFlags(flagSet *pflag.FlagSet) (ftpclient.ConnectorConfig, error) {
	address, err := flagSet.GetString(models.ArgAddress.Long)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}

	user, err := flagSet.GetString(models.ArgUser.Long)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}

	pwd, err := flagSet.GetString(models.ArgPassword.Long)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}

	verbose, err := flagSet.GetBool(models.ArgVerbose.Long)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}

	certFilePath, err := flagSet.GetString(models.ArgTLSCertFilePath.Long)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}
	certFilePath, err = getFileAbsPath(certFilePath)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}

	keyFilePath, err := flagSet.GetString(models.ArgTLSKeyFilePath.Long)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}
	keyFilePath, err = getFileAbsPath(keyFilePath)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}

	insecure, err := flagSet.GetBool(models.ArgTLSInsecure.Long)
	if err != nil {
		return ftpclient.ConnectorConfig{}, err
	}

	return ftpclient.ConnectorConfig{
		Address:         address,
		User:            user,
		Password:        pwd,
		Verbose:         verbose,
		Timeout:         defaultConnectionTimeout,
		TLSCertFilePath: certFilePath,
		TLSKeyFilePath:  keyFilePath,
		TLSInsecure:     insecure,
	}, nil
}
