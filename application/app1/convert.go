package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"strconv"

	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
)

type entry struct {
	id1, id2 int
	and      int
	c        int64
}

func main() {
	glob.Init()
	input, err2 := os.Open(*glob.Filename_flag)
	defer input.Close()
	if err2 != nil {
		panic("Could not read file")
		return
	}
	scanner := bufio.NewScanner(input)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	state := 0 // 0: read size, 1: read graph 1, 2: read graph 2
	vars := 0
	orig_vars := 0
	size := 0
	i := 0
	var entries []entry

	for scanner.Scan() {
		l := strings.Trim(scanner.Text(), " ")
		if l == "" || strings.HasPrefix(l, "%") || strings.HasPrefix(l, "*") {
			continue
		}
		elements := strings.Fields(l)
		var b error
		switch state {
		case 0: // deprecated: for parsing the "header" of pb files, now parser is flexible
			{
				vars, b = strconv.Atoi(elements[0])
				if b != nil {
					panic("bad conversion of numbers")
				}
				orig_vars = vars
				size, b = strconv.Atoi(elements[1])
				if b != nil {
					panic("bad conversion of numbers")
				}
				entries = make([]entry, size)
				state = 1
			}
		case 1:
			{
				entries[i].id1, b = strconv.Atoi(elements[0])
				if b != nil {
					panic("bad conversion of numbers")
				}
				entries[i].id2, b = strconv.Atoi(elements[1])
				if b != nil {
					panic("bad conversion of numbers")
				}
				var f float64
				f, b = strconv.ParseFloat(elements[2], 64)
				if b != nil {
					panic("bad conversion of numbers")
				}
				entries[i].c = int64(f)
				if entries[i].id1 != entries[i].id2 {
					vars++
					entries[i].and = vars
				}
				i++
			}
		}
	}

	var clauses sat.ClauseSet
	var opt constraints.Threshold
	opt.Typ = constraints.OPT

	lits := make([]sat.Literal, vars+1)

	primaryVars := make(map[string]bool, 0)
	for i := 0; i <= vars; i++ {
		primaryVars[sat.NewAtomP1(sat.Pred("x"), i).Id()] = true
	}
	for i, _ := range lits {
		lits[i] = sat.Literal{true, sat.NewAtomP1(sat.Pred("x"), i)}
	}

	for _, e := range entries {
		if e.id1 == e.id2 {
			opt.Entries = append(opt.Entries, constraints.Entry{lits[e.id1], int64(e.c)})
		} else {
			clauses.AddClause(sat.Neg(lits[e.id1]), sat.Neg(lits[e.id2]), lits[e.and])
			clauses.AddClause(lits[e.id1], sat.Neg(lits[e.and]))
			clauses.AddClause(lits[e.id2], sat.Neg(lits[e.and]))
			opt.Entries = append(opt.Entries, constraints.Entry{lits[e.and], int64(e.c)})
		}
	}

	if *glob.Gringo_flag {
		for i := 0; i <= orig_vars; i++ {
			fmt.Println("{x(", i, ")}.")
		}
		for _, e := range entries {
			if e.id1 != e.id2 {
				fmt.Println(lits[e.and].ToTxt(), ":-", lits[e.id1].ToTxt(), ",", lits[e.id2].ToTxt(), ".")

			}
		}
		opt.PrintGringo()
		return
	}

	g := sat.IdGenerator(clauses.Size()*7 + 1)
	g.PrimaryVars = primaryVars
	opt.NormalizePositiveCoefficients()
	opt.Offset = opt.K

	//	opt.PrintGringo()
	//	clauses.PrintDebug()
	glob.D("offset", opt.Offset)

	glob.A(opt.Positive(), "opt only has positive coefficients")
	g.Solve(clauses, &opt, *glob.Opt_bound_flag, -opt.Offset)

}
