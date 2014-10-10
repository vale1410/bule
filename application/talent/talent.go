package main

import (
    "flag"
    "fmt"
    "github.com/vale1410/bule/constraints"
    "github.com/vale1410/bule/sat"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

var f = flag.String("f", "instances/qwh-5-10.pls", "Instance.")
var out = flag.String("o", "out.cnf", "output of conversion.")
var ver = flag.Bool("ver", false, "Show version info.")
var dbg = flag.Bool("d", false, "Print debug information.")
var dbgfile = flag.String("df", "", "File to print debug information.")
var encoding = flag.Int("e", -1, "EncodingType. 0: naive, 1: sort, 2: split, 3: count, 4: log, 5: Feydy, 6: Hull")


var dbgoutput *os.File

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
        fmt.Println(`Model Generator for Talent Scheduling: Tag 0.1
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
        return
    }

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

    // 0 : name 
    // 1 : numActors
    // 2 : numScenes


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
