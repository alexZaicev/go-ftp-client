package ftp

import (
	"context"
	"fmt"
	"sort"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

type ListFilesUseCase interface {
	Execute(context.Context, *ListFilesRepos, *ListFilesInput) ([]*entities.Entry, error)
}

type ListFilesInput struct {
	ShowAll  bool
	Path     string
	SortType entities.SortType
}

type ListFilesRepos struct {
	Logger     logging.Logger
	Connection connection.Connection
}

type ListFiles struct {
}

func (u *ListFiles) Execute(_ context.Context, repos *ListFilesRepos, input *ListFilesInput) ([]*entities.Entry, error) {
	listOptions := &connection.ListOptions{
		Path:    input.Path,
		ShowAll: input.ShowAll,
	}
	entries, err := repos.Connection.List(listOptions)
	if err != nil {
		repos.Logger.WithError(err).Error("failed to list files")
		return nil, errors.NewInternalError("failed to list files", nil)
	}

	if len(entries) == 0 {
		return nil, errors.NewNotFoundError(
			fmt.Sprintf("no entries found under %s path", input.Path),
			nil,
		)
	}

	sort.Slice(entries, func(i, j int) bool {
		switch input.SortType {
		case entities.SortTypeName:
			return entries[i].Name < entries[j].Name
		case entities.SortTypeSize:
			return entries[i].SizeInBytes < entries[j].SizeInBytes
		case entities.SortTypeDate:
			return entries[i].LastModificationDate.UnixMilli() < entries[j].LastModificationDate.UnixMilli()
		default:
			return false
		}
	})

	return entries, nil
}
