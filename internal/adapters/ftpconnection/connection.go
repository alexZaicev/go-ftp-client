package ftpconnection

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/parsers"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	CommandQuit                 = "QUIT"
	CommandAuthTLS              = "AUTH TLS"
	CommandUser                 = "USER %s"
	CommandPass                 = "PASS %s"
	CommandFeat                 = "FEAT"
	CommandProtectionBufferSize = "PBSZ 0"
	CommandProtocol             = "PROT P"
	CommandType                 = "TYPE %s"
	CommandOptions              = "OPTS %s %s"
	CommandStatus               = "STAT"
	CommandSystem               = "SYST"
	CommandList                 = "LIST %s"
	CommandListHidden           = "LIST -a %s"
	CommandPreTransfer          = "PRET %s"
	CommandPassive              = "PASV"
	CommandExtendedPassiveMode  = "EPSV"
	CommandRestartTransfer      = "REST %d"
	CommandListMachineReadable  = "MLSD %s"
	CommandStore                = "STOR %s"
	CommandMakeDir              = "MKD %s"
	CommandChangeWorkDir        = "CWD %s"
	CommandSize                 = "SIZE %s"

	Off = "OFF"
	On  = "ON"
)

type TransferType string

const (
	TransferTypeASCII  = "A"
	TransferTypeBinary = "I"
)

const (
	decimalBase = 10
	bitSize     = 64

	defaultShutTimeout = 500 * time.Millisecond
)

type TextConnection interface {
	ReadResponse(expectedCode int) (int, string, error)
	Cmd(format string, args ...any) (uint, error)
	Close() error
}

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	DialContextTLS(ctx context.Context, network, address string, tlsConfig *tls.Config) (net.Conn, error)
}

type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type ServerConnection struct {
	host string

	dialer Dialer

	conn    TextConnection
	tcpConn net.Conn

	parser parsers.Parser

	features *serverFeatures

	disableUTF8   bool
	disableEPSV   bool
	verboseWriter io.Writer
	tlsConfig     *tls.Config
	shutTimeout   time.Duration
}

func NewConnection(
	host string,
	dialer Dialer,
	conn net.Conn,
	textConn TextConnection,
	options ...Option,
) (connection.Connection, error) {
	if host == "" {
		return nil, ftperrors.NewInvalidArgumentError("host", ftperrors.ErrMsgCannotBeBlank)
	}
	if dialer == nil {
		return nil, ftperrors.NewInvalidArgumentError("dialer", ftperrors.ErrMsgCannotBeNil)
	}
	if conn == nil {
		return nil, ftperrors.NewInvalidArgumentError("conn", ftperrors.ErrMsgCannotBeNil)
	}
	if textConn == nil {
		return nil, ftperrors.NewInvalidArgumentError("textConn", ftperrors.ErrMsgCannotBeNil)
	}

	sc := &ServerConnection{
		host:        host,
		dialer:      dialer,
		tcpConn:     conn,
		conn:        textConn,
		parser:      parsers.NewGenericListParser(),
		features:    &serverFeatures{},
		shutTimeout: defaultShutTimeout,
	}

	for _, opt := range options {
		if err := opt(sc); err != nil {
			return nil, err
		}
	}

	return sc, nil
}

// cmd function executes a command and validates the expected response code.
func (c *ServerConnection) cmd(expectedStatusCode int, format string, args ...any) (code int, msg string, err error) {
	if _, err = c.conn.Cmd(format, args...); err != nil {
		return 0, "", err
	}
	return c.conn.ReadResponse(expectedStatusCode)
}

func (c *ServerConnection) cmdWithDataConn(
	ctx context.Context,
	offset uint,
	format string,
	args ...any,
) (conn io.ReadWriteCloser, err error) {
	// For more information on PRET see: https://datatracker.ietf.org/doc/html/draft-dd-pret-00
	if c.features.supportPRET {
		_, _, cmdErr := c.cmd(StatusCommandOK, fmt.Sprintf(CommandPreTransfer, format), args...)
		if cmdErr != nil {
			return nil, ftperrors.NewInternalError("failed to issue pre-transfer configuration", cmdErr)
		}
	}

	tcpConn, err := c.openDataConn(ctx)
	if err != nil {
		return nil, err
	}

	if offset != 0 {
		if _, _, cmdErr := c.cmd(StatusRequestFilePending, CommandRestartTransfer, offset); cmdErr != nil {
			defer func() {
				if closeErr := tcpConn.Close(); closeErr != nil {
					err = closeErr
				}
			}()
			return nil, ftperrors.NewInternalError("failed to restart file transport from specified offset", cmdErr)
		}
	}

	code, msg, err := c.cmd(StatusNoCheck, format, args...)
	if err != nil {
		defer func() {
			if closeErr := tcpConn.Close(); closeErr != nil {
				err = closeErr
			}
		}()
		return nil, err
	}

	if code != StatusAlreadyOpen && code != StatusAboutToSend {
		defer func() {
			if closeErr := tcpConn.Close(); closeErr != nil {
				err = closeErr
			}
		}()
		return nil, ftperrors.NewInternalError(msg, nil)
	}

	// wrap newly establish connection connection
	conn = c.wrapConnection(tcpConn)

	return conn, nil
}

