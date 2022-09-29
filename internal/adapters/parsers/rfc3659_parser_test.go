package parsers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/parsers"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
)

func Test_rfc3659ListParser_Success(t *testing.T) {
	// TODO: add below test cases:
	// modify=20150813224845;perm=fle;type=cdir;unique=119FBB87U4;UNIX.group=0;UNIX.mode=0755;UNIX.owner=0; .
	// modify=20150814172949;perm=flcdmpe;type=dir;unique=85A0C168U4;UNIX.group=0;UNIX.mode=0777;UNIX.owner=0; _upload
	// Modify=20150813175250;Perm=adfr;Size=951;Type=file;Unique=119FBB87UE;UNIX.group=0;UNIX.mode=0644;UNIX.owner=0; welcome.msg

	testCases := []struct {
		name          string
		input         string
		expectedEntry *entities.Entry
	}{
		{
			name:  "file entry",
			input: "Type=file;Size=1024990;Perm=r; /tmp/cap60.pl198.tar.gz",
			expectedEntry: &entities.Entry{
				Type:                 entities.EntryTypeFile,
				Permissions:          "r",
				SizeInBytes:          1024990,
				LastModificationDate: time.Time{},
				Name:                 "/tmp/cap60.pl198.tar.gz",
			},
		},
		{
			name:  "hidden . directory",
			input: "Type=cdir;Modify=19981107085215;Perm=el; /tmp",
			expectedEntry: &entities.Entry{
				Type:                 entities.EntryTypeDir,
				Permissions:          "el",
				LastModificationDate: time.Date(1998, 11, 7, 8, 52, 15, 0, time.UTC),
				Name:                 "/tmp",
			},
		},
		{
			name:  "hidden . directory",
			input: "Type=pdir;Modify=19981107085215;Perm=el; ..",
			expectedEntry: &entities.Entry{
				Type:                 entities.EntryTypeDir,
				Permissions:          "el",
				LastModificationDate: time.Date(1998, 11, 7, 8, 52, 15, 0, time.UTC),
				Name:                 "..",
			},
		},
		{
			name:  "hidden . directory",
			input: "Type=file;Size=1024990;Modify=19981107085215;Perm=r; cap60.pl198.tar.gz",
			expectedEntry: &entities.Entry{
				Type:                 entities.EntryTypeFile,
				Permissions:          "r",
				SizeInBytes:          1024990,
				LastModificationDate: time.Date(1998, 11, 7, 8, 52, 15, 0, time.UTC),
				Name:                 "cap60.pl198.tar.gz",
			},
		},
		{
			name:  "hidden . directory",
			input: "Type=file;Size=1024990;Modify=19981107085215;Perm=; cap60.pl198.tar.gz",
			expectedEntry: &entities.Entry{
				Type:                 entities.EntryTypeFile,
				SizeInBytes:          1024990,
				LastModificationDate: time.Date(1998, 11, 7, 8, 52, 15, 0, time.UTC),
				Name:                 "cap60.pl198.tar.gz",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parsers.NewGenericListParser()
			actual, err := p.Parse(tc.input, &parsers.Options{
				Location: time.UTC,
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedEntry, actual)
		})
	}
}

func Test_rfc3659ListParser_Errors(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "missing entry name",
			input: "Type=file;Size=1024990;Perm=r;",
		},
		{
			name:  "missing entry metadata",
			input: "/tmp",
		},
		{
			name:  "missing metadata name with assign char",
			input: "Type=file;1024990;Perm=r; /tmp",
		},
		{
			name:  "missing metadata name",
			input: "Type=file;=1024990;Perm=r; /tmp",
		},
		{
			name:  "invalid entry type",
			input: "Type=not-valid-type;Size=1024990;Perm=r; /tmp",
		},
		{
			name:  "invalid entry size in bytes",
			input: "Type=file;Size=not-valid-size;Perm=r; /tmp",
		},
		{
			name:  "invalid entry last modify date",
			input: "Type=file;Size=1024990;Modify=not-valid;Perm=r; /tmp",
		},
		{
			name:  "unknown metadata key=value",
			input: "Type=file;Size=1024990;Perm=r;Foo=bar; /tmp",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parsers.NewGenericListParser()
			actual, err := p.Parse(tc.input, &parsers.Options{
				Location: time.UTC,
			})
			assert.Nil(t, actual)
			assert.EqualError(t, err, "an internal error occurred: unsupported entry format")
		})
	}
}
