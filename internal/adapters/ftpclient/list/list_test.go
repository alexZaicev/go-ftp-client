package list_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient/list"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	"github.com/alexZaicev/go-ftp-client/internal/drivers/cli/models"
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

	path = "/foo/bar/baz"

	dateFormat = "2006-01-02 15:04"
)

func Test_PerformListFiles_Success(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	expectedEntries := getEntries(t)

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.ListFilesRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.ListFilesInput{
		Path:     path,
		ShowAll:  true,
		SortType: entities.SortTypeName,
	}

	useCaseMock := useCaseMocks.NewListFilesUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(expectedEntries, nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &list.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &list.CmdListInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		Path:     path,
		ShowAll:  true,
		SortType: models.SortTypeName,
	}

	expectedStatusStr := `+------+-------------+----------------+-------------------+-------------------+----------+
| TYPE | PERMISSIONS |     OWNERS     |       NAME        |   LAST MODIFIED   |   SIZE   |
+------+-------------+----------------+-------------------+-------------------+----------+
| F    | rwxrwxrwx   | user01:group01 | file5             | Wed, 12 Jan 16:23 | 167 B    |
| D    | rwxrwxrwx   | user01:group01 | dir1              | Mon, 24 Jan 16:23 | 39.09 KB |
| F    | rwxrwxrwx   | user01:group01 | file7             | Thu, 02 Apr 14:23 | 4.25 KB  |
| D    | rwxrwxrwx   | user01:group01 | dir2              | Sun, 02 Jan 15:23 | 100 B    |
| L    | rwxrwxrwx   | user01:group01 | dir-link -> /dir1 | Fri, 02 Sep 10:23 | 1.01 KB  |
| F    | rwxrwxrwx   | user01:group01 | file4             | Wed, 12 Jan 19:23 | 4.92 KB  |
+------+-------------+----------------+-------------------+-------------------+----------+
`

	err := list.PerformListFiles(ctx, logger, deps, input)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatusStr, buffer.String())
}

func Test_PerformListFiles_NotFoundError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.ExpectInfo("no entries found under specified path")

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.ListFilesRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.ListFilesInput{
		Path:     path,
		ShowAll:  true,
		SortType: entities.SortTypeName,
	}

	useCaseMock := useCaseMocks.NewListFilesUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil, ftperrors.NewNotFoundError("mock error", nil)).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &list.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &list.CmdListInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		Path:     path,
		ShowAll:  true,
		SortType: models.SortTypeName,
	}

	err := list.PerformListFiles(ctx, logger, deps, input)
	assert.NoError(t, err)
	assert.Equal(t, "", buffer.String())
}

func Test_PerformListFiles_ConnectionError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to connect to server").
		WithError(assertlogging.EqualError("mock error"))

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(nil, errors.New("mock error")).
		Once()

	useCaseMock := useCaseMocks.NewListFilesUseCase(t)

	buffer := bytes.NewBufferString("")

	deps := &list.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &list.CmdListInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		Path:     path,
		ShowAll:  true,
		SortType: models.SortTypeName,
	}

	err := list.PerformListFiles(ctx, logger, deps, input)
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformListFiles_ConnectionStopError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)
	logger.
		ExpectError("failed to stop server connection").
		WithError(assertlogging.EqualError("mock error"))

	expectedEntries := getEntries(t)

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(errors.New("mock error")).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.ListFilesRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.ListFilesInput{
		Path:     path,
		ShowAll:  true,
		SortType: entities.SortTypeName,
	}

	useCaseMock := useCaseMocks.NewListFilesUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(expectedEntries, nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &list.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &list.CmdListInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		Path:     path,
		ShowAll:  true,
		SortType: models.SortTypeName,
	}

	err := list.PerformListFiles(ctx, logger, deps, input)
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformListFiles_SortTypeConvertError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseMock := useCaseMocks.NewListFilesUseCase(t)

	buffer := bytes.NewBufferString("")

	deps := &list.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &list.CmdListInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		Path:     path,
		ShowAll:  true,
		SortType: models.SortType("not-a-valid-sort-type"),
	}

	err := list.PerformListFiles(ctx, logger, deps, input)
	require.EqualError(t, err, "an unknown error occurred: unexpected sort type: not-a-valid-sort-type")
	assert.IsType(t, ftperrors.UnknownErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformListFiles_UseCaseError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.ListFilesRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.ListFilesInput{
		Path:     path,
		ShowAll:  true,
		SortType: entities.SortTypeName,
	}

	useCaseMock := useCaseMocks.NewListFilesUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(nil, errors.New("mock error")).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &list.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &list.CmdListInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		Path:     path,
		ShowAll:  true,
		SortType: models.SortTypeName,
	}

	err := list.PerformListFiles(ctx, logger, deps, input)
	require.EqualError(t, err, "mock error")
	assert.NoError(t, errors.Unwrap(err))
}

