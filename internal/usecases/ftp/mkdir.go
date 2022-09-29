package ftp

import (
	"context"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

type MkdirUseCase interface {
	Execute(ctx context.Context, repos *MkdirRepos, input *MkdirInput) error
}

type MkdirInput struct {
	Path string
}

type MkdirRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
}

type Mkdir struct {
}

func (u *Mkdir) Execute(_ context.Context, repos *MkdirRepos, input *MkdirInput) error {
	if err := repos.Connection.Mkdir(input.Path); err != nil {
		repos.Logger.WithError(err).Error("failed to create directory")
		return ftperrors.NewInternalError("failed to create directory", nil)
	}
	return nil
}
