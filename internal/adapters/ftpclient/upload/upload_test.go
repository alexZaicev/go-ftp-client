package upload_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/upload"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
	ftpclientMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpclient"
	connectionMocks "github.com/alexZaicev/go-ftp-client/mocks/domain/connection"
	useCaseMocks "github.com/alexZaicev/go-ftp-client/mocks/usecases/ftp"
)

const (
	address  = "10.0.0.1:21"
	user     = "user01"
	password = "pwd01"
	timeout  = 5 * time.Second

	filePath = "testdata/file-1.txt"
	dirPath  = "testdata"

	remotePath     = "/foo/bar/baz"
	remoteFilePath = "/foo/bar/baz/file-1.txt"
)

func Test_PerformUploadFile_Success(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectInfo("OK!")

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(ftpConnMock, nil).
		Once()

	mkdirUseCaseRepos := &ftp.MkdirRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	mkdirUseCaseInput := &ftp.MkdirInput{
		Path: remotePath,
	}

	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	mkdirUseCaseMock.
		On("Execute", ctx, mkdirUseCaseRepos, mkdirUseCaseInput).
		Return(nil).
		Once()

	uploadUseCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}

	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)
	uploadUseCaseMock.
		On("Execute", ctx, uploadUseCaseRepos, mock.AnythingOfType("*ftp.UploadFileInput")).
		Run(func(args mock.Arguments) {
			bytesToRead := make([]byte, 1024)
			_, useCaseMockErr := args.Get(2).(*ftp.UploadFileInput).FileReader.Read(bytesToRead)
			require.NoError(t, useCaseMockErr)
		}).
		Return(nil).
		Once()

	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			filePath: {Data: []byte("this is content of the file")},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}
	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       filePath,
		RemoteFilePath: remoteFilePath,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	assert.NoError(t, err)
}

func Test_PerformUploadFile_Recursive_Success(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectInfo("OK!")

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(ftpConnMock, nil).
		Once()

	mkdirUseCaseRepos := &ftp.MkdirRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}

	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	mkdirUseCaseMock.
		On("Execute", ctx, mkdirUseCaseRepos, &ftp.MkdirInput{
			Path: remotePath,
		}).
		Return(nil).
		Twice()
	mkdirUseCaseMock.
		On("Execute", ctx, mkdirUseCaseRepos, &ftp.MkdirInput{
			Path: fmt.Sprintf("%s/dir1", remotePath),
		}).
		Return(nil).
		Once()
	mkdirUseCaseMock.
		On("Execute", ctx, mkdirUseCaseRepos, &ftp.MkdirInput{
			Path: fmt.Sprintf("%s/dir2", remotePath),
		}).
		Return(nil).
		Times(3)

	uploadUseCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}

	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)
	uploadUseCaseMock.
		On("Execute", ctx, uploadUseCaseRepos, mock.AnythingOfType("*ftp.UploadFileInput")).
		Run(func(args mock.Arguments) {
			bytesToRead := make([]byte, 1024)
			_, useCaseMockErr := args.Get(2).(*ftp.UploadFileInput).FileReader.Read(bytesToRead)
			require.NoError(t, useCaseMockErr)
		}).
		Return(nil).
		Times(6)

	absFilePath, err := filepath.Abs(fmt.Sprintf("./%s", dirPath))
	require.NoError(t, err)

	fsabsFilePath := absFilePath[1:]
	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			fsabsFilePath: {Mode: fs.ModeDir},
			fmt.Sprintf("%s/file-1.txt", fsabsFilePath):      {Data: []byte("this is content of the file")},
			fmt.Sprintf("%s/file-2.txt", fsabsFilePath):      {Data: []byte("this is content of the file")},
			fmt.Sprintf("%s/dir1", fsabsFilePath):            {Mode: fs.ModeDir},
			fmt.Sprintf("%s/dir1/file-1.txt", fsabsFilePath): {Data: []byte("this is content of the file")},
			fmt.Sprintf("%s/dir2", fsabsFilePath):            {Mode: fs.ModeDir},
			fmt.Sprintf("%s/dir2/file-1.txt", fsabsFilePath): {Data: []byte("this is content of the file")},
			fmt.Sprintf("%s/dir2/file-2.txt", fsabsFilePath): {Data: []byte("this is content of the file")},
			fmt.Sprintf("%s/dir2/file-3.txt", fsabsFilePath): {Data: []byte("this is content of the file")},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}

	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       absFilePath,
		RemoteFilePath: remotePath,
		Recursive:      true,
	}

	err = upload.PerformUploadFile(ctx, logger, deps, input)
	assert.NoError(t, err)
}

func Test_PerformUploadFile_FileOpenError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := ftpclientMocks.NewConnector(t)
	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)

	deps := &upload.Dependencies{
		Filesystem:    fstest.MapFS{},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}
	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       filePath,
		RemoteFilePath: remoteFilePath,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	require.EqualError(t, err, "an internal error occurred: failed to open file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "open testdata/file-1.txt: file does not exist")
}

