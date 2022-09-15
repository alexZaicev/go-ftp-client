package parsers

import "github.com/alexZaicev/go-ftp-client/internal/domain/entities"

type rfc3659ListParser struct {
}

func (p *rfc3659ListParser) Parse(data string) (entry *entities.Entry, err error) {
	return nil, nil
}
