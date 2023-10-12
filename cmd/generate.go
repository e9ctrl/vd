package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Args:  cobra.NoArgs,
	Short: "Generate example of config file",
	Long: `This commands generate an example of .TOML config file in the current directory.
Usage:
	vd generate
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Writing vdfile config file...")
		err := generateConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
