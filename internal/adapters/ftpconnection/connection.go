package ftpconnection

import (
	"bufio"
	"crypto/tls"
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
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
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

	Off = "OFF"
	On  = "ON"

	FeatureMLST = "MLST"
	FeatureMDTM = "MDTM"
	FeatureMFMT = "MFMT"
	FeaturePRET = "PRET"
	FeatureUTF8 = "UTF8"
)

type TransferType string

const (
	TransferTypeASCII  = "A"
	TransferTypeBinary = "I"
)

type serverConnection struct {
	dialOptions *dialOptions
	host        string
	conn        *textproto.Conn
	tcpConn     net.Conn
	parser      parsers.Parser

	// server features
	mlstSupported bool
	mdtmSupported bool
	mfmtSupported bool
	usePRET       bool
	mdtmCanWrite  bool
	skipEPSV      bool
}

func newConnection(conn net.Conn, options *dialOptions) connection.Connection {
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	return &serverConnection{
		dialOptions: options,
		host:        remoteAddr.IP.String(),
		tcpConn:     conn,
		conn:        textproto.NewConn(options.wrapConnection(conn)),
		parser:      parsers.NewGenericListParser(),
	}
}

// Ready function validates that the FTP server is ready to proceed.
func (c *serverConnection) Ready() error {
	if _, _, err := c.conn.ReadResponse(StatusReady); err != nil {
		defer c.Stop()
		return errors.NewInternalError("failed to check if FTP server is ready", err)
	}
	return nil
}

// EnableExplicitTLSMode function enables TLS modes on established TCP connection.
func (c *serverConnection) EnableExplicitTLSMode() error {
	if _, _, err := c.cmd(StatusAuthOK, CommandAuthTLS); err != nil {
		defer c.Stop()
		return errors.NewInternalError("failed to authenticate with TLS", err)
	}
	tlsConn := tls.Client(c.tcpConn, c.dialOptions.tlsConfig)
	c.tcpConn = tlsConn
	c.conn = textproto.NewConn(c.dialOptions.wrapConnection(tlsConn))
	return nil
}

// Stop function sends a quit command to FTP server and closes the TCP connection.
func (c *serverConnection) Stop() error {
	defer c.conn.Close()
	if _, err := c.conn.Cmd(CommandQuit); err != nil {
		return errors.NewInternalError("failed to send quit command to FTP server", err)
	}
	return nil
}

// Login function authenticate user with provided account username and password. Upon successful authentication,
// server is then queried to list supported features to update connection settings at runtime.
func (c *serverConnection) Login(user, password string) error {
	code, msg, err := c.cmd(StatusNoCheck, CommandUser, user)
	if err != nil {
		return errors.NewInternalError("failed to send user command", err)
	}

	switch code {
	case StatusLoggedIn:
	case StatusUserOK:
		if _, _, err := c.cmd(StatusLoggedIn, CommandPass, password); err != nil {
			return errors.NewInternalError("failed to authenticate user", err)
		}
	default:
		return errors.NewInternalError(msg, nil)
	}

	if err := c.updateFeatures(); err != nil {
		return errors.NewInternalError("failed to establish supported features of connected server", err)
	}

	return nil
}

func (c *serverConnection) List(options *connection.ListOptions) ([]*entities.Entry, error) {
	if options == nil {
		return nil, errors.NewInvalidArgumentError("options", errors.ErrMsgCannotBeNil)
	}

	cmd := CommandList
	if c.mlstSupported {
		cmd = CommandListMachineReadable
	} else if options.ShowAll {
		cmd = CommandListHidden
	}

	conn, err := c.cmdWithDataConn(0, cmd, options.Path)
	if err != nil {
		return nil, errors.NewInternalError("failed to list files", err)
	}
	defer conn.Close()

	entries := make([]*entities.Entry, 0)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		entryStr := scanner.Text()
		entry, parseErr := c.parser.Parse(entryStr)
		if parseErr != nil {
			return nil, errors.NewInternalError(fmt.Sprintf("failed to parser list entry: %s", entryStr), parseErr)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (c *serverConnection) Mkdir(options *connection.MkdirOptions) error {
	if options == nil {
		return errors.NewInvalidArgumentError("options", errors.ErrMsgCannotBeNil)
	}

	pathTokens := strings.Split(options.Path, string(filepath.Separator))
	var builder strings.Builder
	for idx, pathToken := range pathTokens {
		if idx > 0 {
			builder.WriteRune(filepath.Separator)
		}
		builder.WriteString(pathToken)
		_, _, err := c.cmd(StatusPathCreated, CommandMakeDir, builder.String())
		if err != nil {
			return errors.NewInternalError("failed to create directory", err)
		}
	}
	return nil
}

func (c *serverConnection) Status() (*entities.Status, error) {
	_, msg, err := c.cmd(StatusSystem, CommandStatus)
	if err != nil {
		return nil, errors.NewInternalError("failed to get server status", err)
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
			tokens := strings.SplitN(line, " ", 3)
			status.RemoteAddress = tokens[2]
			continue
		}

		if strings.Contains(line, "logged in") {
			tokens := strings.SplitN(line, " ", 4)
			status.LoggedInUser = tokens[3]
			continue
		}
		// TODO: add status check TLS
	}

	_, msg, err = c.cmd(StatusName, CommandSystem)
	if err != nil {
		return nil, errors.NewInternalError("failed to get server system information", err)
	}

	msg = strings.TrimSpace(msg)
	tokens := strings.SplitN(msg, " ", 2)
	status.System = tokens[0]

	return status, nil
}

func (c *serverConnection) Upload(options *connection.UploadOptions) error {
	if options == nil {
		return errors.NewInvalidArgumentError("options", errors.ErrMsgCannotBeNil)
	}

	conn, err := c.cmdWithDataConn(0, CommandStore, options.Path)
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
		return errors.NewInternalError("failed to upload file to remote path", err)
	}

	return nil
}

