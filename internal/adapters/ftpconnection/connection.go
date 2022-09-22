package ftpconnection

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/parsers"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
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

// Ready function validates that the FTP server is ready to proceed.
func (c *ServerConnection) Ready() (err error) {
	if _, _, readErr := c.conn.ReadResponse(StatusReady); readErr != nil {
		defer func() {
			if stopErr := c.Stop(); stopErr != nil {
				err = stopErr
			}
		}()
		return ftperrors.NewInternalError("failed to check if server is ready", readErr)
	}
	return nil
}

// EnableExplicitTLSMode function enables TLS modes on established TCP connection.
func (c *ServerConnection) EnableExplicitTLSMode() (err error) {
	if _, _, readErr := c.cmd(StatusAuthOK, CommandAuthTLS); readErr != nil {
		defer func() {
			if stopErr := c.Stop(); stopErr != nil {
				err = stopErr
			}
		}()
		return ftperrors.NewInternalError("failed to enable explicit TLS mode", readErr)
	}
	tlsConn := tls.Client(c.tcpConn, c.tlsConfig)
	c.tcpConn = tlsConn
	c.conn = textproto.NewConn(c.wrapConnection(tlsConn))
	return nil
}

// Stop function sends a quit command to FTP server and closes the TCP connection.
func (c *ServerConnection) Stop() (err error) {
	defer func() {
		if closeErr := c.conn.Close(); closeErr != nil {
			err = ftperrors.NewInternalError("failed to close connection", closeErr)
		}
	}()
	if _, cmdErr := c.conn.Cmd(CommandQuit); cmdErr != nil {
		return ftperrors.NewInternalError("failed to disconnect from the server", cmdErr)
	}
	return nil
}

// Login function authenticate user with provided account username and password. Upon successful authentication,
// server is then queried to list supported features to update connection settings at runtime.
func (c *ServerConnection) Login(user, password string) error {
	code, msg, err := c.cmd(StatusNoCheck, CommandUser, user)
	if err != nil {
		return ftperrors.NewInternalError("failed to start username authentication", err)
	}

	switch code {
	case StatusLoggedIn:
	case StatusUserOK:
		if _, _, pwdErr := c.cmd(StatusLoggedIn, CommandPass, password); pwdErr != nil {
			return ftperrors.NewInternalError("failed to authenticate user", pwdErr)
		}
	default:
		return ftperrors.NewInternalError(msg, nil)
	}

	if updateErr := c.updateFeatures(); updateErr != nil {
		return updateErr
	}

	return nil
}

func (c *ServerConnection) List(ctx context.Context, options *connection.ListOptions) (entries []*entities.Entry, err error) {
	if options == nil {
		return nil, ftperrors.NewInvalidArgumentError("options", ftperrors.ErrMsgCannotBeNil)
	}

	cmd := CommandList
	if c.features.supportMLST {
		cmd = CommandListMachineReadable
	} else if options.ShowAll {
		cmd = CommandListHidden
	}

	conn, err := c.cmdWithDataConn(ctx, 0, cmd, options.Path)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to list files", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		entryStr := scanner.Text()
		entry, parseErr := c.parser.Parse(entryStr)
		if parseErr != nil {
			return nil, ftperrors.NewInternalError(fmt.Sprintf("failed to parser list entry: %s", entryStr), parseErr)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (c *ServerConnection) Mkdir(path string) error {
	if path == "" {
		return ftperrors.NewInvalidArgumentError("path", ftperrors.ErrMsgCannotBeBlank)
	}

	if strings.HasPrefix(path, string(filepath.Separator)) {
		path = path[1:]
	}
	pathTokens := strings.Split(path, string(filepath.Separator))

	var builder strings.Builder
	builder.WriteString("/")
	for idx, pathToken := range pathTokens {
		if idx > 0 {
			builder.WriteRune(filepath.Separator)
		}
		builder.WriteString(pathToken)

		pathToCreate := builder.String()

		if err := c.Cd(pathToCreate); err != nil {
			var notFoundErr *ftperrors.NotFoundError
			if !errors.As(err, &notFoundErr) {
				return err
			}

			_, _, mkdErr := c.cmd(StatusPathCreated, CommandMakeDir, pathToCreate)
			if mkdErr != nil {
				return ftperrors.NewInternalError("failed to create directory", mkdErr)
			}
		}
	}

	if err := c.Cd("/"); err != nil {
		return err
	}

	return nil
}

func (c *ServerConnection) Status() (*entities.Status, error) {
	_, msg, err := c.cmd(StatusSystem, CommandStatus)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to fetch server status", err)
	}

	status := &entities.Status{}

	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)
		if strings.Contains(line, "server status:") || strings.Contains(line, "end of status") {
			continue
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "connected to") {
			const tokenSize = 3
			tokens := strings.SplitN(line, " ", tokenSize)
			if len(tokens) >= tokenSize {
				status.RemoteAddress = tokens[2]
			}
			continue
		}

		if strings.Contains(line, "logged in") {
			const tokenSize = 4
			tokens := strings.SplitN(line, " ", tokenSize)
			if len(tokens) >= tokenSize {
				status.LoggedInUser = tokens[3]
			}
			continue
		}
		// TODO: add status check TLS
	}

	_, msg, err = c.cmd(StatusName, CommandSystem)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to fetch system type", err)
	}

	msg = strings.TrimSpace(msg)
	const tokenSize = 2
	tokens := strings.SplitN(msg, " ", tokenSize)
	status.System = tokens[0]

	return status, nil
}

