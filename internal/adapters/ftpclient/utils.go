package ftpclient

import (
	"fmt"

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

type SizePostfix string

const (
	B  SizePostfix = "B"
	KB SizePostfix = "KB"
	MB SizePostfix = "MB"
	GB SizePostfix = "GB"
	TB SizePostfix = "TB"
)

type CallbackWriter struct {
	Callback func(bytesRead int64)
}

func (cw *CallbackWriter) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	cw.Callback(int64(n))
	return
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
	postfixMap := map[SizePostfix]SizeType{
		TB: TiB,
		GB: GiB,
		MB: MiB,
		KB: KiB,
	}

	for _, postfix := range []SizePostfix{TB, GB, MB, KB} {
		bits, _ := postfixMap[postfix]
		converted := float64(bytes) / float64(bits)
		if converted < 1 {
			continue
		}
		return fmt.Sprintf("%.2f %s", converted, postfix)
	}
	return fmt.Sprintf("%d %s", bytes, B)
}
