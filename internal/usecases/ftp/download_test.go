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
	repositoryMocks "github.com/alexZaicev/go-ftp-client/mocks/domain/repositories"
)

func Test_Download_Execute_File_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remotePathNoDir).
		Return(false, nil).
		Once()
	connMock.
		On("Size", remotePathNoDir).
		Return(sizeInBytes, nil).
		Once()
	connMock.
		On("Download", ctx, remotePathNoDir).
		Return(fileContent, nil).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)
	fileStoreMock.
		On("SaveFile", localPathWithDir, fileContent).
		Return(nil).
		Once()

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remotePathNoDir,
		Path:       localPathWithDir,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	assert.NoError(t, err)
}

func Test_Download_Execute_Directory_Success(t *testing.T) {
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
		On("Size", filepath.Join(remoteDirPath, "file-1")).
		Return(sizeInBytes, nil).
		Once()
	connMock.
		On("Size", filepath.Join(remoteDirPath, "dir-1", "file-2")).
		Return(sizeInBytes, nil).
		Once()
	connMock.
		On("Download", ctx, filepath.Join(remoteDirPath, "file-1")).
		Return(fileContent, nil).
		Once()
	connMock.
		On("Download", ctx, filepath.Join(remoteDirPath, "dir-1", "file-2")).
		Return(fileContent, nil).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)
	fileStoreMock.
		On("CreateDir", dirPath).
		Return(nil).
		Once()
	fileStoreMock.
		On("CreateDir", filepath.Join(dirPath, "dir-1")).
		Return(nil).
		Once()
	fileStoreMock.
		On("SaveFile", filepath.Join(dirPath, "file-1"), fileContent).
		Return(nil).
		Once()
	fileStoreMock.
		On("SaveFile", filepath.Join(dirPath, "dir-1", "file-2"), fileContent).
		Return(nil).
		Once()

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remoteDirPath,
		Path:       dirPath,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	assert.NoError(t, err)
}

