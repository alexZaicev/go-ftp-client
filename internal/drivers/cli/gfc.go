package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	ArgAddress            = "address"
	ArgAddressShort       = "a"
	ArgUser               = "user"
	ArgUserShort          = "u"
	ArgPassword           = "password"
	ArgPasswordShort      = "p"
	ArgFile               = "file"
	ArgFileShort          = "f"
	ArgCreateParents      = "create-parents"
	ArgCreateParentsShort = "c"
	ArgVerbose            = "verbose"
	ArgVerboseShort       = "v"
	ArgRecursive          = "recursive"
	ArgRecursiveShort     = "r"

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

	return rootCMD, nil
}

func GetLogLevel(verbose bool) string {
	if verbose {
		return "debug"
	}
	return "info"
}
