package main

import (
    "flag"
    "fmt"
    "github.com/vale1410/bule/sat"
    "github.com/vale1410/bule/threshold"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

var f = flag.String("f", "test.pb", "Path of to PB file.")
var out = flag.String("o", "out.cnf", "Path of output file.")
var ver = flag.Bool("ver", false, "Show version info.")
var check_clause = flag.Bool("clause", true, "Checks if Pseudo-Boolean is not just a simple clause.")
var dbg = flag.Bool("d", false, "Print debug information.")
var dbgfile = flag.String("df", "", "File to print debug information.")
var reformat = flag.Bool("reformat", false, "Reformat PB files into correct format. Decompose = into >= and <=")
var gurobi = flag.Bool("gurobi", false, "Reformat to Gurobi input, output to stdout.")

//var model = flag.String("model", "model.lp", "path to model file")
//var solve = flag.Bool("solve", false, "Pass problem to clasp and solve.")
//var ttimeout = flag.Int("timeout", 10, "Timeout in seconds.")

var digitRegexp = regexp.MustCompile("([0-9]+ )*[0-9]+")

var dbgoutput *os.File

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
        fmt.Println(`Bule CNF Grounder: Tag 0.1 Pseudo Booleans
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
        return
    }

    pbs := parse(*f)

    if *gurobi {
        fmt.Println("Subject To")
        atoms := make(map[string]bool, len(pbs))
        for _, pb := range pbs {
            pb.NormalizeAtLeast(true)
            pb.PrintGurobi()
            for _, x := range pb.Entries {
                atoms[x.Literal.A.Id()] = true
            }
        }
        fmt.Println("Binary")
        for aS, _ := range atoms {
            fmt.Print("x" + aS + " ")
        }
        fmt.Println()
    } else if *reformat {
        for _, pb := range pbs {
            if pb.Typ == threshold.Equal {
                pb.Typ = threshold.AtLeast
                pb.NormalizeAtLeast(true)
                pb.Print10()
                pb.Typ = threshold.AtMost
            }
            pb.NormalizeAtLeast(true)
            pb.Print10()
        }
    } else {

        var clauses sat.ClauseSet

        for i, pb := range pbs {
            TranslatePB2Clauses(i, &pb)
            clauses.AddClauseSet(pb.Clauses)
            debug("number of clause", pb.Clauses.Size())
        }

        g := sat.IdGenerator(clauses.Size() * 7)
        g.GenerateIds(clauses)
        //g.Filename = strings.Split(*f, ".")[0] + ".cnf"
        g.Filename = *out
        g.PrintClausesDIMACS(clauses)
    }
}

func TranslatePB2Clauses(id int, pb *threshold.Threshold) {

    pb.Translate()

    if pb.Trans == threshold.SingleClause || pb.Trans == threshold.Facts {
    } else {

        wh := 2
        var which [8]bool

        switch wh {
        case 1:
            which = [8]bool{false, false, false, true, true, true, false, false}
        case 2:
            which = [8]bool{false, false, false, true, true, true, false, true}
        case 3:
            which = [8]bool{false, true, true, true, true, true, true, false}
        case 4:
            which = [8]bool{false, true, true, true, true, true, true, true}
        }

        pred := sat.Pred("auxPB" + strconv.Itoa(id))
        pb.Clauses = sat.CreateEncoding(pb.LitIn, which, []sat.Literal{}, "BnB", pred, pb.Sorter)
    }
    return
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

func parse(filename string) (pbs []threshold.Threshold) {

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
    var count int
    state := 0
    t := 0

    for _, l := range lines {

        //if strings.HasPrefix(l, "*") {
        //  fmt.Println(l)
        //}

        if state > 0 && (l == "" || strings.HasPrefix(l, "*")) {
            continue
        }

        elements := strings.Fields(l)

        switch state {
        case 0:
            {
                debug(l)
                var b1 error
                count, b1 = strconv.Atoi(elements[4])
                vars, b2 := strconv.Atoi(elements[2])
                if b1 != nil || b2 != nil {
                    debug("cant convert to threshold:", l)
                    panic("bad conversion of numbers")
                }
                debug("Found PB file with", count, "constraints and", vars, "variables")
                pbs = make([]threshold.Threshold, count)
                state = 1
            }
        case 1:
            {
                if t >= count {
                    panic("Number of constraints incorrectly specified in pb input file " + filename)
                }
                pbs[t].Desc = l

                n := (len(elements) - 3) / 2
                pbs[t].Entries = make([]threshold.Entry, n)

                for i := 0; i < len(elements)-3; i++ {

                    weight, b1 := strconv.ParseInt(elements[i], 10, 64)
                    i++
                    variable, b2 := strconv.Atoi(digitRegexp.FindString(elements[i]))

                    if b1 != nil || b2 != nil {
                        debug("cant convert to threshold:", l)
                        panic("bad conversion of numbers")
                    }
                    atom := sat.NewAtomP1(sat.Pred("x"), variable)
                    pbs[t].Entries[i/2] = threshold.Entry{sat.Literal{true, atom}, weight}
                }

                pbs[t].K, _ = strconv.ParseInt(elements[len(elements)-2], 10, 64)
                typS := elements[len(elements)-3]

                if typS == ">=" {
                    pbs[t].Typ = threshold.AtLeast
                } else if typS == "<=" {
                    pbs[t].Typ = threshold.AtMost
                } else if typS == "=" {
                    pbs[t].Typ = threshold.Equal
                } else {
                    debug("cant convert to threshold:", l)
                    panic("bad conversion of symbols" + typS)
                }
                t++
            }
        }
    }
    return
}
