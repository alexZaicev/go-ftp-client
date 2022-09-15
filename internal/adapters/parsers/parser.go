package parsers

import (
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	"github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

type Parser interface {
	Parse(data string) (*entities.Entry, error)
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

func (p *genericListParser) Parse(data string) (*entities.Entry, error) {
	for _, parser := range p.parsers {
		entry, err := parser.Parse(data)
		if entry != nil && err == nil {
			return entry, nil
		}
	}
	return nil, errors.NewInternalError("unsupported list format", nil)
}