// setPassiveMode function sets server into passive mode retrieving host and port for
// future data connection.
func (c *ServerConnection) setPassiveMode() (string, error) {
	// Response of the command: Entering Passive Mode (h1,h2,h3,h4,p1,p2).
	_, msg, err := c.cmd(StatusPassiveMode, CommandPassive)
	if err != nil {
		return "", ftperrors.NewInternalError("failed to set passive mode", err)
	}
	msg = strings.ToLower(msg)
	if !strings.HasPrefix(msg, "entering passive mode") {
		return "", ftperrors.NewInternalError("invalid format of passive mode message", nil)
	}
	start := strings.Index(msg, "(")
	end := strings.Index(msg, ")")
	if start < 0 || end < 0 {
		return "", ftperrors.NewInternalError("failed to extract host and port from passive mode message", nil)
	}
	const tokenSize = 6
	hostPortTokens := strings.SplitN(msg[start+1:end], ",", tokenSize)

	host := strings.Join(hostPortTokens[:4], ".")

	port1, err := strconv.ParseUint(hostPortTokens[4], decimalBase, bitSize)
	if err != nil {
		return "", ftperrors.NewInternalError("failed to parse first part of connection port", err)
	}

	port2, err := strconv.ParseUint(hostPortTokens[5], decimalBase, bitSize)
	if err != nil {
		return "", ftperrors.NewInternalError("failed to parse second part of connection port", err)
	}

	port := port1*256 + port2

	return fmt.Sprintf("%s:%d", host, port), nil
}

// setPassiveMode function sets server into extended passive mode retrieving port for
// future data connection.
func (c *ServerConnection) setExtendedPassiveMode() (string, error) {
	_, msg, err := c.cmd(StatusExtendedPassiveMode, CommandExtendedPassiveMode)
	if err != nil {
		return "", ftperrors.NewInternalError("failed to set extended passive mode", err)
	}
	msg = strings.ToLower(msg)
	if !strings.HasPrefix(msg, "entering extended passive mode") {
		return "", ftperrors.NewInternalError("invalid format of extended passive mode message", nil)
	}
	start := strings.Index(msg, "(")
	end := strings.Index(msg, ")")
	if start < 0 || end < 0 {
		return "", ftperrors.NewInternalError("failed to extract port from extended passive mode message", nil)
	}
	port := strings.ReplaceAll(msg[start+1:end], "|", "")

	if _, parseErr := strconv.ParseUint(port, decimalBase, bitSize); parseErr != nil {
		return "", ftperrors.NewInternalError("failed to parse connection port", parseErr)
	}

	return port, nil
}

// getDataConnPort function retrieves data connection port by setting server into extended or standard
// passive mode.
func (c *ServerConnection) getDataConnPort() (string, error) {
	if !c.disableEPSV && c.features.supportEPSV {
		port, err := c.setExtendedPassiveMode()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%s", c.host, port), nil
	}
	return c.setPassiveMode()
}

// openDataConn function opens a new connection on address provided dynamically by the server.
func (c *ServerConnection) openDataConn(ctx context.Context) (net.Conn, error) {
	address, err := c.getDataConnPort()
	if err != nil {
		return nil, err
	}
	// TODO: add custom dial function
	if c.tlsConfig != nil {
		return c.dialer.DialContextTLS(ctx, "tcp", address, c.tlsConfig)
	}
	return c.dialer.DialContext(ctx, "tcp", address)
}

// checkDataConnShut function validates whether data connection is closed.
func (c *ServerConnection) checkDataConnShut() error {
	if c.shutTimeout != 0 {
		shutDeadline := time.Now().Add(c.shutTimeout)
		if err := c.tcpConn.SetDeadline(shutDeadline); err != nil {
			return err
		}
	}
	_, _, err := c.conn.ReadResponse(StatusClosingDataConnection)
	return err
}

// wrapConnection function wraps TCP connection with verbose writer providing the ability
// to debug incoming/outgoing server communications.
func (c *ServerConnection) wrapConnection(conn net.Conn) io.ReadWriteCloser {
	if c.verboseWriter == nil {
		return conn
	}
	return newVerboseConnectionWrapper(conn, c.verboseWriter)
}
