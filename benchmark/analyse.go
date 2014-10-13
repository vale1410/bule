
package main

import (
    "bufio"
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

// files need to be organized in this structure: 
// <solver identifier>/<configuration>/<encoding id>/<logfile>.log
// the time is given by a line with 
// xTIME <time in seconds>
// if time is >= timeout then the instance is considered a timeout
// same name of instances in all subfolders/experiments

var ver = flag.Bool("ver", false, "Show version info.")
var dir = flag.String("d", "", "Path of to log files in conform folder structure.")
var timeout = flag.Int("timeout", 3600, "Timeout.")

var digitRegexp = regexp.MustCompile("([0-9]+ )*[0-9]+.[0-9]+")

var timeString = "xTIME"



type statistic struct { 
    solver string
    experiments []experiment
}

type experiment struct { 
    conf string
    encoding string
    times []float64
    solved int
    average float64
    std float64
}


func main() {
    flag.Parse()

    if *ver {
        fmt.Println(`Logfile analyser: Tag 0.2 
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
        return
    }

    analyseLogs(*dir)
}


func analyseLogs(path string) {


    solvers,_ := ioutil.ReadDir(path)
    statistics := make([]statistic,len(solvers))

    for i,solver := range solvers { 



        statistics[i].solver = solver.Name()
        confs,_ := ioutil.ReadDir(path+"/"+solver.Name())

        // determine number of encodings; assumption is that each solver has same
        // number of encodings for each configuration
        if len(confs) > 0 { 
            encodings,_ := ioutil.ReadDir(path+"/"+solver.Name()+"/"+confs[0].Name())

            statistics[i].experiments = make([]experiment,len(confs)*len(encodings))
            for c,conf := range confs {
                for e,enc := range encodings {
                    pos := c*len(encodings)+e
                    statistics[i].experiments[pos].encoding = enc.Name()
                    statistics[i].experiments[pos].conf = conf.Name()
                    statistics[i].experiments[pos].times = 
                        getTimes(path+"/"+solver.Name()+"/"+conf.Name()+"/"+enc.Name())
                }
            }
        }
    }

    fmt.Printf("encoding , solver , seed , #solved , avg. time\n")

    for _,stat := range statistics { 
        //fmt.Println("statistics for "+stat.solver)
        for _,exp := range stat.experiments { 
            analyseTimes(&exp)
            fmt.Printf("%v , %v , %v , %v , %6.2f\n",exp.encoding,stat.solver,exp.conf,exp.solved,exp.average)

        }
    }
}

func analyseTimes(exp* experiment) {

    solved := 0
    total := 0.0

    for i,t := range exp.times {

        if t >= float64(*timeout) { 
            exp.times[i] = -1
        } else { 
            solved++
            total += t
        }

    }

    exp.solved = solved
    exp.average = total/float64(solved)

}

func getTimes(path string) (results []float64) {

    instances,_ := ioutil.ReadDir(path)
    results = make([]float64,len(instances))

    for i,inst := range instances { 

        results[i],_ = parseTime(path+"/"+inst.Name())

    }

    return
}


func parseTime(path string) (time float64,err error) {

    inFile, _ := os.Open(path)

    defer inFile.Close()

    scanner := bufio.NewScanner(inFile)
    scanner.Split(bufio.ScanLines)

    found := false


    for scanner.Scan() {
        s := scanner.Text()
        if strings.HasPrefix(s,timeString) { 

            found = true
            time, err = strconv.ParseFloat(digitRegexp.FindString(s),64)

            if err  != nil { 
            panic(err.Error())
            }

        }
    }

    if !found { 
        fmt.Println(path, "does not contain a line with",timeString)
        panic("Log file does not contain line with time")
    }

    return 

}