func Test_PerformListFiles_EntryTypeConvertError(t *testing.T) {
	ctx := context.Background()

	logger := assertlogging.NewLogger(t)

	entries := []*entities.Entry{
		newEntry(
			t,
			entities.EntryType(0),
			"file5",
			167,
			"2022-01-12 16:23",
		),
	}

	ftpConnMock := connectionMocks.NewConnection(t)
	ftpConnMock.On("Stop").Return(nil).Once()

	options := &ftpclient.ConnectorOptions{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
	}
	connMock := ftpclientMocks.NewConnector(t)
	connMock.
		On("Connect", ctx, options).
		Return(ftpConnMock, nil).
		Once()

	useCaseRepos := &ftp.ListFilesRepos{
		Logger:     logger,
		Connection: ftpConnMock,
	}
	useCaseInput := &ftp.ListFilesInput{
		Path:     path,
		ShowAll:  true,
		SortType: entities.SortTypeName,
	}

	useCaseMock := useCaseMocks.NewListFilesUseCase(t)
	useCaseMock.
		On("Execute", ctx, useCaseRepos, useCaseInput).
		Return(entries, nil).
		Once()

	buffer := bytes.NewBufferString("")

	deps := &list.Dependencies{
		Connector: connMock,
		UseCase:   useCaseMock,
		OutWriter: buffer,
	}
	input := &list.CmdListInput{
		Address:  address,
		User:     user,
		Password: password,
		Verbose:  true,
		Timeout:  timeout,
		Path:     path,
		ShowAll:  true,
		SortType: models.SortTypeName,
	}

	err := list.PerformListFiles(ctx, logger, deps, input)
	require.EqualError(t, err, "an unknown error occurred: unexpected entry type: 0")
	assert.IsType(t, ftperrors.UnknownErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}

func getEntries(t *testing.T) []*entities.Entry {
	return []*entities.Entry{
		newEntry(t, entities.EntryTypeFile, "file5", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeDir, "dir1", 40032, "2022-01-24 16:23"),
		newEntry(t, entities.EntryTypeFile, "file7", 4352, "2020-04-02 14:23"),
		newEntry(t, entities.EntryTypeDir, "dir2", 100, "2022-01-02 15:23"),
		newEntry(t, entities.EntryTypeLink, "dir-link -> /dir1", 1034, "2022-09-02 10:23"),
		newEntry(t, entities.EntryTypeFile, "file4", 5043, "2022-01-12 19:23"),
	}
}

func newEntry(t *testing.T, entryType entities.EntryType, name string, sizeInBytes uint64, dateStr string) *entities.Entry {
	date, err := time.Parse(dateFormat, dateStr)
	require.NoError(t, err, "Failed to parse test case date")

	return &entities.Entry{
		Type:                 entryType,
		Permissions:          "rwxrwxrwx",
		Name:                 name,
		OwnerUser:            "user01",
		OwnerGroup:           "group01",
		SizeInBytes:          sizeInBytes,
		NumHardLinks:         2,
		LastModificationDate: date,
	}
}