func (c *ServerConnection) Upload(ctx context.Context, options *connection.UploadOptions) error {
	if options == nil {
		return ftperrors.NewInvalidArgumentError("options", ftperrors.ErrMsgCannotBeNil)
	}

	conn, err := c.cmdWithDataConn(ctx, 0, CommandStore, options.Path)
	if err != nil {
		return err
	}

	var multiErr *multierror.Error

	if _, err = io.Copy(conn, options.FileReader); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	// opened data connection needs to be closed prior to the bellow check
	if err = conn.Close(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	if err = c.checkDataConnShut(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	err = multiErr.ErrorOrNil()
	if err != nil {
		return ftperrors.NewInternalError("failed to upload file to remote path", err)
	}

	return nil
}

func (c *ServerConnection) Cd(path string) error {
	code, msg, err := c.cmd(StatusNoCheck, CommandChangeWorkDir, path)
	if err != nil {
		return ftperrors.NewInternalError("failed to change working directory", err)
	}
	if code == StatusRequestedFileActionOK {
		return nil
	}
	if code == StatusFileUnavailable {
		return ftperrors.NewNotFoundError(fmt.Sprintf("path %s does not exist", path), nil)
	}
	return ftperrors.NewInternalError(msg, nil)
}

func (c *ServerConnection) Size(path string) (uint64, error) {
	_, msg, err := c.cmd(StatusFile, CommandSize, path)
	if err != nil {
		return 0, ftperrors.NewInternalError("failed to fetch file size", err)
	}
	sizeInBytes, err := strconv.ParseUint(msg, decimalBase, bitSize)
	if err != nil {
		return 0, ftperrors.NewInternalError("failed to parse file size to a non-zero integer", err)
	}
	return sizeInBytes, err
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
		return nil, ftperrors.NewInternalError("failed to open new data connection", err)
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

// updateFeatures function queries FTP server for supported features and adjusts connection settings
// based on user and received settings.
func (c *ServerConnection) updateFeatures() error {
	code, msg, err := c.cmd(StatusNoCheck, CommandFeat)
	if err != nil {
		return ftperrors.NewInternalError("failed to list supported features", err)
	}

	if code != StatusSystem || msg == "" {
		// The server does not support the FEAT command. This is not an
		// error, as we consider that there is no additional features.
		return nil
	}

	features := c.getFeaturesMap(msg)
	c.features = newServerFeatures(features)

	// switch to binary mode
	if _, _, cmdErr := c.cmd(StatusCommandOK, CommandType, TransferTypeBinary); cmdErr != nil {
		return ftperrors.NewInternalError("failed to set binary transfer mode", cmdErr)
	}

	if c.features.supportUTF8 && !c.disableUTF8 {
		if utfErr := c.setUTF8(); utfErr != nil {
			return ftperrors.NewInternalError("failed to turn UTF-8 option on", utfErr)
		}
	}

	// If using implicit TLS, make data connections also use TLS
	if c.tlsConfig != nil {
		if _, _, err = c.cmd(StatusCommandOK, CommandProtectionBufferSize); err != nil {
			return ftperrors.NewInternalError("failed to set protocol buffer size", err)
		}
		if _, _, err = c.cmd(StatusCommandOK, CommandProtocol); err != nil {
			return ftperrors.NewInternalError("failed to enable TLS protocol", err)
		}
	}

	return nil
}

// getFeaturesMap function processes value parameter returned by the FEAT command and
// composes a map[COMMAND]COMMAND_DESC of supporter server features.
func (c *ServerConnection) getFeaturesMap(value string) map[string]string {
	features := make(map[string]string, 0)
	for _, line := range strings.Split(value, "\n") {
		loweredLine := strings.ToLower(line)
		if strings.Contains(loweredLine, "features") || strings.Contains(loweredLine, "end") {
			continue
		}
		line = strings.TrimSpace(line)
		const tokenSize = 2
		tokens := strings.SplitN(line, " ", tokenSize)

		var cmdDesc string
		if len(tokens) == tokenSize {
			cmdDesc = tokens[1]
		}
		features[tokens[0]] = cmdDesc
	}
	return features
}

// setUTF8 function sets UTF-8 format on connected server. If server does not support this option,
// it's ignored.
func (c *ServerConnection) setUTF8() error {
	code, msg, err := c.cmd(StatusNoCheck, CommandOptions, FeatureUTF8, On)
	if err != nil {
		return err
	}

	// Workaround for FTP servers, that does not support this option.
	if code == StatusBadArguments || code == StatusNotImplementedParameter {
		return nil
	}

	// The ftpd "filezilla-server" has FEAT support for UTF8, but always returns
	// "202 UTF8 mode is always enabled. No need to send this command." when
	// trying to use it. That's OK
	if code == StatusCommandNotImplemented {
		return nil
	}

	if code != StatusCommandOK {
		return errors.New(msg)
	}
	return nil
}

// setPassiveMode function sets server into passive mode retrieving host and port for
// future data connection.
func (c *ServerConnection) setPassiveMode() (string, error) {
	// Response of the command: Entering Passive Mode (h1,h2,h3,h4,p1,p2).
	_, msg, err := c.cmd(StatusPassiveMode, CommandPassive)
	if err != nil {
		return "", ftperrors.NewInternalError("failed to set passive mode on server", err)
	}
	msg = strings.ToLower(msg)
	if !strings.HasPrefix(msg, "entering passive mode") {
		return "", ftperrors.NewInternalError("failed to extract host and port from passive mode response", nil)
	}
	start := strings.Index(msg, "(")
	end := strings.Index(msg, ")")
	if start < 0 || end < 0 {
		return "", ftperrors.NewInternalError("failed to extract host and port from passive mode response", nil)
	}
	const tokenSize = 6
	hostPortTokens := strings.SplitN(msg[start+1:end], ",", tokenSize)

	host := strings.Join(hostPortTokens[:4], ".")

	port1, err := strconv.ParseUint(hostPortTokens[4], decimalBase, bitSize)
	if err != nil {
		return "", ftperrors.NewInternalError("failed to parse first part of passive port", err)
	}

	port2, err := strconv.ParseUint(hostPortTokens[5], decimalBase, bitSize)
	if err != nil {
		return "", ftperrors.NewInternalError("failed to parse second part of passive port", err)
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
		return "", ftperrors.NewInternalError("failed to extract port from extended passive mode response", nil)
	}
	start := strings.Index(msg, "(")
	end := strings.Index(msg, ")")
	if start < 0 || end < 0 {
		return "", ftperrors.NewInternalError("failed to extract port from extended passive mode response", nil)
	}
	port := strings.ReplaceAll(msg[start+1:end], "|", "")

	if _, err := strconv.ParseUint(port, decimalBase, bitSize); err != nil {
		return "", ftperrors.NewInternalError("failed to parse passive port", err)
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
	return c.dialer.Dial("tcp", address)
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
