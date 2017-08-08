package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var Timeout_flag = flag.Int("time", 2, "Timeout in seconds")
var Seed_flag = flag.Int("seed", 12, "Random Seed")
var Solver_flag = flag.String("solver", "minisat", "Choose solver. Supported is minisat,microsat,cmsat,lingeling,glucose,clasp")
var Debug_flag = flag.Bool("d", false, "debug.")

func main() {

	flag.Parse()

	finish := make(chan int, 1)

	time_total := time.Now()

	solver := getSolver()
	stdin, err := solver.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdout, err := solver.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if solver.Start() != nil {
		panic(err)
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, os.Stdin)
	}()

	go func() {
		r := bufio.NewReader(stdout)

		for {
			s, err := r.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err.Error())
			}
			if *Debug_flag {
				fmt.Println(s)
			}

			if strings.HasPrefix(s, "c ") {
				continue
			}

			if strings.Contains(s, "UNSATISFIABLE") {
				fmt.Println("UNSATISFIABLE")
				finish <- 0
			} else if strings.Contains(s, "SATISFIABLE") {
				fmt.Println("SATISFIABLE")
				finish <- 1
			}
		}
	}()

	select {
	case i := <-finish:
		fmt.Printf("0,0,0,0,0,0,%v,0,%.3f\n", i, time.Since(time_total).Seconds())
	case <-time.After(time.Duration(*Timeout_flag) * time.Second):
		if err := solver.Process.Kill(); err != nil {
			fmt.Println("failed to kill: ", err)
			panic("")
		}
		fmt.Println("Time limit exceeded!")
	}
	close(finish)

}

func getSolver() (solver *exec.Cmd) {

	//	var Solver_flag *string
	//	if len(os.Args) > 0 {
	//		Solver_flag = &os.Args[1]
	//	}

	seed := strconv.FormatInt(int64(*Seed_flag), 10)

	switch *Solver_flag {
	case "minisat":
		solver = exec.Command("minisat", "-rnd-init", "-rnd-seed="+seed)
	case "glucose":
		solver = exec.Command("glucose", "-rnd-init", "-rnd-seed="+seed)
	case "clasp":
		solver = exec.Command("clasp", "--seed="+seed)
	case "lingeling":
		solver = exec.Command("lingeling", "--seed="+seed)
	case "cmsat":
		solver = exec.Command("cmsat", "--verb=0", "--random="+seed)
		//	case "microsat":
		//		solver = exec.Command("microsat")
		//	case "treengeling":
		//		solver = exec.Command("treengeling")
		//	case "plingeling":
		//		solver = exec.Command("plingeling")
		//	case "dimetheus":
		//		solver = exec.Command("dimetheus", "-seed="+seed)
		//	case "local":
		//		solver = exec.Command("CCAnr", seed)
	default:
		fmt.Println("Solver not available", *Solver_flag)
		os.Exit(1)
	}
	return
}
