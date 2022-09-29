package ftpconnection

import (
	"bufio"
	"context"

	"github.com/hashicorp/go-multierror"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/parsers"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) List(ctx context.Context, options *connection.ListOptions) (entries []*entities.Entry, err error) {
	if options == nil {
		return nil, ftperrors.NewInvalidArgumentError("options", ftperrors.ErrMsgCannotBeNil)
	}

	cmd := CommandList
	if c.features.supportMLST {
		cmd = CommandListMachineReadable
	} else if options.ShowAll {
		cmd = CommandListHidden
	}

	conn, err := c.cmdWithDataConn(ctx, 0, cmd, options.Path)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to list files", err)
	}

	var multiErr *multierror.Error

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		entryStr := scanner.Text()
		entry, parseErr := c.parser.Parse(entryStr, &parsers.Options{
			Location: c.location,
		})
		if parseErr != nil {
			multiErr = multierror.Append(multiErr, parseErr)
			break
		}
		entries = append(entries, entry)
	}

	if scanErr := scanner.Err(); scanErr != nil {
		multiErr = multierror.Append(multiErr, scanErr)
	}

	if closeErr := conn.Close(); closeErr != nil {
		multiErr = multierror.Append(multiErr, closeErr)
	}

	if checkShutErr := c.checkDataConnShut(); checkShutErr != nil {
		multiErr = multierror.Append(multiErr, checkShutErr)
	}

	err = multiErr.ErrorOrNil()
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to list files", err)
	}

	return entries, nil
}