func Test_Download_Execute_IsDirError(t *testing.T) {
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

	fileStoreMock := repositoryMocks.NewFileStore(t)

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remotePathNoDir,
		Path:       localPathWithDir,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to check if entry is a directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Download_Execute_SizeError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to retrieve file size").
		WithError(assertlogging.EqualError("mock error")).
		WithField("remote-path", assertlogging.Equal(remotePathNoDir))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remotePathNoDir).
		Return(false, nil).
		Once()
	connMock.
		On("Size", remotePathNoDir).
		Return(uint64(0), errors.New("mock error")).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remotePathNoDir,
		Path:       localPathWithDir,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to retrieve file size")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Download_Execute_DownloadError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to download file").
		WithError(assertlogging.EqualError("mock error")).
		WithField("remote-path", assertlogging.Equal(remotePathNoDir))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remotePathNoDir).
		Return(false, nil).
		Once()
	connMock.
		On("Size", remotePathNoDir).
		Return(sizeInBytes, nil).
		Once()
	connMock.
		On("Download", ctx, remotePathNoDir).
		Return(fileContent, errors.New("mock error")).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remotePathNoDir,
		Path:       localPathWithDir,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to download file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Download_Execute_SizeMismatchError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError(fmt.Sprintf("downloaded file size %d does not match the actual 1024", sizeInBytes)).
		WithFields(
			assertlogging.NewField("remote-path", assertlogging.Equal(remotePathNoDir)),
			assertlogging.NewField("actual-size-in-bytes", assertlogging.Equal(uint64(1024))),
			assertlogging.NewField("downloaded-size-in-bytes", assertlogging.Equal(sizeInBytes)),
		)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remotePathNoDir).
		Return(false, nil).
		Once()
	connMock.
		On("Size", remotePathNoDir).
		Return(uint64(1024), nil).
		Once()
	connMock.
		On("Download", ctx, remotePathNoDir).
		Return(fileContent, nil).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remotePathNoDir,
		Path:       localPathWithDir,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(
		t,
		err,
		fmt.Sprintf("an internal error occurred: downloaded file size %d does not match the actual 1024", sizeInBytes),
	)
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Download_Execute_SaveError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to save file").
		WithError(assertlogging.EqualError("mock error")).
		WithFields(
			assertlogging.NewField("remote-path", assertlogging.Equal(remotePathNoDir)),
			assertlogging.NewField("path", assertlogging.Equal(localPathWithDir)),
		)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remotePathNoDir).
		Return(false, nil).
		Once()
	connMock.
		On("Size", remotePathNoDir).
		Return(sizeInBytes, nil).
		Once()
	connMock.
		On("Download", ctx, remotePathNoDir).
		Return(fileContent, nil).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)
	fileStoreMock.
		On("SaveFile", localPathWithDir, fileContent).
		Return(errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remotePathNoDir,
		Path:       localPathWithDir,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to save file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Download_Execute_CreateDirError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to create directory").
		WithError(assertlogging.EqualError("mock error")).
		WithField("path", assertlogging.Equal(dirPath))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("IsDir", ctx, remoteDirPath).
		Return(true, nil).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)
	fileStoreMock.
		On("CreateDir", dirPath).
		Return(errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remoteDirPath,
		Path:       dirPath,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to create directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Download_Execute_ListError(t *testing.T) {
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
		On("Size", filepath.Join(remoteDirPath, "file-1")).
		Return(sizeInBytes, nil).
		Once()
	connMock.
		On("Download", ctx, filepath.Join(remoteDirPath, "file-1")).
		Return(fileContent, nil).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)
	fileStoreMock.
		On("CreateDir", dirPath).
		Return(nil).
		Once()
	fileStoreMock.
		On("CreateDir", filepath.Join(dirPath, "dir-1")).
		Return(nil).
		Once()
	fileStoreMock.
		On("SaveFile", filepath.Join(dirPath, "file-1"), fileContent).
		Return(nil).
		Once()

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remoteDirPath,
		Path:       dirPath,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to list directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Download_Execute_Directory_SizeError(t *testing.T) {
	// arrange
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to retrieve file size").
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
		On("Size", filepath.Join(remoteDirPath, "file-1")).
		Return(sizeInBytes, nil).
		Once()
	connMock.
		On("Size", filepath.Join(remoteDirPath, "dir-1", "file-2")).
		Return(uint64(0), errors.New("mock error")).
		Once()
	connMock.
		On("Download", ctx, filepath.Join(remoteDirPath, "file-1")).
		Return(fileContent, nil).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)
	fileStoreMock.
		On("CreateDir", dirPath).
		Return(nil).
		Once()
	fileStoreMock.
		On("CreateDir", filepath.Join(dirPath, "dir-1")).
		Return(nil).
		Once()
	fileStoreMock.
		On("SaveFile", filepath.Join(dirPath, "file-1"), fileContent).
		Return(nil).
		Once()

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remoteDirPath,
		Path:       dirPath,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an internal error occurred: failed to retrieve file size")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_Download_Execute_UnknownError(t *testing.T) {
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
		On("Size", filepath.Join(remoteDirPath, "file-1")).
		Return(sizeInBytes, nil).
		Once()
	connMock.
		On("Download", ctx, filepath.Join(remoteDirPath, "file-1")).
		Return(fileContent, nil).
		Once()

	fileStoreMock := repositoryMocks.NewFileStore(t)
	fileStoreMock.
		On("CreateDir", dirPath).
		Return(nil).
		Once()
	fileStoreMock.
		On("CreateDir", filepath.Join(dirPath, "dir-1")).
		Return(nil).
		Once()
	fileStoreMock.
		On("SaveFile", filepath.Join(dirPath, "file-1"), fileContent).
		Return(nil).
		Once()

	useCaseRepos := &ftp.DownloadRepos{
		Logger:     logger,
		Connection: connMock,
		FileStore:  fileStoreMock,
	}

	useCaseInput := &ftp.DownloadInput{
		RemotePath: remoteDirPath,
		Path:       dirPath,
	}

	useCase := &ftp.Download{}

	// act
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)

	// assert
	require.EqualError(t, err, "an unknown error occurred: unexpected entry type: 0")
	assert.IsType(t, ftperrors.UnknownErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}
