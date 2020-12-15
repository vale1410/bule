/*
Copyright © 2020 Sebastian J <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

$ source <(bule completion bash)

# To load completions for each session, execute once:
Linux:
  $ bule completion bash > /etc/bash_completion.d/bule
MacOS:
  $ bule completion bash > /usr/local/etc/bash_completion.d/bule

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ bule completion zsh > "${fpath[1]}/_bule"

# You will need to start a new shell for this setup to take effect.

Fish:

$ bule completion fish | source

# To load completions for each session, execute once:
$ bule completion fish > ~/.config/fish/completions/bule.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

// Autocompletion rules
func autoCompleteBuleFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"bul", "bule"}, cobra.ShellCompDirectiveFilterFileExt
}

func autoCompleteSolverInstance(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var s Solvers
	if err := s.load(); !err.isNil() {
		BuleExit(os.Stderr, err)
	}
	labelsSat := make([]string, 0, len(s.Sat))
	labelsQbf := make([]string, 0, len(s.Qbf))
	labelsAll := make([]string, 0, cap(labelsSat)+cap(labelsQbf))
	// add Sat instances
	for label := range s.Sat {
		labelsSat = append(labelsSat, label)
	}
	// add Qbf instances
	for label := range s.Qbf {
		labelsQbf = append(labelsQbf, label)
	}
	// sort them and make default instance 1st
	sortSwap(&labelsSat, "default")
	sortSwap(&labelsQbf, "default")
	// merge sorted instances
	for _, label := range labelsSat {
		labelsAll = append(labelsAll, label)
	}
	for _, label := range labelsQbf {
		labelsAll = append(labelsAll, label)
	}
	return labelsAll, cobra.ShellCompDirectiveDefault
}
