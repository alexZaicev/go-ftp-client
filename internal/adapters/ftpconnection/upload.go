package ftpconnection

import (
	"context"
	"io"

	"github.com/hashicorp/go-multierror"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Upload(ctx context.Context, options *connection.UploadOptions) error {
	if options == nil {
		return ftperrors.NewInvalidArgumentError("options", ftperrors.ErrMsgCannotBeNil)
	}

	conn, err := c.cmdWithDataConn(ctx, 0, models.CommandStore, options.Path)
	if err != nil {
		return ftperrors.NewInternalError("failed to open data transfer connection", err)
	}

	var multiErr *multierror.Error

	if _, err = io.Copy(conn, options.FileReader); err != nil {
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
		return ftperrors.NewInternalError("failed to upload file", err)
	}

	return nil
}
