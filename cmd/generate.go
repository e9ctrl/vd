package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Args:  cobra.NoArgs,
	Short: "Generate example of config file",
	Long: `This commands generate an example of .TOML config file in the current directory.
Usage:
	vd generate
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprint(cmd.OutOrStdout(), "Writing vdfile config file...")
		return generateConfig()
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)
}
