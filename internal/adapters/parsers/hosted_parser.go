package parsers

import "github.com/alexZaicev/go-ftp-client/internal/domain/entities"

type hostedListParser struct {
}

func (p *hostedListParser) Parse(data string) (entry *entities.Entry, err error) {
	return nil, nil
}
