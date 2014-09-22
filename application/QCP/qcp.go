package main

import (
    "flag"
    "fmt"
    "github.com/vale1410/bule/sat"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

var f = flag.String("f", "qwh-5-10.pls", "Instance.")
var out = flag.String("o", "out.cnf", "Path of output file.")
var ver = flag.Bool("ver", false, "Show version info.")
var dbg = flag.Bool("d", false, "Print debug information.")
var dbgfile = flag.String("df", "", "File to print debug information.")

var digitRegexp = regexp.MustCompile("([0-9]+ )*[0-9]+")

var dbgoutput *os.File

type QGC [][]int

func main() {
    flag.Parse()

    if *dbgfile != "" {
        var err error
        dbgoutput, err = os.Create(*dbgfile)
        if err != nil {
            panic(err)
        }
        defer dbgoutput.Close()
    }

    debug("Running Debug Mode...")

    if *ver {
        fmt.Println(`QGC-Translator: Tag 0.1
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
        return
    }

    g := parse(*f)

    clauses := translateToSAT(g)

    s := sat.IdGenerator(len(clauses))
    s.GenerateIds(clauses)
    s.Filename = *out
    s.PrintClausesDIMACS(clauses)

    if *dbg {
        s.PrintDebug(clauses)
    }
}

func debug(arg ...interface{}) {
    if *dbg {
        if *dbgfile == "" {
            fmt.Print("dbg: ")
            for _, s := range arg {
                fmt.Print(s, " ")
            }
            fmt.Println()
        } else {
            ss := "dbg: "
            for _, s := range arg {
                ss += fmt.Sprintf("%v", s) + " "
            }
            ss += "\n"

            if _, err := dbgoutput.Write([]byte(ss)); err != nil {
                panic(err)
            }
        }
    }
}

func translateToInstance(g QGC) (clauses sat.ClauseSet) {

    p := sat.Pred("v")

    for i, r := range g {
        for j, k := range r {
            fmt.Print(" ", k)
            if k >= 0 {
                l1 := sat.Literal{true, sat.NewAtomP3(p, i, j, k)}
                clauses.AddTaggedClause("Instance", l1)
            }
        }
        fmt.Println("")
    }

    return
}

// in every row/column each value has to occur at least once
func translateRedundant(n int) (clauses sat.ClauseSet) {
    p := sat.Pred("v")

    for i := 0; i < n; i++ {
        for k := 0; k < n; k++ {
            lits := make([]sat.Literal, n)
            for j := 0; j < n; j++ {
                lits[k] = sat.Literal{true, sat.NewAtomP3(p, i, j, k)}
            }
            clauses.AddTaggedClause("AtLeastValueV", lits...)
        }
    }

    for j := 0; j < n; j++ {
        for k := 0; k < n; k++ {
            lits := make([]sat.Literal, n)
            for i := 0; i < n; i++ {
                lits[k] = sat.Literal{true, sat.NewAtomP3(p, i, j, k)}
            }
            clauses.AddTaggedClause("AtLeastValueH", lits...)
        }
    }
}

func translateAMO(n int,typ AtMostType) (clauses sat.ClauseSet) {

    p := sat.Pred("v")

    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            lits := make([]sat.Literal, n)
            for k := 0; k < n; k++ {
                lits[k] = sat.Literal{true, sat.NewAtomP3(p, i, j, k)}
            }
            clauses.AddTaggedClause("AtLeast", lits...)
        }
    }

    for k := 0; k < n; k++ {
        for i := 0; i < n; i++ {
            lits := make([]sat.Literal, n)
            for j := 0; j < n; j++ {
                lits[k] = sat.Literal{true, sat.NewAtomP3(p, i, j, k)}
            }
            // in each row each value at most one
            clauses.AddClauseSet(constraints.atMostOne(typ, "amo", lits))
        }
    }

    for k := 0; k < n; k++ {
        for j := 0; j < n; j++ {
            lits := make([]sat.Literal, n)
            for i := 0; i < n; i++ {
                lits[k] = sat.Literal{true, sat.NewAtomP3(p, i, j, k)}
            }
            // in each column each value at most one
            clauses.AddClauseSet(constraints.atMostOne(typ, "amo", lits))
        }
    }

    return
}

func translateNaive(n int) (clauses sat.ClauseSet) {

    p := sat.Pred("v")

    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            lits := make([]sat.Literal, n)
            for k := 0; k < n; k++ {
                lits[k] = sat.Literal{true, sat.NewAtomP3(p, i, j, k)}
            }
            clauses.AddTaggedClause("AtLeast", lits...)
        }
    }

    for k := 0; k < n; k++ {
        for i := 0; i < n; i++ {
            for j1 := 0; j1 < n; j1++ {
                for j2 := j1 + 1; j2 < n; j2++ {
                    l1 := sat.Literal{false, sat.NewAtomP3(p, i, j1, k)}
                    l2 := sat.Literal{false, sat.NewAtomP3(p, i, j2, k)}
                    clauses.AddTaggedClause("Horizontal", l1, l2)
                }
            }
        }
    }

    for k := 0; k < n; k++ {
        for i := 0; i < n; i++ {
            for j1 := 0; j1 < n; j1++ {
                for j2 := j1 + 1; j2 < n; j2++ {
                    l1 := sat.Literal{false, sat.NewAtomP3(p, j1, i, k)}
                    l2 := sat.Literal{false, sat.NewAtomP3(p, j2, i, k)}
                    clauses.AddTaggedClause("Vertical", l1, l2)
                }
            }
        }
    }

    return
}

func translateEncodingNaive(n int) (clauses sat.ClauseSet) {

    p := sat.Pred("v")

    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            lits := make([]sat.Literal, n)
            for k := 0; k < n; k++ {
                lits[k] = sat.Literal{true, sat.NewAtomP3(p, i, j, k)}
            }
            clauses.AddTaggedClause("AtLeast", lits...)
        }
    }

    for k := 0; k < n; k++ {
        for i := 0; i < n; i++ {
            for j1 := 0; j1 < n; j1++ {
                for j2 := j1 + 1; j2 < n; j2++ {
                    l1 := sat.Literal{false, sat.NewAtomP3(p, i, j1, k)}
                    l2 := sat.Literal{false, sat.NewAtomP3(p, i, j2, k)}
                    clauses.AddTaggedClause("Horizontal", l1, l2)
                }
            }
        }
    }

    for k := 0; k < n; k++ {
        for i := 0; i < n; i++ {
            for j1 := 0; j1 < n; j1++ {
                for j2 := j1 + 1; j2 < n; j2++ {
                    l1 := sat.Literal{false, sat.NewAtomP3(p, j1, i, k)}
                    l2 := sat.Literal{false, sat.NewAtomP3(p, j2, i, k)}
                    clauses.AddTaggedClause("Vertical", l1, l2)
                }
            }
        }
    }

    return
}

func parse(filename string) (g QGC) {

    input, err := ioutil.ReadFile(filename)

    if err != nil {
        fmt.Println("Please specifiy correct path to instance. File does not exist: ", filename)
        panic(err)
    }

    output, err := os.Create(*out)
    if err != nil {
        panic(err)
    }
    defer output.Close()

    lines := strings.Split(string(input), "\n")

    if matched, _ := regexp.MatchString(".dzn", filename); matched {
        g = parseDZN(lines)
    } else if matched, _ := regexp.MatchString(".pls", filename); matched {
        g = parsePLS(lines)
    } else {
        debug("filename/type unknown:", filename)
        panic("")
    }

    return
}

func parseDZN(lines []string) (g QGC) {
    // 0 : first line, 1 : rest of the lines
    state := 0
    t := 0

    for ln, l := range lines {

        if state > 0 && (l == "" || strings.HasPrefix(l, "%")) {
            continue
        }

        elements := digitRegexp.FindAllString(l, -1)

        switch state {
        case 0:
            {
                debug(l)

                n, b := strconv.Atoi(elements[0])

                if b != nil && len(elements) != 1 {
                    debug("first line in data file wrong:", l)
                    panic("bad conversion of numbers")
                }

                debug("Size of problem", n)

                g = make(QGC, n)
                for i, _ := range g {
                    g[i] = make([]int, n)
                }
                state = 1
            }
        case 1:
            {
                // skip this one :-)
                state++
            }
        case 2:
            {
                fmt.Println(elements)

                if len(elements) == 0 {
                    continue
                }
                if t > len(g) {
                    debug(t, " ", l)
                    panic("incorrect number of elements.")
                }

                for i, p := range elements {

                    a, b := strconv.Atoi(p)

                    if b != nil {
                        debug("cant convert to instance:", l)
                        panic("bad conversion of numbers")
                    }

                    // -1 means unknown
                    // domain is 0 .. n-1
                    g[ln-2][i] = a - 1
                }

            }
        }
    }
    fmt.Println(g)
    return
}

func parsePLS(lines []string) (g QGC) {

    // 0 : first line, 1 : rest of the lines
    state := 0
    t := 0

    for ln, l := range lines {

        if state > 0 && (l == "" || strings.HasPrefix(l, "*")) {
            continue
        }

        elements := strings.Fields(l)

        switch state {
        case 0:
            {
                debug(l)

                n, b := strconv.Atoi(elements[1])

                if b != nil || elements[0] != "order" {
                    debug("no proper stuff to read:", l)
                    panic("bad conversion of numbers")
                }

                debug("Size of problem", n)

                g = make(QGC, n)
                for i, _ := range g {
                    g[i] = make([]int, n)
                }
                state = 1
            }
        case 1:
            {
                if t > len(g) {
                    debug(t, " ", l)
                    panic("incorrect number of elements.")
                }

                for i, p := range elements {

                    a, b := strconv.Atoi(p)

                    if b != nil {
                        debug("cant convert to instance:", l)
                        panic("bad conversion of numbers")
                    }

                    g[ln-1][i] = a
                }

            }
        }
    }
    return
}
