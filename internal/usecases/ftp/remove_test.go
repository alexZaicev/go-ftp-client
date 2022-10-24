package ftp_test

import (
	"context"
	"errors"
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

func Test_Remove_Execute_File_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remotePathNoDir).
		Return(false, nil).
		Once()
	connMock.
		On("RemoveFile", remotePathNoDir).
		Return(nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: remotePathNoDir,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	assert.NoError(t, err)
}

func Test_Remove_Execute_Directory_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remoteDirPath).
		Return(true, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    remoteDirPath,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryTypeFile, "file-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeLink, "link-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "dir-1", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    filepath.Join(remoteDirPath, "dir-1"),
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryTypeFile, "file-2", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("RemoveDir", remoteDirPath).
		Return(nil).
		Once()
	connMock.
		On("RemoveDir", filepath.Join(remoteDirPath, "dir-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "file-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "link-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "dir-1", "file-2")).
		Return(nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: remoteDirPath,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	assert.NoError(t, err)
}

func Test_Remove_Execute_IsDirError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to check if entry is a directory").
		WithError(assertlogging.EqualError("mock error")).
		WithField("remote-path", assertlogging.Equal(remotePathNoDir))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remotePathNoDir).
		Return(false, errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: remotePathNoDir,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to check if entry is a directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_RemoveFileError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to remove file").
		WithError(assertlogging.EqualError("mock error")).
		WithField("remote-path", assertlogging.Equal(remotePathNoDir))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remotePathNoDir).
		Return(false, nil).
		Once()
	connMock.
		On("RemoveFile", remotePathNoDir).
		Return(errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: remotePathNoDir,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to remove file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_ListError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to list directory").
		WithError(assertlogging.EqualError("mock error")).
		WithField("remote-path", assertlogging.Equal(filepath.Join(remoteDirPath, "dir-1")))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remoteDirPath).
		Return(true, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    remoteDirPath,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryTypeFile, "file-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeLink, "link-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "dir-1", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    filepath.Join(remoteDirPath, "dir-1"),
			ShowAll: true,
		}).
		Return(nil, errors.New("mock error")).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "file-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "link-1")).
		Return(nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: remoteDirPath,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to list directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_Directory_RemoveFileError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to remove file").
		WithError(assertlogging.EqualError("mock error")).
		WithField("remote-path", assertlogging.Equal(filepath.Join(remoteDirPath, "dir-1", "file-2")))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remoteDirPath).
		Return(true, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    remoteDirPath,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryTypeFile, "file-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeLink, "link-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "dir-1", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    filepath.Join(remoteDirPath, "dir-1"),
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryTypeFile, "file-2", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "file-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "link-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "dir-1", "file-2")).
		Return(errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: remoteDirPath,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to remove file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_RemoveDirError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to remove directory").
		WithError(assertlogging.EqualError("mock error")).
		WithField("remote-path", assertlogging.Equal(remoteDirPath))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remoteDirPath).
		Return(true, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    remoteDirPath,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryTypeFile, "file-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeLink, "link-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "dir-1", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    filepath.Join(remoteDirPath, "dir-1"),
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryTypeFile, "file-2", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("RemoveDir", remoteDirPath).
		Return(errors.New("mock error")).
		Once()
	connMock.
		On("RemoveDir", filepath.Join(remoteDirPath, "dir-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "file-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "link-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "dir-1", "file-2")).
		Return(nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: remoteDirPath,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to remove directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Remove_Execute_UnknownError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remoteDirPath).
		Return(true, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    remoteDirPath,
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryTypeFile, "file-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeLink, "link-1", sizeInBytes, "2022-01-12 16:23"),
			newEntry(t, entities.EntryTypeDir, "dir-1", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("List", ctx, &connection.ListOptions{
			Path:    filepath.Join(remoteDirPath, "dir-1"),
			ShowAll: true,
		}).
		Return([]*entities.Entry{
			rootDir1,
			rootDir2,
			newEntry(t, entities.EntryType(0), "file-2", sizeInBytes, "2022-01-12 16:23"),
		}, nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "file-1")).
		Return(nil).
		Once()
	connMock.
		On("RemoveFile", filepath.Join(remoteDirPath, "link-1")).
		Return(nil).
		Once()

	useCaseRepos := &ftp.RemoveRepos{
		Logger:     logger,
		Connection: connMock,
	}

	useCaseInput := &ftp.RemoveInput{
		Path: remoteDirPath,
	}

	useCase := ftp.Remove{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an unknown error occurred: unexpected entry type: 0")
	assert.IsType(t, ftperrors.UnknownErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}
