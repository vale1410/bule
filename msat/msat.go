package main

import (
	"time"
)

var Timeout_flag = flags.Int(2, "-t", "Timeout in seconds")
var Seed_flag = flags.Int(0, "-s", "Random Seed")

func main() {

	flags.Parse()
	var Solver_flag *int

	timeout := make(chan bool, 1)
	time_total := time.Now()

	go func() {
		time.Sleep(time.Duration(*Timeout_flag) * time.Second)
		timeout <- true
	}()

	solver := getSolver
	stdin, err := solver.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdout, err := solver.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = solver.Start()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		time_before := time.Now()
		defer stdin.Close()
		defer wg.Done()
		io.Copy(stdin, os.Stdin)
		fmt.Printf("Read :\t%.3f s\n", time.Since(time_before).Seconds())
	}()

	go func() {
		defer wg.Done()
		r := bufio.NewReader(stdout)
		s, err := r.ReadString('\n')
		for {
			if strings.HasPrefix(s, "v ") {
				//res.assignment += s[1:]
			} else if strings.HasPrefix(s, "s ") {
				if strings.Contains(s, "UNSATISFIABLE") {
					res.solved = true
					res.satisfiable = false
				} else if strings.Contains(s, "SATISFIABLE") {
					res.solved = true
					res.satisfiable = true
				} else {
					res.solved = false
					panic("whats up? Result of sat solver does not contain proper answer!")
				}
			}
			s, err = r.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err.Error())
			}
		}
	}()

	wg.Wait()
	//	err_tmp := solver.Wait()
	//
	//	if err_tmp != nil {
	//		fmt.Println("return value:", err_tmp())
	//	}

}

func getSolver() (solver *exec.Cmd) {

	if len(os.Args) > 0 {
		Solver_flag = os.Args[0]
	}

	seed := strconv.FormatInt(*Seed_flag, 10)

	switch Solver_flag {
	case "minisat":
		//solver = exec.Command("minisat", "-rnd-seed=123")
		solver = exec.Command("minisat", "-rnd-seed="+seed)
		//solver = exec.Command("minisat")
	case "glucose":
		solver = exec.Command("glucose")
	case "clasp":
		solver = exec.Command("clasp")
	case "lingeling":
		solver = exec.Command("lingeling")
	case "cmsat":
		solver = exec.Command("cmsat")
	case "clasp":
		solver = exec.Command("clasp")
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
