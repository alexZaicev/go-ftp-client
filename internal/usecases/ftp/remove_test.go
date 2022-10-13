package ftp_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
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

const (
	path = "foo"
)

func Test_Remove_Execute_RemoveFile_Success(t *testing.T) {
	testCases := []struct {
		name      string
		recursive bool
	}{
		{
			name:      "remove file non-recursive",
			recursive: false,
		},
		{
			name:      "remove file recursive",
			recursive: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			ctx := context.Background()

			logger := assertlogging.NewLogger(t)

			entries := []*entities.Entry{
				newEntry(t, entities.EntryTypeFile, "file-1", 167, "2022-01-12 16:23"),
				newEntry(t, entities.EntryTypeFile, "file-2", 167, "2022-01-12 16:23"),
				newEntry(t, entities.EntryTypeFile, path, 167, "2022-01-12 16:23"),
			}

			connMock := connectionMocks.NewConnection(t)
			connMock.
				On("List", ctx, &connection.ListOptions{
					Path:    "",
					ShowAll: true,
				}).
				Return(entries, nil).
				Once()
			connMock.On("RemoveFile", path).Return(nil).Once()

			useCaseRepos := &ftp.RemoveRepos{
				Logger:     logger,
				Connection: connMock,
			}

			useCaseInput := &ftp.RemoveInput{
				Path:      path,
				Recursive: tc.recursive,
			}

			useCase := ftp.Remove{}

			// act
			err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

			// assert
			assert.NoError(t, err)
		})
	}
}

func Test_Remove_Execute_RemoveDir_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	entries := []*entities.Entry{
		newEntry(t, entities.EntryTypeFile, "file-1", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeFile, "file-2", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeDir, path, 167, "2022-01-12 16:23"),
	}

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return(entries, nil).
		Once()
	connMock.On("RemoveDir", path).Return(nil).Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: path,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	assert.NoError(t, err)
}

func Test_Remove_Execute_RemoveDirRecursively_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeDir, path, 167, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    path,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeFile, "file-1", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeLink, "link-1", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "dir-1", 167, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    filepath.Join(path, "dir-1"),
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeDir, ".", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "..", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeFile, "file-2", 167, "2022-01-12 16:23"),
		}, nil).
		Once()

	connMock.On("RemoveFile", filepath.Join(path, "file-1")).
		Return(nil).
		Once()
	connMock.On("RemoveFile", filepath.Join(path, "link-1")).
		Return(nil).
		Once()
	connMock.On("RemoveFile", filepath.Join(path, "dir-1", "file-2")).
		Return(nil).
		Once()

	connMock.On("RemoveDir", filepath.Join(path, "dir-1")).
		Return(nil).
		Once()
	connMock.On("RemoveDir", path).
		Return(nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path:      path,
		Recursive: true,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	assert.NoError(t, err)
}

func Test_Remove_Execute_ListError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to list directory").
		WithField("path", assertlogging.Equal("/")).
		WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return(nil, errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: path,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to list directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_NotFoundError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectInfo(fmt.Sprintf("entry not found under [%s] path", path)).
		WithField("path", assertlogging.Equal(path))

	entries := []*entities.Entry{
		newEntry(t, entities.EntryTypeFile, "file-1", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeFile, "file-2", 167, "2022-01-12 16:23"),
	}

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return(entries, nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: path,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, fmt.Sprintf("not found error occurred: entry not found under [%s] path", path))
	assert.IsType(t, ftperrors.NotFoundErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

// nolint:dupl // similar to Test_Remove_Execute_RemoveDir_Error
func Test_Remove_Execute_RemoveFile_Error(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to remove file").
		WithField("path", assertlogging.Equal(path)).
		WithError(assertlogging.EqualError("mock error"))

	entries := []*entities.Entry{
		newEntry(t, entities.EntryTypeFile, "file-1", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeFile, "file-2", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeFile, path, 167, "2022-01-12 16:23"),
	}

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return(entries, nil).
		Once()
	connMock.On("RemoveFile", path).Return(errors.New("mock error")).Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: path,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to remove file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

// nolint:dupl // similar to Test_Remove_Execute_RemoveFile_Error
func Test_Remove_Execute_RemoveDir_Error(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to remove directory").
		WithField("path", assertlogging.Equal(path)).
		WithError(assertlogging.EqualError("mock error"))

	entries := []*entities.Entry{
		newEntry(t, entities.EntryTypeFile, "file-1", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeFile, "file-2", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeDir, path, 167, "2022-01-12 16:23"),
	}

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return(entries, nil).
		Once()
	connMock.On("RemoveDir", path).Return(errors.New("mock error")).Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: path,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to remove directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_RemoveDirRecursively_ListError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to list directory").
		WithField("path", assertlogging.Equal(path)).
		WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeDir, path, 167, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    path,
			ShowAll: true,
		}).
		Return(nil, errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path:      path,
		Recursive: true,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to list directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_RemoveDirRecursively_RemoveFileError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to remove file").
		WithField("path", assertlogging.Equal(filepath.Join(path, "dir-1", "file-2"))).
		WithError(assertlogging.EqualError("mock error"))
	logger.
		ExpectError("failed to recursively remove directory").
		WithField("path", assertlogging.Equal(filepath.Join(path, "dir-1"))).
		WithError(
			assertlogging.EqualError("an internal error occurred: failed to remove file"),
			assertlogging.IsType(ftperrors.InternalErrorType),
		)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeDir, path, 167, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    path,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeFile, "file-1", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeLink, "link-1", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "dir-1", 167, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    filepath.Join(path, "dir-1"),
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeDir, ".", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "..", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeFile, "file-2", 167, "2022-01-12 16:23"),
		}, nil).
		Once()

	connMock.On("RemoveFile", filepath.Join(path, "file-1")).
		Return(nil).
		Once()
	connMock.On("RemoveFile", filepath.Join(path, "link-1")).
		Return(nil).
		Once()
	connMock.On("RemoveFile", filepath.Join(path, "dir-1", "file-2")).
		Return(errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path:      path,
		Recursive: true,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to recursively remove directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_RemoveDirRecursively_RemoveDirError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to remove directory").
		WithField("path", assertlogging.Equal(filepath.Join(path, "dir-1"))).
		WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeDir, path, 167, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    path,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeFile, "file-1", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeLink, "link-1", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "dir-1", 167, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    filepath.Join(path, "dir-1"),
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeDir, ".", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "..", 167, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeFile, "file-2", 167, "2022-01-12 16:23"),
		}, nil).
		Once()

	connMock.On("RemoveFile", filepath.Join(path, "file-1")).
		Return(nil).
		Once()
	connMock.On("RemoveFile", filepath.Join(path, "link-1")).
		Return(nil).
		Once()
	connMock.On("RemoveFile", filepath.Join(path, "dir-1", "file-2")).
		Return(nil).
		Once()

	connMock.On("RemoveDir", filepath.Join(path, "dir-1")).
		Return(errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path:      path,
		Recursive: true,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to remove directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_RemoveDirRecursively_UnknownError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    "",
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryTypeDir, path, 167, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    path,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			newEntry(t, entities.EntryType(0), "file-1", 167, "2022-01-12 16:23"),
		}, nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path:      path,
		Recursive: true,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an unknown error occurred: unexpected entry type: 0")
	assert.IsType(t, ftperrors.UnknownErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}
