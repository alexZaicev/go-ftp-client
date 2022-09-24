package ftpconnection

import (
	"context"
	"io"

	"github.com/hashicorp/go-multierror"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Upload(ctx context.Context, options *connection.UploadOptions) error {
	if options == nil {
		return ftperrors.NewInvalidArgumentError("options", ftperrors.ErrMsgCannotBeNil)
	}

	conn, err := c.cmdWithDataConn(ctx, 0, CommandStore, options.Path)
	if err != nil {
		return err
	}

	var multiErr *multierror.Error

	if _, err = io.Copy(conn, options.FileReader); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	// opened data connection needs to be closed prior to the bellow check
	if err = conn.Close(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	if err = c.checkDataConnShut(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	err = multiErr.ErrorOrNil()
	if err != nil {
		return ftperrors.NewInternalError("failed to upload file to remote path", err)
	}

	return nil
}
