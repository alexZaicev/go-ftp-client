package ftpconnection_test

import (
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	mocks "github.com/alexZaicev/go-ftp-client/mocks/adapters/ftpconnection"
)

const (
	uid uint = 1

	remotePath       = "/foo/bar/baz"
	remoteParentPath = "/foo/bar/"
	newRemotePath    = "/baz/bar/foo"

	host     = "ftp-dev-client"
	user     = "user01"
	password = "pwd01"
)

// Feature messages returned by FEAT command
const (
	featureMsg = `211-Features:
 EPRT
 EPSV
 MDTM
 PASV
 REST STREAM
 SIZE
 TVFS
 UTF8
 MLST
211 End`

	featureMsgWithoutMLST = `211-Features:
 EPRT
 EPSV
 MDTM
 PASV
 REST STREAM
 SIZE
 TVFS
 UTF8
 PRET
211 End`

	featureMsgWithoutUTF8 = `211-Features:
 EPRT
 EPSV
 MDTM
 PASV
 REST STREAM
 SIZE
 TVFS
 MLST
211 End`
)

// Status messages returned by STAT and SYST commands.
const (
	statusMsg = `211-FTP server status:
     Connected to 172.22.0.2
     Logged in as ftpuser01
     TYPE: BINARY
     No session bandwidth limit
     Session timeout in seconds is 300
     Control connection is plain text
     Data connections will be plain text
     At session startup, client count was 1
     vsFTPd 3.0.2 - secure, fast, stable
211 End of status`

	systemMsg = `UNIX Type: L8`
)

const (
	extendedPassiveModeMessage = "Entering Extended Passive Mode (|||21103|)."
	passiveModeMessage         = "Entering Passive Mode (10,0,0,1,82,111)"
	listMessage                = "Here comes the directory listing."
	entryFileMessage           = "-rw-r--r--    1 ftp      ftp           187 Sep 16 14:34 file-1.txt"
	entryDirMessage            = "drwxrw-rw-    1 ftp      ftp           187 Sep 16 14:34 baz"
)

func setMocksForLogin(connMock *mocks.TextConnection, useTLS bool) {
	connMock.
		On("Cmd", models.CommandUser, user).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusLoggedIn, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandFeat).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusSystem, featureMsgWithoutMLST, nil).
		Once()
	connMock.
		On("Cmd", models.CommandType, models.TransferTypeBinary).
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusCommandOK).
		Return(models.StatusCommandOK, "", nil).
		Once()
	connMock.
		On("Cmd", models.CommandOptions, models.FeatureUTF8, "ON").
		Return(uid, nil).
		Once()
	connMock.
		On("ReadResponse", models.StatusNoCheck).
		Return(models.StatusCommandOK, "", nil).
		Once()

	if useTLS {
		connMock.
			On("Cmd", models.CommandProtectionBufferSize).
			Return(uid, nil).
			Once()
		connMock.
			On("Cmd", models.CommandProtocol).
			Return(uid, nil).
			Once()
		connMock.
			On("ReadResponse", models.StatusCommandOK).
			Return(models.StatusCommandOK, "", nil).
			Twice()
	}
}
