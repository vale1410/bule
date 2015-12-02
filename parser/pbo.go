package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
)

type Problem struct {
	Opt *constraints.Threshold
	Pbs []*constraints.Threshold
}

func (p *Problem) PrintPBO() {
	atoms := make(map[string]bool, len(p.Pbs))

	for _, pb := range p.Pbs {
		for _, x := range pb.Entries {
			atoms[x.Literal.A.Id()] = true
		}
	}
	fmt.Printf("* #variable= %v #constraint= %v\n", len(atoms), len(p.Pbs)-1)

	for _, pb := range p.Pbs {
		pb.PrintPBO()
	}
}

func (p *Problem) PrintGringo() {
	fmt.Println("#hide.")
	atoms := make(map[string]bool, len(p.Pbs))

	for _, pb := range p.Pbs {
		pb.PrintGringo()
		for _, x := range pb.Entries {
			atoms[x.Literal.A.Id()] = true
		}
	}
	for x, _ := range atoms {
		fmt.Println("{", x, "}.")
	}
}

func (p *Problem) PrintGurobi() {
	if !p.Opt.Empty() {
		fmt.Println("Minimize")
		p.Opt.PrintGurobi()
	}
	fmt.Println("Subject To")
	atoms := make(map[string]bool, len(p.Pbs))
	for i, pb := range p.Pbs {
		if i > 0 {
			pb.Normalize(constraints.GE, false)
			pb.PrintGurobi()
			for _, x := range pb.Entries {
				atoms[x.Literal.A.Id()] = true
			}
		}
	}
	fmt.Println("Binary")
	for aS, _ := range atoms {
		fmt.Print(aS + " ")
	}
	fmt.Println()
}

func New(filename string) Problem {
	pbs, err := parse(filename)

	if err != nil {
		panic(err)
	}
	opt := pbs[0] // per convention first in pbs is opt statement (possibly empty)
	return Problem{opt, pbs}
}

// returns list of *pb; first one is optimization statement, possibly empty
func parse(filename string) (pbs []*constraints.Threshold, err error) {

	input, err2 := os.Open(filename)
	defer input.Close()
	if err2 != nil {
		err = errors.New("Please specifiy correct path to instance. Does not exist")
		return
	}
	scanner := bufio.NewScanner(input)

	// 0 : first line, 1 : rest of the lines
	var count int
	state := 1
	t := 0
	pbs = make([]*constraints.Threshold, 0)

	for scanner.Scan() {
		l := strings.Trim(scanner.Text(), " ")
		if l == "" || strings.HasPrefix(l, "%") || strings.HasPrefix(l, "*") {
			continue
		}

		elements := strings.Fields(l)

		if len(elements) == 1 { // quick hack to ignore single element lines (not neccessary)
			continue
		}

		switch state {
		case 0: // deprecated: for parsing the "header" of pb files, now parser is flexible
			{
				glob.D(l)
				var b1 error
				count, b1 = strconv.Atoi(elements[4])
				vars, b2 := strconv.Atoi(elements[2])
				if b1 != nil || b2 != nil {
					glob.D("cant convert to threshold:", l)
					panic("bad conversion of numbers")
				}
				glob.D("File PB file with", count, "constraints and", vars, "variables")
				state = 1
			}
		case 1:
			{

				var n int  // number of entries
				var f int  // index of entry
				var o bool //optimization
				var pb constraints.Threshold

				offset_back := 0
				if elements[len(elements)-1] != ";" {
					offset_back = 1
				}

				if elements[0] == "min:" || elements[0] == "Min" {
					o = true
					n = (len(elements) + offset_back - 2) / 2
					f = 1
				} else {
					o = false
					n = (len(elements) + offset_back - 3) / 2
					f = 0
				}

				pb.Entries = make([]constraints.Entry, n)

				for i := f; i < 2*n; i++ {

					weight, b1 := strconv.ParseInt(elements[i], 10, 64)
					i++
					if b1 != nil {
						glob.D("cant convert to threshold:", elements[i], "\nin PB\n", l)
						panic("bad conversion of numbers")
					}
					atom := sat.NewAtomP(sat.Pred(elements[i]))
					pb.Entries[(i-f)/2] = constraints.Entry{sat.Literal{true, atom}, weight}
				}
				// fake empty opt in case it does not exist
				if t == 0 && !o {
					pbs = append(pbs, &constraints.Threshold{})
					t++
				}
				pb.Id = t
				if o {
					pb.Typ = constraints.OPT
					glob.D("Scanned optimization statement")
				} else {
					pb.K, err = strconv.ParseInt(elements[len(elements)-2+offset_back], 10, 64)

					if err != nil {
						glob.A(false, " cant parse threshold, error", err.Error(), pb.K)
					}
					typS := elements[len(elements)-3+offset_back]

					if typS == ">=" {
						pb.Typ = constraints.GE
					} else if typS == "<=" {
						pb.Typ = constraints.LE
					} else if typS == "==" || typS == "=" {
						pb.Typ = constraints.EQ
					} else {
						glob.A(false, "cant convert to threshold, equationtype typS:", typS)
					}
				}

				pbs = append(pbs, &pb)
				t++
				//fmt.Println(pb.Id)
				//pb.Print10()
			}
		}
	}

	glob.A(len(pbs) == t, "Id of constraint must correspond to position")
	glob.D("Scanned", t-1, "PB constraints.")
	if !pbs[0].Empty() {
		glob.D("Scanned OPT statement.")
	}
	return
}
