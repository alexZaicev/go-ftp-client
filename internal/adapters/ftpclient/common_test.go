package ftpclient_test

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	anonymous = "anonymous"
)

type ftpMock struct {
	t        *testing.T
	address  string
	listener *net.TCPListener
	proto    *textproto.Conn
	commands []string // list of received commands
	lastFull string   // full last command
	sync.WaitGroup
}

// newFtpMock returns a mock implementation of a FTP server
func newFtpMock(t *testing.T) (*ftpMock, error) {
	var err error
	mock := &ftpMock{
		t: t,
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	mock.address = l.Addr().String()

	tcpListener, ok := l.(*net.TCPListener)
	if !ok {
		return nil, errors.New("listener is not a net.TCPListener")
	}
	mock.listener = tcpListener

	go mock.listen()
	return mock, nil
}

func (mock *ftpMock) listen() {
	// Listen for an incoming connection.
	conn, err := mock.listener.Accept()
	if err != nil {
		mock.t.Errorf("can not accept: %s", err)
		return
	}

	// Do not accept incoming connections anymore
	mock.listener.Close()

	mock.Add(1)
	defer mock.Done()
	defer conn.Close()

	mock.proto = textproto.NewConn(conn)
	mock.printLinef("220 FTP Server ready.")

	for {
		fullCommand, _ := mock.proto.ReadLine()
		mock.lastFull = fullCommand

		cmdParts := strings.Split(fullCommand, " ")

		// Append to list of received commands
		mock.commands = append(mock.commands, cmdParts[0])

		// At least one command must have a multiline response
		switch cmdParts[0] {
		case "AUTH":
			mock.printLinef("234 Proceed with negotiation.")

			cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
			require.NoError(mock.t, err)

			mock.proto = textproto.NewConn(tls.Server(conn, &tls.Config{
				MinVersion: tls.VersionTLS12,
				Certificates: []tls.Certificate{
					cert,
				},
				Rand:               rand.Reader,
				Time:               time.Now,
				InsecureSkipVerify: true, //nolint:gosec // testing server
			}))
		case "FEAT":
			features := "211-Features:\r\n FEAT\r\n PASV\r\n EPSV\r\n UTF8\r\n SIZE\r\n MLST\r\n"
			features += "211 End"
			mock.printLinef(features)
		case "USER":
			if cmdParts[1] == anonymous {
				mock.printLinef("331 Please send your password")
			} else {
				mock.printLinef("530 This FTP server is anonymous only")
			}
		case "PASS":
			mock.printLinef("230-Hey,\r\nWelcome to my FTP\r\n230 Access granted")
		case "OPTS":
			if len(cmdParts) != 3 {
				mock.printLinef("500 wrong number of arguments")
				break
			}
			if (strings.Join(cmdParts[1:], " ")) == "UTF8 ON" {
				mock.printLinef("200 OK, UTF-8 enabled")
			}
		case "TYPE":
			mock.printLinef("200 Type set ok")
		case "PBSZ":
			mock.printLinef("200 PBSZ set to %s.", cmdParts[1])
		case "PROT":
			if cmdParts[1] != "P" {
				mock.printLinef("501 Bad arguments %s.", cmdParts[1])
				continue
			}
			mock.printLinef("200 PROT now Private.")
		default:
			mock.printLinef("500 Unknown command %s.", cmdParts[0])
		}
	}
}

func (mock *ftpMock) printLinef(format string, args ...interface{}) {
	if err := mock.proto.Writer.PrintfLine(format, args...); err != nil {
		mock.t.Fatal(err)
	}
}

func (mock *ftpMock) Addr() string {
	return mock.listener.Addr().String()
}

// Closes the listening socket
func (mock *ftpMock) Close() {
	mock.listener.Close()
}
