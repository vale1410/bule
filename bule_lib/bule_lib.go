package bule_lib

import (
    "fmt"
	"io/ioutil"
	"bytes"
    "strings"
    "unicode"
)

type itemType int32

const (
    itemError itemType = iota
    itemIdentifier
    itemElement
    itemVariable
    itemDisjunct
    itemImplication
    itemParanthesesOpen
    itemParanthesesClose
)

type item struct {
    typ itemType  // Type, such as itemNumber.
    val string    // Value, such as "23.2".
}

func (i item) String() string {
    if len(i.val) > 10 {
        return fmt.Sprintf("%.10q...", i.val)
    }
    return fmt.Sprintf("%q", i.val)
}


type element int32
type name string

type Input struct {
	Name    string
	Content []byte
    Lines []string
}

type Atom struct {
    n name
    e element
    arity int8
    id int32
    S string
}

type Template struct {
    variable rune
    name string
    s string
}

func (p *Input) save() error {
	filename := p.Name + ".cnf"
	return ioutil.WriteFile(filename, p.Content, 0600)
}

func NewInput(name string) (Input, error) {

	fileinput, err := ioutil.ReadFile(name)

	input := Input{Name: name, Content: fileinput}

	if err != nil {
		fmt.Printf("problem in newInput:\n ") 
        return input,err
	}

    b :=  bytes.NewBuffer(input.Content)

    var buffer bytes.Buffer
    for _,char := range b.String() {
        if !unicode.IsSpace(char) {
            buffer.WriteString(string(char))
        }
    }
    input.Lines = strings.Split(buffer.String(),".")
    return input,nil
}
