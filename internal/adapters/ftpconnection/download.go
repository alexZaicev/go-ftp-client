package ftpconnection

import (
	"context"
	"io"

	"github.com/hashicorp/go-multierror"

	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Download(ctx context.Context, path string) ([]byte, error) {
	if path == "" {
		return nil, ftperrors.NewInvalidArgumentError("path", ftperrors.ErrMsgCannotBeBlank)
	}

	conn, err := c.cmdWithDataConn(ctx, 0, CommandRetrieve, path)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to open data transfer connection", err)
	}

	var multiErr *multierror.Error

	data, err := io.ReadAll(conn)
	if err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	// opened data connection needs to be closed prior to the bellow check
	if closeErr := conn.Close(); closeErr != nil {
		multiErr = multierror.Append(multiErr, closeErr)
	}

	if shutErr := c.checkDataConnShut(); shutErr != nil {
		multiErr = multierror.Append(multiErr, shutErr)
	}

	err = multiErr.ErrorOrNil()
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to download file", err)
	}

	return data, nil
}
