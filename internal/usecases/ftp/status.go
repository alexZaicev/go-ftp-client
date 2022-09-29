package ftp

import (
	"context"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

type StatusUseCase interface {
	Execute(context.Context, *StatusRepos, *StatusInput) (*entities.Status, error)
}

type StatusInput struct {
}

type StatusRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
}

type Status struct {
}

func (u *Status) Execute(_ context.Context, repos *StatusRepos, _ *StatusInput) (*entities.Status, error) {
	status, err := repos.Connection.Status()
	if err != nil {
		repos.Logger.WithError(err).Error("failed to get server status")
		return nil, errors.NewInternalError("failed to get server status", nil)
	}
	return status, nil
}
