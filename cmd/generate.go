package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [filename]",
	Args:  cobra.ExactArgs(1),
	Short: "Generate example of config file",
	Long: `This commands generate an example of config file in the current directory.
Usage:
	vd generate
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprint(cmd.OutOrStdout(), "Writing vdfile config file...")
		return generateConfig(args[0])
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)
}
