package bule_lib

import (
	"io/ioutil"
)

type Input struct {
	Name    string
	Content []byte
}

func (p *Input) save() error {
	filename := p.Name + ".cnf"
	return ioutil.WriteFile(filename, p.Content, 0600)
}
