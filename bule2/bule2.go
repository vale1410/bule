package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

/*
TODO:
-
*/

func main() {

	if len(os.Args) < 2 {
		fmt.Println("usage: ./ground <filename> | -dimacs | []<unit>")
		return
	}

	printInfoFlag := true
	dimacs := false

	units := make(map[string]bool, 0)

	for i, s := range os.Args {
		if i < 2 {
			continue
		}
		if s == "-dimacs" {
			dimacs = true
			continue
		}
		if strings.HasPrefix(s, "-") {
			s = "~" + strings.TrimLeft(s, "-")
		}
		units[s] = true
	}

	count := 1

	cls := [][]string{}

	// open a file or stream
	var scanner *bufio.Scanner
	file, err := os.Open(os.Args[1])
	if err != nil {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {

		fields := strings.Fields(scanner.Text())

		if len(fields) == 0 || strings.HasPrefix(fields[0], "%") {
			continue
		}

		first := fields[0]

		if first == "c" {
			continue
		}

		clause := fields

		if len(clause) == 1 {
			units[clause[0]] = true
		} else {
			cls = append(cls, clause)
		}
	}

	size := 0

	var cls2 [][]string
	conflict := false

	for size < len(units) {
		//fmt.Println("units", units)
		size = len(units)
		cls2 = make([][]string, 0, len(cls))

		for _, clause := range cls {
			clause2 := make([]string, 0, len(clause))
			keepClause := true

			//fmt.Println("clause", clause)
			for _, lit := range clause {
				if _, b := units[lit]; b {
					keepClause = false
				}
				//fmt.Println(units, lit, neg(lit))
				if _, b := units[neg(lit)]; !b {
					clause2 = append(clause2, lit)
				} else {
					//fmt.Println("remove", lit, "from", clause)
				}
			}
			//fmt.Println("clause2", clause2)
			if len(clause2) == 1 {
				units[clause2[0]] = true
			} else if len(clause2) == 0 {
				fmt.Println("c conflict:", clause)
				conflict = true
			}

			if keepClause && len(clause2) > 1 {
				cls2 = append(cls2, clause2)

			}
		}
		cls = cls2
	}

	vars := make(map[string]int, 0)
	{ // generate id's for variables
		for lit, _ := range units {
			v := pos(lit)
			if _, b := vars[v]; !b {
				vars[v] = count
				count++
			}
		}

		for _, clause := range cls {
			for _, lit := range clause {
				v := pos(lit)
				if _, b := vars[v]; !b {
					vars[v] = count
					count++
				}
			}
		}
	}

	if dimacs {

		if printInfoFlag {
			varids := make([]string, len(vars)+1)
			for v, i := range vars {
				varids[i] = v
			}
			for i, v := range varids {
				if i > 0 {
					fmt.Println("c", i, v)
				}
			}
		}

		if conflict {
			fmt.Println("p cnf 1 2 \n 1 0\n -1 0\n")
			return
		}

		if printInfoFlag {
			fmt.Println("p", "cnf", len(vars), len(cls)+len(units))
		} else {
			fmt.Println("p", "cnf", len(vars)-len(units), len(cls))
		}

		if printInfoFlag {
			for lit, _ := range units {
				if strings.HasPrefix(lit, "~") {
					fmt.Print("-")
				}
				fmt.Print(vars[pos(lit)], " ")
				fmt.Println(0)
			}
		}

		for _, clause := range cls {

			for _, lit := range clause {
				if strings.HasPrefix(lit, "~") {
					fmt.Print("-")
				}
				fmt.Print(vars[pos(lit)], " ")
			}
			fmt.Println("0")
		}

	} else {
		//     fmt.Println("c units")
		for unit, _ := range units {
			fmt.Println(unit)

		}
		//		fmt.Println("c clauses")
		for _, clause := range cls {

			for _, v := range clause {
				fmt.Print(v, " ")
			}
			fmt.Println()
		}

	}

}

func pos(s string) string {
	if strings.HasPrefix(s, "~") {
		return strings.TrimLeft(s, "~")
	} else {
		return s
	}
}
func neg(s string) string {
	if strings.HasPrefix(s, "~") {
		return strings.TrimLeft(s, "~")
	} else {
		return "~" + s
	}
}
