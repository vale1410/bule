package main

// simple rewriter of mps to opb files
// assumes that instances come from the BP category,
// i.e. all variables occurring are 0/1
// and coefficients are integers

import (
	"bytes"
	"flag"
	"fmt"
	t "github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"io/ioutil"
	"strconv"
	"strings"
)

var filename = flag.String("f", "test.txt", "Path of the file.")
var gringo = flag.Bool("gringo", true, "Ouput in Potasscos Gringo Format.")
var pbo = flag.Bool("pbo", false, "Ouput in PseudoBoolean Competition Format.")

type Constraint struct {
	name string
}

func getEquationType(s string) t.EquationType {
	if s == "E" {
		return t.Equal
	} else if s == "L" {
		return t.AtMost
	} else if s == "G" {
		return t.AtLeast
	}
	if s != "N" {
		panic("UNknown equation type " + s)
	}
	return t.Optimization
}

func main() {

	flag.Parse()

	pbs, vars := ParseMPS(*filename)

	if *pbo {
		PrintPBO(pbs, vars)
	} else if *gringo {
		PrintGringo(pbs, vars)
	}

	return
}

func PrintGringo(pbs []t.Threshold, vars map[string]bool) {
	// print problem to Gringo

	fmt.Println("#hide.")

	for x, _ := range vars {
		fmt.Println("{", x, "}.")
	}
	for _, t := range pbs {
		t.PrintGringo()
	}
}

func PrintPBO(pbs []t.Threshold, vars map[string]bool) {
	// print problem to Gringo

	fmt.Println("* #variable=", len(vars), "#constraints=", len(vars))

	for _, t := range pbs {
		t.NormalizeAtLeast(true)
		t.PrintPBO()
	}

}

func ParseMPS(f string) (pbs []t.Threshold, vars map[string]bool) {
	input, err := ioutil.ReadFile(f)

	if err != nil {
		panic("Please specifiy correct path to instance. Does not exist")
	}

	b := bytes.NewBuffer(input)

	lines := strings.Split(strings.Trim(b.String(), " "), "\n")

	state := 0

	pbs = make([]t.Threshold, 0, 100)
	vars = make(map[string]bool)
	rowMap := make(map[string]int)

	for _, l := range lines {
		if l == "" {
			continue
		}

		entries := strings.Fields(l)
		ts := t.Threshold{}

		switch state {
		case 0:
			{
				if entries[0] == "NAME" {
					state = 1
				}
				//fmt.Println("name : ", entries[1])
			}
		case 1:
			{
				if entries[0] == "ROWS" {
					state = 2
				}
			}
		case 2: // rows
			if entries[0] == "COLUMNS" {
				state = 3
			} else {
				ts.Typ = getEquationType(entries[0])
				ts.Desc = entries[1]
				pbs = append(pbs, ts)
				rowMap[ts.Desc] = len(pbs) - 1
			}
		case 3: // COLUMNS
			if entries[0] == "INT1" || entries[0] == "INT1END" || entries[0] == "INTEND" ||
				entries[0] == "INTSTART" || strings.HasPrefix(entries[0], "MARK") {
				state = 3
			} else if entries[0] == "RHS" {
				state = 4
			} else {
				a, err := strconv.ParseInt(entries[2], 10, 64)
				if err != nil {
					panic("wrong number " + entries[2])
				}
				v := strings.ToLower(strings.Replace(entries[0], "#", "_", -1))
				if !vars[v] {
					vars[v] = true
				}
				row := strings.ToLower(strings.Replace(entries[1], "#", "_", -1))
				//fmt.Println(v, row)
				e := t.Entry{sat.NewLit(v), a}
				pbs[rowMap[row]].Entries =
					append(pbs[rowMap[row]].Entries, e)
			}
		case 4:
			{
				if entries[0] == "BOUNDS" {
					state = 5
				} else {
					a, err := strconv.ParseInt(entries[2], 10, 64)
					if err != nil {
						panic("wrong number " + entries[2])
					}
					row := strings.ToLower(strings.Replace(entries[1], "#", "_", -1))
					pbs[rowMap[row]].K = a
				}
			}
		case 5:
			{
			}
		}
	}
	return
}
