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


var digitString = "[0-9]+"
var varString = "[A-Z][0-9a-zA-Z]*"
var nameString = "[a-z][0-9a-zA-Z]*"
var elemString = "(("+digitString+")|("+varString+")|("+nameString+"))"
var atomString = nameString+"\\("+elemString+"(,"+elemString+")*\\)"
var cnfString = atomString + "(," + atomString + ")+$"

var digitRegexp = regexp.MustCompile(digitString)
var nameRegexp = regexp.MustCompile(nameString)
var elemRegexp = regexp.MustCompile(elemString)
var atomRegexp = regexp.MustCompile(atomString)
var cnfRegexp = regexp.MustCompile(cnfString)

type Input struct {
	Name    string
	Content []byte
    Lines   []string
    Atoms   []string
    Rules   []string
}


type itemType int32

const (
    emUnknown itemType = iota
    itemCNF
    itemAtom
    itemImplication
)

func classify(s string) itemType {
    if cnfRegexp.MatchString(s) {
        return itemCNF
    } else if atomRegexp.MatchString(s) {
        return itemAtom
    }
    return itemUnknown
}


func (abox *Abox) String() string {
    s := ""
    for k,v := range *abox {
        s += fmt.Sprintf("%s: ", k.String())
        for _,vv := range v {
            s += "("
            first := true
            for _,j := range vv {
                if first {
                    first = false
                } else {
                    s += ","
                }
                s += strconv.FormatInt(int64(j),10)
            }
            s += ") "
        }
        s += "\n"
    }
    return s
}

func (ident *Identifier) String() string {
    return ident.Name + "/" + strconv.FormatInt(int64(ident.Arity),10)
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

func ParseLines(input *Input) (problem Problem) {

    problem.abox = make(Abox)
    problem.tbox = make(Tbox,0)

    for _,s := range input.Lines {
        switch classify(s) {
        case itemAtom: {
            var ident Identifier
            var elements []int32
            var st string
            st,ident.Name = extract(s,nameRegexp)
            in := digitRegexp.FindAllString(st,-1)
            ident.Arity = int8(len(in))
            elements = make([]int32,len(in),len(in))

            if (len(in) == 0) {
                fmt.Println("no ground elements in atom: ",s)
                continue
            }
            for i,ss := range in {
                x,err := strconv.ParseInt(ss,10,32)
                if err != nil {
                    fmt.Println("could not parse int")
                    fmt.Println(err.Error())
                } else {
                    elements[i] = int32(x)
                }
            }
            _,ok := problem.abox[ident]
            if !ok {
                problem.abox[ident] = make([][]int32,0)
            }
            problem.abox[ident] = append(problem.abox[ident],elements)
        }
        case itemCNF: {
            fmt.Println("we got a rule: ",s)
            elements := strings.Split(s,"),")
            fmt.Println(elements)
            clause := make(Clause,len(elements))
            for i,ss := range elements {
                st, name := extract(ss,nameRegexp)
                fmt.Println(st)
                in := elemRegexp.FindAllString(st,-1)
                ident := Identifier{name,int8(len(in))}
                lit := Literal{ident,in,true}
                fmt.Println(lit)
                clause[i] = lit
            }
            problem.tbox = append(problem.tbox,clause)
        }
        default: {
            fmt.Println("this is shit: ",s)
        }
        }
    }
    return 
}


type Abox map[Identifier]([][]int32)
type Tbox   []Clause

type Identifier struct {
    Name    string
    Arity   int8
}

type Literal struct {
    Ident   Identifier
    Fields  []string
    Sign    bool
}

type Clause []Literal

type Problem struct {
    abox    Abox
    tbox    Tbox
}

func parseGroundLiteral(s string) (ident Identifier, elements []int32) {
    s,ident.Name = extract(s,nameRegexp)
    in := digitRegexp.FindAllString(s,-1)
    ident.Arity = int8(len(in))
    elements = make([]int32,len(in),len(in))
    for i,ss := range in {
        x,err := strconv.ParseInt(ss,10,32)
        if err != nil {
            fmt.Println("could not parse int")
            fmt.Println(err.Error())
        } else {
            elements[i] = int32(x)
        }
    }
    return
}

type Element interface {
    isGround() bool
    getVariable() string
    getValue() string
}

func extract(s string, reg *regexp.Regexp) (sout string, regout string) {
    loc := reg.FindStringIndex(s)
    if loc == nil {
        fmt.Println("extract: string not found: ", s)
        return s,""
    }
    regout = string(s[loc[0]:loc[1]])
    sout = s[:loc[0]] + s[loc[1]+1:]
    return
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
    input.Lines = input.Lines[:len(input.Lines)-1]
    return input,nil
}
