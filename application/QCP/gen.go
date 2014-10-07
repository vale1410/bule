
package main

import (
    "flag"
    "fmt"
    "math/rand"

)

var n = flag.Int("n", 4,"size of matrix.")
var m = flag.Int("m", 2,"Only numbers from 1..m on the matrix.")
var p = flag.Int("p", 10,"percentage of numbers spread across the matrix. ")
var out = flag.String("o", "out.pls", "output of instance.")

var dzn = flag.Bool("dzn", false, "output in dzn format.")

var ver = flag.Bool("ver", false, "Show version info.")

var shuffle = flag.Bool("shuffle",false,"Shuffles numbers.")

type QCP [][]int

func main() {
    flag.Parse()


    if *ver {
        fmt.Println(`QGC-Generator: Tag 0.1
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
        return
    }


    generate(*n,*m,*p)

}

func generate(size int, highest int, percentage int) {

    //for each number
    //  guess percentage times
    //      a new number for row/column
    //          block row column
    //          add to QGC

    q := make([][]int,size)

    for i,_ := range q {
        q[i] = make([]int,size)
    }


    maxCount := (percentage*size)/100

    //fmt.Println(size,highest,maxCount)

    for number := 1 ; number <= highest; number++ {

        i := 0
        rows := make([]bool,size)
        cols := make([]bool,size)

        for i < maxCount {

            row := int(rand.Int31n(int32(size)))
            col := int(rand.Int31n(int32(size)))
            //fmt.Println(number,row,col)

            if q[row][col] == 0 &&  !rows[row] && !cols[col]{ 
                i++
                //fmt.Println(".")
                q[row][col] = number
                rows[row] = true
                cols[col] = true
            }
        } 
    } 

    if*dzn { 
        fmt.Println("N=",*n,";")
        fmt.Println("start=[|")

        for i,x := range q { 
            for j,y := range x { 
                if j == len(x)-1 { 
                    fmt.Print(y,"|")
                } else {
                    fmt.Print(y,", ")
                }
            } 
            if i == len(q)-1 {
                fmt.Println("];")
            } else { 
            fmt.Println()
            }
            } 
    } else {  // ouput in dzn format

        fmt.Println("order",*n)
        for _,x := range q { 
            for _,y := range x { 
                fmt.Print(y-1," ")
            } 
            fmt.Println()
        } 

    }

    return 
}
