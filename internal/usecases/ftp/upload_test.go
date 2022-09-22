package ftp_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
	"github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
	connectionMocks "github.com/alexZaicev/go-ftp-client/mocks/domain/connection"
)

func Test_UploadFile_Execute_NoDirSuccess(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On(
			"Upload",
			ctx,
			&connection.UploadOptions{
				Path:       fileName,
				FileReader: buffer,
			}).
		Return(nil).
		Once()
	connMock.
		On("Size", remotePathNoDir).
		Return(sizeInBytes, nil).
		Once()

	useCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.UploadFileInput{
		FileReader:  buffer,
		RemotePath:  remotePathNoDir,
		SizeInBytes: sizeInBytes,
	}

	useCase := &ftp.UploadFile{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	assert.NoError(t, err)
}

func Test_UploadFile_Execute_WithDirSuccess(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	logger := assertlogging.NewLogger(t)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("Cd", remoteDirPath).
		Return(nil).
		Once()
	connMock.
		On(
			"Upload",
			ctx,
			&connection.UploadOptions{
				Path:       fileName,
				FileReader: buffer,
			}).
		Return(nil).
		Once()
	connMock.
		On("Size", remotePathWithDir).
		Return(sizeInBytes, nil).
		Once()

	useCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.UploadFileInput{
		FileReader:  buffer,
		RemotePath:  remotePathWithDir,
		SizeInBytes: sizeInBytes,
	}

	useCase := &ftp.UploadFile{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	assert.NoError(t, err)
}

func Test_UploadFile_Execute_DirNotFoundError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError(fmt.Sprintf("directory %s not found", remoteDirPath)).
		WithError(assertlogging.EqualError(fmt.Sprintf("not found error: directory %s not found", remoteDirPath)))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("Cd", remoteDirPath).
		Return(ftperrors.NewNotFoundError(fmt.Sprintf("directory %s not found", remoteDirPath), nil)).
		Once()

	useCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.UploadFileInput{
		FileReader:  buffer,
		RemotePath:  remotePathWithDir,
		SizeInBytes: sizeInBytes,
	}

	useCase := &ftp.UploadFile{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	require.EqualError(t, err, fmt.Sprintf("not found error: directory %s not found", remoteDirPath))
	assert.IsType(t, ftperrors.NotFoundErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_UploadFile_Execute_CdError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to change directory").
		WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("Cd", remoteDirPath).
		Return(errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.UploadFileInput{
		FileReader:  buffer,
		RemotePath:  remotePathWithDir,
		SizeInBytes: sizeInBytes,
	}

	useCase := &ftp.UploadFile{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	require.EqualError(t, err, "an internal error occurred: failed to change directory")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_UploadFile_Execute_UploadError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to upload file").
		WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("Cd", remoteDirPath).
		Return(nil).
		Once()
	connMock.
		On(
			"Upload",
			ctx,
			&connection.UploadOptions{
				Path:       fileName,
				FileReader: buffer,
			}).
		Return(errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.UploadFileInput{
		FileReader:  buffer,
		RemotePath:  remotePathWithDir,
		SizeInBytes: sizeInBytes,
	}

	useCase := &ftp.UploadFile{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	require.EqualError(t, err, "an internal error occurred: failed to upload file")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_UploadFile_Execute_SizeError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to check file size").
		WithError(assertlogging.EqualError("mock error"))

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("Cd", remoteDirPath).
		Return(nil).
		Once()
	connMock.
		On(
			"Upload",
			ctx,
			&connection.UploadOptions{
				Path:       fileName,
				FileReader: buffer,
			}).
		Return(nil).
		Once()
	connMock.
		On("Size", remotePathWithDir).
		Return(uint64(0), errors.New("mock error")).
		Once()

	useCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.UploadFileInput{
		FileReader:  buffer,
		RemotePath:  remotePathWithDir,
		SizeInBytes: sizeInBytes,
	}

	useCase := &ftp.UploadFile{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	require.EqualError(t, err, "an internal error occurred: failed to check file size")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}

func Test_UploadFile_Execute_SizeMismatchError(t *testing.T) {
	ctx := context.Background()

	buffer := bytes.NewBufferString("this is content of awesome file")

	msg := fmt.Sprintf("uploaded file size %d does not match the actual %d", sizeInBytes-100, sizeInBytes)

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError(msg).
		WithFields(
			assertlogging.NewField("uploaded-size-in-bytes", assertlogging.Equal(sizeInBytes-100)),
			assertlogging.NewField("actual-size-in-bytes", assertlogging.Equal(sizeInBytes)),
		)

	connMock := connectionMocks.NewConnection(t)
	connMock.
		On("Cd", remoteDirPath).
		Return(nil).
		Once()
	connMock.
		On(
			"Upload",
			ctx,
			&connection.UploadOptions{
				Path:       fileName,
				FileReader: buffer,
			}).
		Return(nil).
		Once()
	connMock.
		On("Size", remotePathWithDir).
		Return(sizeInBytes-100, nil).
		Once()

	useCaseRepos := &ftp.UploadFileRepos{
		Logger:     logger,
		Connection: connMock,
	}
	useCaseInput := &ftp.UploadFileInput{
		FileReader:  buffer,
		RemotePath:  remotePathWithDir,
		SizeInBytes: sizeInBytes,
	}

	useCase := &ftp.UploadFile{}
	err := useCase.Execute(ctx, useCaseRepos, useCaseInput)
	require.EqualError(t, err, fmt.Sprintf("an internal error occurred: %s", msg))
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.Nil(t, errors.Unwrap(err))
}
