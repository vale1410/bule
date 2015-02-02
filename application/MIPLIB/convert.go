package main

// simple rewriter of mps to opb files
// assumes that instances come from the BP category,
// i.e. all variables occurring are 0/1
// and coefficients are integers

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type EquationType int

const (
	AtMost EquationType = iota
	AtLeast
	Equal
	Optimization
)

type Entry struct {
	Literal string
	Weight  int64
}

type Threshold struct {
	Desc    string
	Entries []Entry
	K       int64
	Typ     EquationType
}

var filename = flag.String("file", "test.txt", "Path of the file specifying the Knapsack Problem.")

type Constraint struct {
	name string
}

func getEquationType(t string) EquationType {
	if t == "E" {
		return Equal
	} else if t == "L" {
		return AtMost
	} else if t == "G" {
		return AtLeast
	}
	if t != "N" {
		panic("UNknown equation type " + t)
	}
	return Optimization
}

func main() {

	flag.Parse()

	input, err := ioutil.ReadFile(*filename)

	if err != nil {
		panic("Please specifiy correct path to instance. Does not exist")
	}

	b := bytes.NewBuffer(input)

	lines := strings.Split(strings.Trim(b.String(), " "), "\n")

	state := 0

	pbs := make([]Threshold, 0, 100)
	rowMap := make(map[string]int)

	for _, l := range lines {
		if l == "" {
			continue
		}
		entries := strings.Fields(l)
		ts := Threshold{}

		switch state {
		case 0:
			{
				if entries[0] == "NAME" {
					state = 1
				}
				fmt.Println("name : ", entries[1])
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
			if entries[0] == "INT1" || strings.HasPrefix(entries[0], "MARK") {
				state = 3
			} else if entries[0] == "RHS" {
				state = 4
			} else {
				a, err := strconv.ParseInt(entries[2], 10, 64)
				if err != nil {
					panic("wrong number " + entries[2])
				}
				e := Entry{entries[0], a}
				pbs[rowMap[entries[1]]].Entries =
					append(pbs[rowMap[entries[1]]].Entries, e)
			}
		case 4:
			{
				if len(entries) < 3 {
					fmt.Println(len(entries))
				}
				if entries[0] == "BOUNDS" {
					state = 5
				}
				a, err := strconv.ParseInt(entries[3], 10, 64)
				if err != nil {
					panic("wrong number " + entries[2])
				}
				pbs[rowMap[entries[1]]].K = a
			}
		case 5:
			{
			}
		}
	}

	for _, t := range pbs {
		fmt.Println(t)
	}
	return
}