func (c *serverConnection) Cd(path string) error {
	_, _, err := c.cmd(StatusRequestedFileActionOK, CommandChangeWorkDir, path)
	if err != nil {
		return errors.NewInternalError("failed to change working directory", err)
	}
	return nil
}

// cmd function executes a command and validates the expected response code.
func (c *serverConnection) cmd(expectedStatusCode int, format string, args ...any) (int, string, error) {
	if _, err := c.conn.Cmd(format, args...); err != nil {
		return 0, "", err
	}
	return c.conn.ReadResponse(expectedStatusCode)
}

func (c *serverConnection) cmdWithDataConn(offset uint, format string, args ...any) (net.Conn, error) {
	// For more information on PRET see: https://datatracker.ietf.org/doc/html/draft-dd-pret-00
	if c.usePRET {
		_, _, err := c.cmd(StatusCommandOK, fmt.Sprintf(CommandPreTransfer, format), args...)
		if err != nil {
			return nil, errors.NewInternalError("failed to issue pre-transfer configuration", err)
		}
	}

	conn, err := c.openDataConn()
	if err != nil {
		return nil, errors.NewInternalError("failed to open new data connection", err)
	}

	if offset != 0 {
		if _, _, err := c.cmd(StatusRequestFilePending, CommandRestartTransfer, offset); err != nil {
			defer conn.Close()
			return nil, errors.NewInternalError("failed to restart file transport from specified offset", err)
		}
	}

	code, msg, err := c.cmd(StatusNoCheck, format, args...)
	if err != nil {
		defer conn.Close()
		return nil, err
	}

	if code != StatusAlreadyOpen && code != StatusAboutToSend {
		defer conn.Close()
		return nil, errors.NewInternalError(msg, nil)
	}

	return conn, nil
}

// updateFeatures function queries FTP server for supported features and adjusts connection settings
// based on user and received settings.
func (c *serverConnection) updateFeatures() error {
	code, msg, err := c.cmd(StatusNoCheck, CommandFeat)
	if err != nil {
		return errors.NewInternalError(msg, err)
	}

	if code != StatusSystem {
		// The server does not support the FEAT command. This is not an
		// error: we consider that there is no additional feature.
		return nil
	}

	features := make(map[string]string, 0)
	for _, line := range strings.Split(msg, "\n") {
		if strings.Contains(line, "Features") || strings.Contains(line, "End") {
			continue
		}
		line = strings.TrimSpace(line)
		feature := strings.SplitN(line, " ", 2)

		command := feature[0]
		var commandDesc string
		if len(feature) == 2 {
			commandDesc = feature[1]
		}
		features[command] = commandDesc
	}

	if _, ok := features[FeatureMLST]; ok && !c.dialOptions.disableMLST {
		c.mlstSupported = true
	}
	_, c.usePRET = features[FeaturePRET]
	_, c.mdtmSupported = features[FeatureMDTM]
	_, c.mfmtSupported = features[FeatureMFMT]
	c.mdtmCanWrite = c.mdtmSupported && c.dialOptions.writingMDTM

	// switch to binary mode
	if _, _, err := c.cmd(StatusCommandOK, CommandType, TransferTypeBinary); err != nil {
		return errors.NewInternalError("failed to set transfer type to binary mode", err)
	}

	if _, ok := features[FeatureUTF8]; ok && !c.dialOptions.disableUTF8 {
		if err := c.setUTF8(); err != nil {
			return errors.NewInternalError("failed to turn UTF-8 option on", err)
		}
	}

	// If using implicit TLS, make data connections also use TLS
	if c.dialOptions.tlsConfig != nil {
		if _, _, err = c.cmd(StatusCommandOK, CommandProtectionBufferSize); err != nil {
			return errors.NewInternalError("failed to set protocol buffer size", err)
		}
		if _, _, err = c.cmd(StatusCommandOK, CommandProtocol); err != nil {
			return errors.NewInternalError("failed to set protocol", err)
		}
	}

	return nil
}