func Test_PerformUploadFile_NonRecursiveDirUploadError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := ftpclientMocks.NewConnector(t)
	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)

	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			filePath: {Mode: fs.ModeDir},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}
	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       filePath,
		RemoteFilePath: remoteFilePath,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	require.EqualError(t, err, "an internal error occurred: path is not a regular file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformUploadFile_RecursiveNonDirUploadError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := ftpclientMocks.NewConnector(t)
	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)

	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			filePath: {Data: []byte("this is content of the file")},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}
	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       filePath,
		RemoteFilePath: remoteFilePath,
		Recursive:      true,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	require.EqualError(t, err, "an internal error occurred: path is not a directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformUploadFile_WalkError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := ftpclientMocks.NewConnector(t)
	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)

	absFilePath := "/non-existing-dir"

	fsabsFilePath := absFilePath[1:]
	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			fsabsFilePath: {Mode: fs.ModeDir},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}

	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       absFilePath,
		RemoteFilePath: remotePath,
		Recursive:      true,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	require.EqualError(t, err, "an internal error occurred: failed to walk directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "lstat /non-existing-dir: no such file or directory")
}

func Test_PerformUploadFile_RecursiveFileOpenErrorError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	connMock := ftpclientMocks.NewConnector(t)
	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)

	absFilePath, err := filepath.Abs(fmt.Sprintf("./%s", dirPath))
	require.NoError(t, err)

	fsabsFilePath := absFilePath[1:]
	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			fsabsFilePath: {Mode: fs.ModeDir},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}

	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       absFilePath,
		RemoteFilePath: remotePath,
		Recursive:      true,
	}

	err = upload.PerformUploadFile(ctx, logger, deps, input)
	require.EqualError(t, err, "an internal error occurred: failed to walk directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "an internal error occurred: failed to open file")
}

func Test_PerformUploadFile_ConnectionError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to connect to server").
		WithError(assertlogging.EqualError("mock error"))

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(nil, errors.New("mock error")).
		Once()

	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)

	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			filePath: {Data: []byte("this is content of the file")},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}
	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       filePath,
		RemoteFilePath: remoteFilePath,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	assert.EqualError(t, err, "mock error")
}

func Test_PerformUploadFile_ConnectionStopError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectInfo("OK!")
	logger.
		ExpectError("failed to stop server connection").
		WithError(assertlogging.EqualError("mock error"))

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(errors.New("mock error")).Once()

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(ftpConnMock, nil).
		Once()

	mkdirUseCaseRepos := &ftp.MkdirRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	mkdirUseCaseInput := &ftp.MkdirInput{
		Path: remotePath,
	}

	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	mkdirUseCaseMock.
		On("Execute", ctx, mkdirUseCaseRepos, mkdirUseCaseInput).
		Return(nil).
		Once()

	uploadUseCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}

	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)
	uploadUseCaseMock.
		On("Execute", ctx, uploadUseCaseRepos, mock.AnythingOfType("*ftp.UploadFileInput")).
		Run(func(args mock.Arguments) {
			bytesToRead := make([]byte, 1024)
			_, useCaseMockErr := args.Get(2).(*ftp.UploadFileInput).FileReader.Read(bytesToRead)
			require.NoError(t, useCaseMockErr)
		}).
		Return(nil).
		Once()

	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			filePath: {Data: []byte("this is content of the file")},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}
	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       filePath,
		RemoteFilePath: remoteFilePath,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	assert.EqualError(t, err, "mock error")
}

func Test_PerformUploadFile_MkdirError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(ftpConnMock, nil).
		Once()

	mkdirUseCaseRepos := &ftp.MkdirRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	mkdirUseCaseInput := &ftp.MkdirInput{
		Path: remotePath,
	}

	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	mkdirUseCaseMock.
		On("Execute", ctx, mkdirUseCaseRepos, mkdirUseCaseInput).
		Return(errors.New("mock error")).
		Once()

	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)

	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			filePath: {Data: []byte("this is content of the file")},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}
	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},
		FilePath:       filePath,
		RemoteFilePath: remoteFilePath,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	assert.EqualError(t, err, "mock error")
}

func Test_PerformUploadFile_UploadError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	config := ftpclient.ConnectorConfig{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, config).
		Return(ftpConnMock, nil).
		Once()

	mkdirUseCaseRepos := &ftp.MkdirRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	mkdirUseCaseInput := &ftp.MkdirInput{
		Path: remotePath,
	}

	mkdirUseCaseMock := useCaseMocks.NewMkdirUseCase(t)
	mkdirUseCaseMock.
		On("Execute", ctx, mkdirUseCaseRepos, mkdirUseCaseInput).
		Return(nil).
		Once()

	uploadUseCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}

	uploadUseCaseMock := useCaseMocks.NewUploadFileUseCase(t)
	uploadUseCaseMock.
		On("Execute", ctx, uploadUseCaseRepos, mock.AnythingOfType("*ftp.UploadFileInput")).
		Return(errors.New("mock error")).
		Once()

	deps := &upload.Dependencies{
		Filesystem: fstest.MapFS{
			filePath: {Data: []byte("this is content of the file")},
		},
		Connector:     connMock,
		UploadUseCase: uploadUseCaseMock,
		MkdirUseCase:  mkdirUseCaseMock,
	}
	input := &upload.CmdUploadInput{
		Config: ftpclient.ConnectorConfig{
			Address:  address,
			User:     user,
			Password: password,
			Verbose:  true,
			Timeout:  timeout,
		},

		FilePath:       filePath,
		RemoteFilePath: remoteFilePath,
	}

	err := upload.PerformUploadFile(ctx, logger, deps, input)
	assert.EqualError(t, err, "mock error")
}
