package main

import (
    "flag"
    "fmt"
//    "github.com/vale1410/bule/constraints"
//    "github.com/vale1410/bule/sat"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

var file = flag.String("f", "instances/data/small", "Instance.")

var asp = flag.Bool("asp",false, "Output facts in ASP format to use with gringo/clasp.")
var dzn = flag.Bool("dzn",false, "Output instance in dzn format to use with minizinc.")
var out = flag.String("o", "out.cnf", "output of conversion.")


var ver = flag.Bool("ver", false, "Show version info.")
var dbg = flag.Bool("d", false, "Print debug information.")
var dbgfile = flag.String("df", "", "File to print debug information.")

//var encoding = flag.Int("e", -1, "EncodingType. 0: naive, 1: sort, 2: split, 3: count, 4: log, 5: Feydy, 6: Hull")
var digitRegexp = regexp.MustCompile("[0-9]+")

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

    p := parse(*file)

    if *asp {
        printASP(p)
    } else if *dzn {
        printDZN(p)
    } else {
        fmt.Println(p)
    }
}

type  problem struct {
    name string
    numActors int
    numScenes int
    actorInScene [][]bool
    actorPay []int
    sceneDuration []int
}

func printDZN(p problem) {

    fmt.Println("numActors =",p.numActors,";")
    fmt.Println("numScenes =",p.numScenes,";")


    fmt.Print("actorPay = [")
    for i,pay := range p.actorPay { 
        fmt.Print(pay)
        if i < len(p.actorPay)-1 { 
            fmt.Print(",")
        }
    }
    fmt.Println("];")

    fmt.Print("sceneDuration = [")
    for i,scene := range p.sceneDuration { 
        fmt.Print(scene)
        if i < len(p.sceneDuration)-1 { 
            fmt.Print(",")
        }
    }
    fmt.Println("];")

    fmt.Print("actorInScene = array2d(Actors,Scenes,[")
    for actor,x := range p.actorInScene { 
        for scene,b := range x { 
            if b { 
                fmt.Print(1)
            } else { 
                fmt.Print(0)
            }
            if actor < p.numActors-1 || scene < p.numScenes-1 { 
            fmt.Print(",")
            }
        }
    }
    fmt.Println("]);")

}
//assumes len()arg
func printArray(name string, arg []int) {
}

func printASP(p problem) {

    printFact("numActors",p.numActors)
    printFact("numScenes",p.numScenes)

    for actor,pay := range p.actorPay { 
        printFact("actorPay",actor,pay)
    }

    for scene,duration := range p.sceneDuration { 
        printFact("sceneDuration",scene,duration)
    }

    for actor,x := range p.actorInScene { 
        for scene,b := range x { 
            if b { 
                printFact("actorInScene",actor,scene)
            }
        }
    }
}

//assumes len()arg
func printFact(arg ...interface{}) {

    if len(arg) < 2 { 
        panic("not enough arguments to generate a fact for asp output.")
    }

    fmt.Print(arg[0])
    fmt.Print("(")
    fmt.Print(arg[1])

    for i:= 2; i < len(arg);i++ {
        fmt.Print(",")
        fmt.Print(arg[i])
    }

    fmt.Println(").")
}

func parse(filename string) (p problem) {

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
    // 3 : ia; actorInScene
    // 4 : c; actorPay
    // 5 : d; sceneDuration


    state := 0
    actor := 0
    var b error

    for _, l := range lines {

        if state > 0 && (l == "" || strings.HasPrefix(l, "%")) {
            continue
        }

        elements := digitRegexp.FindAllString(l, -1)

        debug("parsing line",l)

        switch state {
        case 0: //name
            {
                p.name = l
                state++
                debug("parsing", p.name)
            }
        case 1: //numActors
            {

                p.numScenes, b = strconv.Atoi(elements[0])

                if b != nil && len(elements) != 1 {
                    debug("not expected line in instance file:", l)
                    panic("bad conversion of numbers")
                }

                debug("numScenes", p.numScenes)
                state++
            }
        case 2: //allocate stuff
            {
                p.numActors, b = strconv.Atoi(elements[0])

                if b != nil && len(elements) != 1 {
                    debug("not expected line in instance file:", l)
                    panic("bad conversion of numbers")
                }

                debug("numActors", p.numActors)
                p.actorPay = make([]int,p.numActors)
                p.sceneDuration = make([]int,p.numScenes)

                p.actorInScene = make([][]bool, p.numActors)
                for i,_ := range p.actorInScene {
                    p.actorInScene[i] = make([]bool,p.numScenes)
                }

                state++
            }
        case 3: //fill actorInScene and actorPay
            {

                if len(elements) == 0 {
                    continue
                }
                if actor == p.numActors-1 {
                    state++
                }
                if len(elements) != p.numScenes+1 {
                    debug(p.numScenes+1,":",len(elements))
                    panic("incorrect number of elements.")
                }

                for scene, tr := range elements {

                    a, b := strconv.Atoi(tr)

                    if b != nil {
                        debug("cant convert line:", l)
                        panic("bad conversion of numbers")
                    }
                    if scene == p.numScenes { 
                        p.actorPay[actor] = a
                    } else {

                        if a == 1 {
                            p.actorInScene[actor][scene] = true
                        }
                    }
                }

                actor++

            }

            case 4: //fill sceneDuration
            {
                if len(elements) == 0 {
                    continue
                }

                if len(elements) != p.numScenes {
                    debug(l)
                    panic("incorrect number of elements.")
                }
                for scene, v := range elements {
                    a, b := strconv.Atoi(v)
                    if b != nil {
                        debug("cant convert line:", l)
                        panic("bad conversion of numbers")
                    }
                    p.sceneDuration[scene] = a
                }

            }
        }
    }
    return
}
