package ftp

import (
	"context"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

type MoveUseCase interface {
	Execute(context.Context, *MoveRepos, *MoveInput) error
}

type MoveInput struct {
	OldPath string
	NewPath string
}

type MoveRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
}

type Move struct {
}

func (u *Move) Execute(ctx context.Context, repos *MoveRepos, input *MoveInput) error {
	if err := repos.Connection.Move(input.OldPath, input.NewPath); err != nil {
		repos.Logger.WithError(err).Error("failed to move file")
		return ftperrors.NewInternalError("failed to move file", nil)
	}

	return nil
}
