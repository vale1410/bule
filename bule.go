package main

import (
	"fmt"
	"io/ioutil"
	"bytes"
)

func main() {
	fmt.Printf("This is bule (pronounced Boul-ee), the state of the art CNF grounder!\n")
	fileinput, err := ioutil.ReadFile("test.lp")
	input := &Input{Name: "test", Content: fileinput}


	if err != nil {
		fmt.Printf("problem...:\n ")
	}
	fmt.Printf(input.Name + "\n")
    b :=  bytes.NewBuffer(input.Content)
    fmt.Printf(b.String())

}

type Input struct {
	Name    string
	Content []byte
}

func (p *Input) save() error {
	filename := p.Name + ".cnf"
	return ioutil.WriteFile(filename, p.Content, 0600)
}
