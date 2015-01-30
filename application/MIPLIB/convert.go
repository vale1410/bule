package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

var name = flag.String("file", "test.txt", "Path of the file specifying the Knapsack Problem.")

type Constraint struct {
	name string
}

func main() {
	flag.Parse()

	input, err := ioutil.ReadFile(*name)

	if err != nil {
		panic("Please specifiy correct path to instance. Does not exist")
	}

	b := bytes.NewBuffer(input)

	lines := strings.Split(strings.Trim(b.String(), " "), "\n")

	state := 0

	for _, l := range lines {
		if l == "" {
			continue
		}
		entries := strings.Fields(l)

		switch state {
		case 0:
			{
				if entries[0] == "NAME" {
					state = 1
				}
				fmt.Println("name : ", entries[1])
			}
		case 1:
			{
				if entries[0] == "ROWS" {
					state = 2
				}
			}
		case 2: // rows
			if entries[0] == "COLUMNS" {
				state = 3
			} else {
				state = 2
			}
		case 3: // rows
			if entries[0] == "*END*" {
				state = 3
			} else {
				state = 2
			}
		case 2:
			{
				ref, b1 := strconv.Atoi(entries[3])
				amb, b2 := strconv.Atoi(entries[4])
				if b1 != nil || b2 != nil {
					panic("bad conversion of numbers")
				}
				fmt.Println("data(", strings.ToLower(entries[1]), ",", -ref, ",", -amb, ").")
				state = 1
			}
		case 3:
			break
		}
	}
	return
}
