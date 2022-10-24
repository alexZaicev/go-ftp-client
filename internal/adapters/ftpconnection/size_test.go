package ftpconnection_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
	ftpConnectionMocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

func Test_ServerConnection_Size_Success(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandSize, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusFile).
		Return(models.StatusFile, "1024", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	sizeInBytes, err := serverConn.Size(remotePath)
	assert.Equal(t, uint64(1024), sizeInBytes)
	assert.NoError(t, err)
}

func Test_ServerConnection_Size_CmdError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandSize, remotePath).
		Return(uid, errors.New("mock error")).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	sizeInBytes, err := serverConn.Size(remotePath)
	assert.Equal(t, uint64(0), sizeInBytes)
	require.EqualError(t, err, "an internal error occurred: failed to fetch file size")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), "mock error")
}

func Test_ServerConnection_Size_ParseError(t *testing.T) {
	tcpConn := ftpConnectionMocks.NewConn(t)
	dialer := ftpConnectionMocks.NewDialer(t)
	connMock := ftpConnectionMocks.NewTextConnection(t)
	connMock.
		On("Cmd", models.CommandSize, remotePath).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusFile).
		Return(models.StatusFile, "no-a-number", nil).
		Once()

	serverConn, err := ftpconnection.NewConnection(host, dialer, tcpConn, connMock)
	require.NoError(t, err)

	sizeInBytes, err := serverConn.Size(remotePath)
	assert.Equal(t, uint64(0), sizeInBytes)
	require.EqualError(t, err, "an internal error occurred: failed to parse file size to a non-zero integer")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.EqualError(t, errors.Unwrap(err), `strconv.ParseUint: parsing "no-a-number": invalid syntax`)
}
