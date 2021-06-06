package cmd

import (
	//"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// Command  logic
//-----------------------------------------------------------------------------

// Flags
var (
	purge bool
	rmsat bool
	rmqbf bool
)

func init() {
	// nothing
}

// rmCmd represents the list command
var rmCmd = &cobra.Command{
	Use:               "rm",
	Short:             "Removes available solver instances.",
	Long:              "Removes available solver instances.",
	ValidArgsFunction: autoCompleteSolverInstance,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 && !purge {
			return
		}

		var s Solvers

		if err := s.load(); !err.isNil() {
			BuleExit(os.Stdout, err)
		}
		if purge {
			fmt.Printf("Purging all solver instances. . .")
			if err := s.purge(); !err.isNil() {
				BuleExit(os.Stderr, err)
			}
			fmt.Println("OK")
			return
		}
		for _, label := range args {
			// check double existance
			if _, exists := s.Qbf[label]; exists {
				if _, exists := s.Sat[label]; exists {
					if !rmsat && !rmqbf {
						fmt.Printf("Ambiguous %s instance, have both! Please annotate --SAT or --QBF.\n", label)
						return
					}
				}
			}
			if !(rmsat && rmqbf) {
				if rmsat {
					fmt.Printf("Removing SAT label '%s' from config. . .", label)
					if err := s.removeSat(label); !err.isNil() {
						BuleLogErr(os.Stderr, err)
					} else {
						fmt.Println("OK")
					}
					return
				}
				if rmqbf {
					fmt.Printf("Removing QBF label '%s' from config. . .", label)
					if err := s.removeQbf(label); !err.isNil() {
						BuleLogErr(os.Stderr, err)
					} else {
						fmt.Println("OK")
					}
					return
				}
			}
			fmt.Printf("Removing label '%s' from config. . .", label)
			if err := s.remove(label); !err.isNil() {
				BuleLogErr(os.Stderr, err)
			} else {
				fmt.Println("OK")
			}
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolVarP(&purge, "purge", "", false, "Remove ALL available solvers.")
	rmCmd.Flags().BoolVarP(&rmsat, "SAT", "", false, "Remove a particular SAT instance.")
	rmCmd.Flags().BoolVarP(&rmqbf, "QBF", "", false, "Remove a particular QBF instance.")
}
