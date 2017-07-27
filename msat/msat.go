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

var Timeout_flag = flag.Int("t", 2, "Timeout in seconds")
var Seed_flag = flag.Int("s", 12, "Random Seed")

func main() {

	flag.Parse()

	finish := make(chan int, 1)
	//	timeout := make(chan bool, 1)

	time_total := time.Now()

	//	go func() {
	//		time.Sleep(time.Duration(*Timeout_flag) * time.Second)
	//		timeout <- true
	//	}()

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
		//	case <-timeout:
		//
	case <-time.After(time.Duration(*Timeout_flag) * time.Second):
		if err := solver.Process.Kill(); err != nil {
			log.Fatal("failed to kill: ", err)
		}
		fmt.Println("Time limit exceeded!")
	}
	close(finish)
	//	close(timeout)

}

func getSolver() (solver *exec.Cmd) {

	var Solver_flag *string
	if len(os.Args) > 0 {
		Solver_flag = &os.Args[1]
	}

	seed := strconv.FormatInt(int64(*Seed_flag), 10)

	switch *Solver_flag {
	case "minisat":
		//solver = exec.Command("minisat", "-rnd-seed=123")
		solver = exec.Command("minisat", "-rnd-seed="+seed)
		//solver = exec.Command("minisat")
	case "glucose":
		solver = exec.Command("glucose", "-rnd-seed="+seed)
	case "clasp":
		solver = exec.Command("clasp", "--seed="+seed)
	case "lingeling":
		solver = exec.Command("lingeling")
	case "cmsat":
		solver = exec.Command("cmsat", "--verb=0", "--random="+seed)
		//	case "treengeling":
		//		solver = exec.Command("treengeling")
		//	case "plingeling":
		//		solver = exec.Command("plingeling")
		//	case "dimetheus":
		//		solver = exec.Command("dimetheus", "-seed="+seed)
		//	case "local":
		//		solver = exec.Command("CCAnr", seed)
		//	case "microsat":
		//		solver = exec.Command("microsat")
	default:
		fmt.Println(false, "Solver not available", Solver_flag)
	}
	return
}
