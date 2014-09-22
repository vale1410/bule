package main

import (
    "flag"
    "fmt"
    "github.com/vale1410/bule/sat"
    "github.com/vale1410/bule/sorters"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

var f = flag.String("f", "test.pb", "Instance.")
var size = flag.Int("k", 0, "Size of the vertex cover.")
var out = flag.String("o", "out.cnf", "Path of output file.")
var ver = flag.Bool("ver", false, "Show version info.")
var dbg = flag.Bool("d", false, "Print debug information.")
var dbgfile = flag.String("df", "", "File to print debug information.")

var digitRegexp = regexp.MustCompile("([0-9]+ )*[0-9]+")

var dbgoutput *os.File

type Edge struct {
    a, b int
}

type VertexCover struct {
    NVertex int
    Edges   []Edge
    Cover   []int
    Size    int
}

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
        fmt.Println(`k-Vertex Cover : Tag 0.1
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
        return
    }

    vc := parse(*f)
    vc.Size = *size

    clauses := translateToSAT(vc)

    g := sat.IdGenerator(len(clauses))
    g.GenerateIds(clauses)
    g.Filename = *out
    g.PrintClausesDIMACS(clauses)
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

func translateToSAT(vc VertexCover) (clauses sat.ClauseSet) {

    p := sat.Pred("vc")

    //at least constraint for each edge

    s := "at least one"
    for _, e := range vc.Edges {
        l1 := sat.Literal{true, sat.Atom{p, e.a, 0}}
        l2 := sat.Literal{true, sat.Atom{p, e.b, 0}}
        clauses.AddClause(s, l1, l2)
    }

    //global counter

    sorter := sorters.CreateCardinalityNetwork(vc.NVertex, vc.Size, sorters.AtMost, sorters.Pairwise)
    sorter.RemoveOutput()

    litIn := make([]sat.Literal, vc.NVertex)

    for i, _ := range litIn {
        litIn[i] = sat.Literal{true, sat.Atom{p, i + 1, 0}}
    }

    which := [8]bool{false, false, false, true, true, true, false, false}
    pred := sat.Pred("aux")
    clauses.AddClauseSet(sat.CreateEncoding(litIn, which, []sat.Literal{}, "atMost", pred, sorter))

    return
}

func parse(filename string) (vc VertexCover) {

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

    // 0 : first line, 1 : rest of the lines
    state := 0
    t := 0

    for _, l := range lines {

        if state > 0 && (l == "" || strings.HasPrefix(l, "*")) {
            continue
        }

        elements := strings.Fields(l)

        switch state {
        case 0:
            {
                debug(l)
                var b1 error
                vc.NVertex, b1 = strconv.Atoi(elements[2])
                nEdges, b2 := strconv.Atoi(elements[3])
                if b1 != nil || b2 != nil {
                    debug("cant convert to vertex cover instance:", l)
                    panic("bad conversion of numbers")
                }
                debug("File with verticies ", vc.NVertex, ", and edge, ", nEdges)
                vc.Edges = make([]Edge, nEdges)
                state = 1
            }
        case 1:
            {
                if t >= len(vc.Edges) {
                    panic("number of edges not correctly specified in header.")
                }
                a, b1 := strconv.Atoi(elements[1])
                b, b2 := strconv.Atoi(elements[2])
                if b1 != nil || b2 != nil {
                    debug("cant convert to vertex cover instance:", l)
                    panic("bad conversion of numbers")
                }
                vc.Edges[t] = Edge{a, b}
                t++

            }
        }
    }
    return
}
