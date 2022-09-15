package parsers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	lastModificationDateFormat = "Jan 2 15:04"
)

type unixListParser struct {
}

func (p *unixListParser) Parse(data string) (*entities.Entry, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, errors.NewInvalidArgumentError("data", errors.ErrMsgCannotBeBlank)
	}

	entry := &entities.Entry{}
	var token string

	// entry type and permissions extraction
	data, token = p.nextToken(data)
	entryType, err := p.getEntryType(token[:1])
	if err != nil {
		return nil, err
	}
	entry.Type = entryType
	entry.Permissions = token[1:]

	// hard links
	data, token = p.nextToken(data)
	numLinks, err := strconv.ParseInt(token, 10, 32)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse number of hard links", err)
	}
	entry.NumHardLinks = int(numLinks)

	// owner user and group
	data, entry.OwnerUser = p.nextToken(data)
	data, entry.OwnerGroup = p.nextToken(data)

	// size in bytes
	data, token = p.nextToken(data)
	sizeInBytes, err := strconv.ParseUint(token, 10, 64)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse size in bytes", err)
	}
	entry.SizeInBytes = sizeInBytes

	// last modification date
	dateTokens := make([]string, 0, 3)
	for idx := 0; idx < 3; idx++ {
		data, token = p.nextToken(data)
		dateTokens = append(dateTokens, token)
	}
	dateStr := strings.Join(dateTokens, " ")
	lastModificationDate, err := time.Parse(lastModificationDateFormat, dateStr)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse last modification date", err)
	}
	entry.LastModificationDate = lastModificationDate

	// name
	data, entry.Name = p.nextToken(data)

	return entry, nil
}

func (p *unixListParser) nextToken(data string) (newData, token string) {
	var start int
	var end int

	var startFound bool
	for idx, ch := range data {
		if ch == ' ' {
			if startFound {
				end = idx
				break
			}
			continue
		}

		if startFound && idx == len(data)-1 {
			end = idx + 1
			break
		}

		if !startFound {
			startFound = true
			start = idx
			continue
		}
	}

	if start >= end {
		end = len(data)
	}

	return strings.TrimSpace(data[end:]), strings.TrimSpace(data[start:end])
}

func (p *unixListParser) getEntryType(val string) (entities.EntryType, error) {
	switch val {
	case "-":
		return entities.EntryTypeFile, nil
	case "d":
		return entities.EntryTypeDir, nil
	case "l":
		return entities.EntryTypeLink, nil
	default:
		return entities.EntryType(0), errors.NewUnknownError(
			fmt.Sprintf("unexpected entry type: %s", val),
			nil,
		)
	}
}
