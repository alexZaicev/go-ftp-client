package parsers

import (
	"strings"

	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

const (
	decimalBase = 10
	bitSize64   = 64
	bitSize32   = 32
)

type Parser interface {
	Parse(data string, options *Options) (*entities.Entry, error)
}

type genericListParser struct {
	parsers []Parser
}

func NewGenericListParser() Parser {
	return &genericListParser{
		parsers: []Parser{
			&hostedListParser{},
			&msDosListParser{},
			&unixListParser{},
			&rfc3659ListParser{},
		},
	}
}

func (p *genericListParser) Parse(data string, options *Options) (*entities.Entry, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, ftperrors.NewInvalidArgumentError("data", ftperrors.ErrMsgCannotBeBlank)
	}
	if options == nil {
		return nil, ftperrors.NewInvalidArgumentError("options", ftperrors.ErrMsgCannotBeNil)
	}
	for _, parser := range p.parsers {
		entry, err := parser.Parse(data, options)
		if entry != nil && err == nil {
			return entry, nil
		}
	}
	return nil, ftperrors.NewInternalError("unsupported entry format", nil)
}
