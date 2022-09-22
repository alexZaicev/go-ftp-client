package ftp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
	connectionMocks "github.com/alexZaicev/go-ftp-client/mocks/domain/connection"
)

func Test_ListFiles_Execute_Success(t *testing.T) {
	testCases := []struct {
		name            string
		sortType        entities.SortType
		entries         []*entities.Entry
		expectedEntries []*entities.Entry
	}{
		{
			name:     "list sort by name",
			sortType: entities.SortTypeName,
			entries:  getEntries(t),
			expectedEntries: []*entities.Entry{
				newEntry(t, "file1", 100, "2022-01-02 15:23"),
				newEntry(t, "file2", 102, "2021-05-02 17:23"),
				newEntry(t, "file3", 40032, "2022-01-24 16:23"),
				newEntry(t, "file4", 5043, "2022-01-12 19:23"),
				newEntry(t, "file5", 167, "2022-01-12 16:23"),
				newEntry(t, "file6", 9635, "2022-01-02 13:23"),
				newEntry(t, "file7", 4352, "2020-04-02 14:23"),
				newEntry(t, "file8", 1034, "2022-09-02 10:23"),
				newEntry(t, "file9", 2, "2022-01-02 11:23"),
			},
		},
		{
			name:     "list sort by size",
			sortType: entities.SortTypeSize,
			entries:  getEntries(t),
			expectedEntries: []*entities.Entry{
				newEntry(t, "file9", 2, "2022-01-02 11:23"),
				newEntry(t, "file1", 100, "2022-01-02 15:23"),
				newEntry(t, "file2", 102, "2021-05-02 17:23"),
				newEntry(t, "file5", 167, "2022-01-12 16:23"),
				newEntry(t, "file8", 1034, "2022-09-02 10:23"),
				newEntry(t, "file7", 4352, "2020-04-02 14:23"),
				newEntry(t, "file4", 5043, "2022-01-12 19:23"),
				newEntry(t, "file6", 9635, "2022-01-02 13:23"),
				newEntry(t, "file3", 40032, "2022-01-24 16:23"),
			},
		},
		{
			name:     "list sort by date",
			sortType: entities.SortTypeDate,
			entries:  getEntries(t),
			expectedEntries: []*entities.Entry{
				newEntry(t, "file7", 4352, "2020-04-02 14:23"),
				newEntry(t, "file2", 102, "2021-05-02 17:23"),
				newEntry(t, "file9", 2, "2022-01-02 11:23"),
				newEntry(t, "file6", 9635, "2022-01-02 13:23"),
				newEntry(t, "file1", 100, "2022-01-02 15:23"),
				newEntry(t, "file5", 167, "2022-01-12 16:23"),
				newEntry(t, "file4", 5043, "2022-01-12 19:23"),
				newEntry(t, "file3", 40032, "2022-01-24 16:23"),
				newEntry(t, "file8", 1034, "2022-09-02 10:23"),
			},
		},
		{
			name:            "list with unknown sort type",
			sortType:        entities.SortType(0),
			entries:         getEntries(t),
			expectedEntries: getEntries(t),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			logger := assertlogging.NewLogger(t)

			connMock := connectionMocks.NewConnection(t)
			connMock.
				On(
					"List",
					ctx,
					&connection.ListOptions{
						Path:    dirPath,
						ShowAll: true,
					}).
				Return(tc.entries, nil).
				Once()

			useCaseRepos := &ftp.ListFilesRepos{
				Logger:     logger,
				Connection: connMock,
			}
			useCaseInput := &ftp.ListFilesInput{
				Path:     dirPath,
				ShowAll:  true,
				SortType: tc.sortType,
			}

			useCase := &ftp.ListFiles{}
			entries, err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
			assert.NoError(t, err)
			if assert.Len(t, entries, len(tc.expectedEntries)) {
				for idx, ee := range tc.expectedEntries {
					assert.Equal(t, ee, entries[idx])
				}
			}
		})
	}
}

func Test_ListFiles_Execute_ListError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to list files").
		WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On(
			"List",
			ctx,
			&connection.ListOptions{
				Path:    dirPath,
				ShowAll: true,
			}).
		Return(nil, errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.ListFilesRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.ListFilesInput{
		Path:     dirPath,
		ShowAll:  true,
		SortType: entities.SortTypeName,
	}

	useCase := &ftp.ListFiles{}
	entries, err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	assert.Nil(t, entries)
	require.EqualError(t, err, "an internal error occurred: failed to list files")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_ListFiles_Execute_NotFoundError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On(
			"List",
			ctx,
			&connection.ListOptions{
				Path:    dirPath,
				ShowAll: true,
			}).
		Return(nil, nil).
		Once()

	useCaseRepos := &ftp.ListFilesRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.ListFilesInput{
		Path:     dirPath,
		ShowAll:  true,
		SortType: entities.SortTypeName,
	}

	useCase := &ftp.ListFiles{}
	entries, err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	assert.Nil(t, entries)
	require.EqualError(t, err, fmt.Sprintf("not found error: no entries found under %s path", dirPath))
	assert.IsType(t, ftperrors.NotFoundErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}
