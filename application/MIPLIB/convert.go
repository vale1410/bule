package main

// simple rewriter of mps to opb files
// assumes that instances come from the BP category,
// i.e. all variables occurring are 0/1
// and coefficients are integers

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/vale1410/bule/constraints"
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

func getEquationType(s string) constraints.EquationType {
	if s == "E" {
		return constraints.Equal
	} else if s == "L" {
		return constraints.AtMost
	} else if s == "G" {
		return constraints.AtLeast
	}
	if s != "N" {
		panic("UNknown equation type " + s)
	}
	return constraints.Optimization
}

func main() {

	flag.Parse()

	var pbs []constraints.Threshold
	var vars map[string]bool

	if strings.HasSuffix(*filename, "opb") || strings.HasSuffix(*filename, "pbo") || strings.HasSuffix(*filename, "pb") {
		pbs, vars = ParsePBO(*filename)
	} else if strings.HasSuffix(*filename, "mps") {
		pbs, vars = ParseMPS(*filename)
	}

	if *pbo {
		PrintPBO(pbs, vars)
	} else if *gringo {
		PrintGringo(pbs, vars)
	}

	return
}

func PrintGringo(pbs []constraints.Threshold, vars map[string]bool) {
	// print problem to Gringo

	fmt.Println("#hide.")

	for x, _ := range vars {
		fmt.Println("{", x, "}.")
	}
	for _, t := range pbs {
		t.PrintGringo()
	}
}

func PrintPBO(pbs []constraints.Threshold, vars map[string]bool) {
	// print problem to Gringo

	fmt.Println("* #variable=", len(vars), "#constraints=", len(pbs))

	for _, t := range pbs {
		t.NormalizePositiveLiterals()
		t.PrintPBO()
	}

}

func ParseMPS(f string) (pbs []constraints.Threshold, vars map[string]bool) {
	input, err := ioutil.ReadFile(f)

	if err != nil {
		panic("Please specifiy correct path to instance. Does not exist")
	}

	b := bytes.NewBuffer(input)

	lines := strings.Split(strings.Trim(b.String(), " "), "\n")

	state := 0

	pbs = make([]constraints.Threshold, 0, 100)
	vars = make(map[string]bool)
	rowMap := make(map[string]int)

	for _, l := range lines {
		if l == "" {
			continue
		}

		entries := strings.Fields(l)
		ts := constraints.Threshold{}

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
				e := constraints.Entry{sat.NewLit(v), a}
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

func ParsePBO(filename string) (pbs []constraints.Threshold, vars map[string]bool) {

	input, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err.Error())
	}

	lines := strings.Split(string(input), "\n")

	// 0 : first line, 1 : rest of the lines
	var count int
	state := 0
	t := 0

	for _, l := range lines {

		if state > 0 && (l == "" || strings.HasPrefix(l, "%") || strings.HasPrefix(l, "*")) {
			continue
		}

		elements := strings.Fields(l)

		switch state {
		case 0:
			{
				var b1 error
				count, b1 = strconv.Atoi(elements[4])
				vn, b2 := strconv.Atoi(elements[2])
				vars = make(map[string]bool, vn)
				if b1 != nil || b2 != nil {
					panic("bad conversion of numbers")
				}
				pbs = make([]constraints.Threshold, count)
				state = 1
			}
		case 1:
			{
				if t >= count {
					panic("Number of constraints incorrectly specified in pb input file " + filename)
				}
				pbs[t].Desc = l

				n := (len(elements) - 3) / 2
				pbs[t].Entries = make([]constraints.Entry, n)

				for i := 0; i < len(elements)-3; i++ {

					weight, b1 := strconv.ParseInt(elements[i], 10, 64)
					i++
					//variable, b2 := strconv.Atoi(digitRegexp.FindString(elements[i]))

					if b1 != nil {
						panic("bad conversion of numbers")
					}
					atom := sat.NewAtomP(sat.Pred(elements[i]))
					vars[elements[i]] = true
					pbs[t].Entries[i/2] = constraints.Entry{sat.Literal{true, atom}, weight}
				}

				pbs[t].K, _ = strconv.ParseInt(elements[len(elements)-2], 10, 64)
				typS := elements[len(elements)-3]

				if typS == ">=" {
					pbs[t].Typ = constraints.AtLeast
				} else if typS == "<=" {
					pbs[t].Typ = constraints.AtMost
				} else if typS == "==" || typS == "=" {
					pbs[t].Typ = constraints.Equal
				} else {
					panic("bad conversion of symbols" + typS)
				}
				t++
			}
		}
	}
	return
}
