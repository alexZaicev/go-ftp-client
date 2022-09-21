package ftp_test

import (
	"fmt"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	dirPath              = "/foo/bar/baz"
	remoteDirPath        = "/doo/dee/daa"
	fileName             = "foobarbaz.txt"
	sizeInBytes   uint64 = 587

	dateFormat = "2006-01-02 15:04"
)

var (
	remotePathNoDir   = fileName
	remotePathWithDir = fmt.Sprintf("%s/%s", remoteDirPath, fileName)
)

func newEntry(t *testing.T, name string, sizeInBytes uint64, dateStr string) *entities.Entry {
	date, err := time.Parse(dateFormat, dateStr)
	require.NoError(t, err, "Failed to parse test case date")

	return &entities.Entry{
		Type:                 entities.EntryTypeFile,
		Permissions:          "rwxrwxrwx",
		Name:                 name,
		OwnerUser:            "user01",
		OwnerGroup:           "group01",
		SizeInBytes:          sizeInBytes,
		NumHardLinks:         2,
		LastModificationDate: date,
	}
}
