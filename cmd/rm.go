package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var ErrNoSuchLabel = errors.New("no such instance in configuration")

//-----------------------------------------------------------------------------
// Extend type logic for removing instances
//-----------------------------------------------------------------------------

func (s *Solvers) remove(label string) error {
	fmt.Printf("> removing label '%s' from config. . .", label)
	if v, exists := s.Sat[label]; exists {
		delete(s.Sat, label)
		if err := s.sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Sync failed, reverting changes")
			s.Sat[label] = v
			return err
		}
	} else {
		if v, exists := s.Qbf[label]; exists {
			delete(s.Qbf, label)
			if err := s.sync(); err != nil {
				fmt.Fprintf(os.Stderr, "Sync failed, reverting changes")
				s.Qbf[label] = v
				return err
			}
		} else {
			return ErrNoSuchLabel
		}
	}
	fmt.Println("[ok]")
	return nil
}

func (s *Solvers) purge() error {
	s.Sat = make(map[string]SolverInstance)
	s.Qbf = make(map[string]SolverInstance)
	return s.sync()
}

//-----------------------------------------------------------------------------
// Command  logic
//-----------------------------------------------------------------------------

// Flags
var (
	purge bool
)

// rmCmd represents the list command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Removes available solver instances.",
	Long:  "Removes available solver instances.",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 && !purge {
			return
		}

		var s Solvers

		if err := s.load(); err != nil {
			ErrExit(err, 1)
		}
		if purge {
			fmt.Printf("Purging all solver instances. . .")
			if err := s.purge(); err != nil {
				ErrExit(err, 1)
			}
			fmt.Println("[ok]")
			return
		}
		for _, arg := range args {
			if err := s.remove(arg); err != nil {
				fmt.Printf("[error: %s]\n", err.Error())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolVarP(&purge, "purge", "", false, "Remove ALL avaialble solver instances.")
}
