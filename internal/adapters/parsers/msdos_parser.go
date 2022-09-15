package parsers

import "github.com/alexZaicev/go-ftp-client/internal/domain/entities"

type msDosListParser struct {
}

func (p *msDosListParser) Parse(data string) (entry *entities.Entry, err error) {
	return nil, nil
}
