package ftpclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/connection"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	DateFormat = "Mon, 02 Jan 15:04"
)

type SizeType uint

const (
	KiB SizeType = 1 << 10
	MiB SizeType = 1 << 20
	GiB SizeType = 1 << 30
	TiB SizeType = 1 << 40
)

type CallbackWriter struct {
	Callback func(bytesRead int64)
}

func (cw *CallbackWriter) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	cw.Callback(int64(n))
	return
}

func Connect(
	ctx context.Context,
	address string,
	user string,
	password string,
	timeout time.Duration,
	verbose bool,
) (conn connection.Connection, err error) {
	var vw io.Writer
	if verbose {
		vw = os.Stdout
	}

	conn, err = ftpconnection.Dial(
		ctx,
		address,
		ftpconnection.WithTimeout(timeout),
		ftpconnection.WithVerboseWriter(vw),
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to establish connection", err)
	}

	if loginErr := conn.Login(user, password); loginErr != nil {
		defer func() {
			if stopErr := conn.Stop(); stopErr != nil {
				err = stopErr
			}
		}()
		return nil, errors.NewInternalError("failed to authenticate with provided user account", loginErr)
	}

	return conn, nil
}

func EntryTypeToStr(entryType entities.EntryType) (string, error) {
	switch entryType {
	case entities.EntryTypeFile:
		return "F", nil
	case entities.EntryTypeDir:
		return "D", nil
	case entities.EntryTypeLink:
		return "L", nil
	default:
		return "", errors.NewUnknownError(
			fmt.Sprintf("unexpected entry type: %d", entryType),
			nil,
		)
	}
}

func FormatSizeInBytes(bytes uint64) string {
	postfixSizeMap := map[string]SizeType{
		"KB": KiB,
		"MB": MiB,
		"GB": GiB,
		"TB": TiB,
	}

	for k, v := range postfixSizeMap {
		converted := float64(bytes) / float64(v)
		if converted < 1 {
			continue
		}
		return fmt.Sprintf("%.2f %s", converted, k)
	}
	return fmt.Sprintf("%d B", bytes)
}
