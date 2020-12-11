package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func formatOutput(st SolverT, msi map[string]SolverInstance) {
	fmt.Printf("[%v]\n", st)
	if len(msi) == 0 {
		fmt.Printf("%15s\n", "-")
	}
	if si, exists := msi["default"]; exists {
		fmt.Printf("%15s:\t%s\t%s\n", "(default)", si.Prog, si.Flags)
		delete(msi, "default")
	}
	for k, v := range msi {
		fmt.Printf("%15s:\t%s\t%s\n", k, v.Prog, v.Flags)
	}
}

//-----------------------------------------------------------------------------
// Command  logic
//-----------------------------------------------------------------------------

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available solvers.",
	Long:  "List available solvers.",
	Run: func(cmd *cobra.Command, args []string) {
		var s Solvers
		if err := s.load(); err != nil {
			ErrExit(err, 1)
		}
		fmt.Println()
		formatOutput(QBF, s.Qbf)
		formatOutput(SAT, s.Sat)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
