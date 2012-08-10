package bule_lib

import (
    "fmt"
	"io/ioutil"
	"bytes"
    "strings"
    "unicode"
    "regexp"
    "strconv"
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


type name string

type Input struct {
	Name    string
	Content []byte
    Lines []string
}

type Atom struct {
    n name
    arity int8
    elements []int32
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

func ParseAtoms(input *Input) []Atom {

    atoms := make([]Atom,0,len(input.Lines))

    var digitRegexp = regexp.MustCompile("[0-9]+")
    var nameRegexp = regexp.MustCompile("[a-z][a-zA-Z]*")

    for _,s := range input.Lines {
        var a Atom
        a.n = name(nameRegexp.FindString(s))
        in := digitRegexp.FindAllString(s,-1)
        fmt.Println(a.n)
        fmt.Println(in)
        a.arity = int8(len(in))
        a.elements = make([]int32,len(in),len(in))
        for i,ss := range in {
            x,err := strconv.ParseInt(ss,10,32)
            if err != nil {
                fmt.Println(err.Error())
            } else {
                a.elements[i] = int32(x)
            }
        }
        atoms = append(atoms,a)
    }
    return atoms
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
