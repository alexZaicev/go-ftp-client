package models

import (
	"fmt"

	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftpErrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

type SortType string

const (
	SortTypeName SortType = "NAME"
	SortTypeSize SortType = "SIZE"
	SortTypeDate SortType = "DATE"
)

func SortTypeToDomain(sortType SortType) (entities.SortType, error) {
	switch sortType {
	case SortTypeName:
		return entities.SortTypeName, nil
	case SortTypeSize:
		return entities.SortTypeSize, nil
	case SortTypeDate:
		return entities.SortTypeDate, nil
	default:
		return entities.SortType(0), ftpErrors.NewUnknownError(
			fmt.Sprintf("unexpected sort type: %s", sortType),
			nil,
		)
	}
}
