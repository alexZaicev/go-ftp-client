package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	ArgAddress        = "address"
	ArgAddressShort   = "a"
	ArgUser           = "user"
	ArgUserShort      = "u"
	ArgPassword       = "password"
	ArgPasswordShort  = "p"
	ArgFile           = "file"
	ArgFileShort      = "f"
	ArgVerbose        = "verbose"
	ArgVerboseShort   = "v"
	ArgRecursive      = "recursive"
	ArgRecursiveShort = "r"

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
		return nil, errors.NewInternalError("failed to setup status command", err)
	}
	if err := AddListCommand(rootCMD); err != nil {
		return nil, errors.NewInternalError("failed to setup list command", err)
	}
	if err := AddUploadCommand(rootCMD); err != nil {
		return nil, errors.NewInternalError("failed to setup upload command", err)
	}
	if err := AddMkdirCommand(rootCMD); err != nil {
		return nil, errors.NewInternalError("failed to setup mkdir command", err)
	}
	if err := AddRemoveCommand(rootCMD); err != nil {
		return nil, errors.NewInternalError("failed to setup remove command", err)
	}
	if err := AddMoveCommand(rootCMD); err != nil {
		return nil, errors.NewInternalError("failed to setup move command", err)
	}
	if err := AddDownloadCommand(rootCMD); err != nil {
		return nil, errors.NewInternalError("failed to setup download command", err)
	}

	return rootCMD, nil
}

func getLogLevel(verbose bool) string {
	if verbose {
		return "debug"
	}
	return "info"
}
