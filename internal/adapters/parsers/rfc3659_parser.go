package parsers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	rfc3659LastModificationDateFormat = "20060102150405"
)

type Metadata string

const (
	MetadataType             Metadata = "type"
	MetadataSize             Metadata = "size"
	MetadataPermissions      Metadata = "perm"
	MetadataLastModifiedDate Metadata = "modify"
)

type MetadataEntryType string

const (
	MetadataEntryTypeFile      MetadataEntryType = "file"
	MetadataEntryTypeDir       MetadataEntryType = "dir"
	MetadataEntryTypeListedDir MetadataEntryType = "cdir"
	MetadataEntryTypeParentDir MetadataEntryType = "pdir"
	// TODO sort out OS.name=type values listed -> https://www.rfc-editor.org/rfc/rfc3659#section-7.7.4
	// TODO investigate more types related to OS.name=type
)

type rfc3659ListParser struct {
}

func (p *rfc3659ListParser) Parse(data string, options *Options) (entry *entities.Entry, err error) {
	entry = &entities.Entry{}

	const tokenSize = 2
	tokens := strings.SplitN(data, " ", tokenSize)
	if len(tokens) != tokenSize {
		return nil, ftperrors.NewInternalError("invalid format of RFC3659 list entry", nil)
	}
	entry.Name = tokens[1]

	metadata := strings.Split(strings.ToLower(tokens[0]), ";")
	for _, md := range metadata {
		if md == "" {
			continue
		}
		mdTokens := strings.SplitN(md, "=", tokenSize)
		if len(mdTokens) != tokenSize {
			return nil, ftperrors.NewInternalError(
				fmt.Sprintf("invalid metadata key=value format: %s", md),
				nil,
			)
		}
		if mdTokens[0] == "" {
			return nil, ftperrors.NewInternalError(
				fmt.Sprintf("metadata key cannot be blank: %s", md),
				nil,
			)
		}
		mdName := mdTokens[0]
		mdValue := mdTokens[1]

		switch Metadata(mdName) {
		case MetadataType:
			entryType, convertErr := p.entryTypeFromMetadata(mdValue)
			if convertErr != nil {
				return nil, convertErr
			}
			entry.Type = entryType
		case MetadataSize:
			sizeInByte, convertErr := strconv.ParseUint(mdValue, decimalBase, bitSize)
			if convertErr != nil {
				return nil, ftperrors.NewInternalError("failed to parse size in bytes", convertErr)
			}
			entry.SizeInBytes = sizeInByte
		case MetadataLastModifiedDate:
			modifyDate, convertErr := time.ParseInLocation(rfc3659LastModificationDateFormat, mdValue, options.Location)
			if convertErr != nil {
				return nil, ftperrors.NewInternalError("failed to parse last modification date", convertErr)
			}
			entry.LastModificationDate = modifyDate
		case MetadataPermissions:
			entry.Permissions = mdValue
		default:
			return nil, ftperrors.NewUnknownError(
				fmt.Sprintf("unexpected entry metadata: %s", md),
				nil,
			)
		}
	}

	return entry, nil
}

func (p *rfc3659ListParser) entryTypeFromMetadata(entryType string) (entities.EntryType, error) {
	switch MetadataEntryType(entryType) {
	case MetadataEntryTypeDir, MetadataEntryTypeParentDir, MetadataEntryTypeListedDir:
		return entities.EntryTypeDir, nil
	case MetadataEntryTypeFile:
		return entities.EntryTypeFile, nil
	default:
		return entities.EntryType(0), ftperrors.NewUnknownError(
			fmt.Sprintf("unexpected entry type: %s", entryType),
			nil,
		)
	}
}
