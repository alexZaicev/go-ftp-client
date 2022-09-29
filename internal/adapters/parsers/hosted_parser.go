package parsers

import "github.com/alexZaicev/go-ftp-client/internal/domain/entities"

type hostedListParser struct {
}

func (p *hostedListParser) Parse(data string, options *Options) (entry *entities.Entry, err error) {
	return nil, nil
}