// setUTF8 function sets UTF-8 format on connected server. If server does not support this option,
// it's ignored.
func (c *serverConnection) setUTF8() error {
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
		return errors.NewInternalError(msg, nil)
	}
	return nil
}

// setPassiveMode function sets server into passive mode retrieving host and port for
// future data connection.
func (c *serverConnection) setPassiveMode() (string, error) {
	// Response of the command: Entering Passive Mode (h1,h2,h3,h4,p1,p2).
	_, msg, err := c.cmd(StatusPassiveMode, CommandPassive)
	if err != nil {
		return "", errors.NewInternalError("failed to set passive mode on server", err)
	}
	msg = strings.ToLower(msg)
	if !strings.HasPrefix(msg, "entering passive mode") {
		return "", errors.NewInternalError("failed to extract host and port from passive mode response", nil)
	}
	start := strings.Index(msg, "(")
	end := strings.Index(msg, ")")
	if start < 0 || end < 0 {
		return "", errors.NewInternalError("failed to extract host and port from passive mode response", nil)
	}
	hostPortTokens := strings.SplitN(msg[start+1:end], ",", 6)

	host := strings.Join(hostPortTokens[:4], ".")

	port1, err := strconv.ParseUint(hostPortTokens[4], 10, 64)
	if err != nil {
		return "", errors.NewInternalError("failed to parse first part of passive port", err)
	}

	port2, err := strconv.ParseUint(hostPortTokens[5], 10, 64)
	if err != nil {
		return "", errors.NewInternalError("failed to parse second part of passive port", err)
	}

	port := port1*256 + port2

	return fmt.Sprintf("%s:%d", host, port), nil
}

// setPassiveMode function sets server into extended passive mode retrieving port for
// future data connection.
func (c *serverConnection) setExtendedPassiveMode() (string, error) {
	_, msg, err := c.cmd(StatusExtendedPassiveMode, CommandExtendedPassiveMode)
	if err != nil {
		return "", errors.NewInternalError("failed to set extended passive mode", err)
	}
	msg = strings.ToLower(msg)
	if !strings.HasPrefix(msg, "entering extended passive mode") {
		return "", errors.NewInternalError("failed to extract port from extended passive mode response", nil)
	}
	start := strings.Index(msg, "(")
	end := strings.Index(msg, ")")
	if start < 0 || end < 0 {
		return "", errors.NewInternalError("failed to extract port from extended passive mode response", nil)
	}
	port := strings.ReplaceAll(msg[start+1:end], "|", "")

	if _, err := strconv.ParseUint(port, 10, 64); err != nil {
		return "", errors.NewInternalError("failed to parse passive port", err)
	}

	return port, nil
}

// getDataConnPort function retrieves data connection port by setting server into extended or standard
// passive mode.
func (c *serverConnection) getDataConnPort() (string, error) {
	if !c.dialOptions.disableEPSV && !c.skipEPSV {
		port, err := c.setExtendedPassiveMode()
		if err != nil {
			return "", err
		}
		c.skipEPSV = true
		return fmt.Sprintf("%s:%s", c.host, port), nil
	}
	return c.setPassiveMode()
}

// openDataConn function opens a new connection on address provided dynamically by the server.
func (c *serverConnection) openDataConn() (net.Conn, error) {
	address, err := c.getDataConnPort()
	if err != nil {
		return nil, err
	}
	// TODO: add custom dial function
	if c.dialOptions.tlsConfig != nil {
		return tls.DialWithDialer(c.dialOptions.dialer, "tcp", address, c.dialOptions.tlsConfig)
	}
	return c.dialOptions.dialer.Dial("tcp", address)
}

// checkDataConnShut function validates whether data connection is closed.
func (c *serverConnection) checkDataConnShut() error {
	if c.dialOptions.shutTimeout != 0 {
		shutDeadline := time.Now().Add(c.dialOptions.shutTimeout)
		if err := c.tcpConn.SetDeadline(shutDeadline); err != nil {
			return err
		}
	}
	_, _, err := c.conn.ReadResponse(StatusClosingDataConnection)
	return err
}
