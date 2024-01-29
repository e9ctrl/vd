package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
)

var getMismatchCmd = &cobra.Command{
	Use:   "mismatch",
	Args:  cobra.NoArgs,
	Short: "Command to get mismatch message",
	Long: `This command reads value of global mismatch. 
It communicates with REST API of the simulator and using HTTP GET it reads mismatch string.
Examples:
	vd get mismatch
	vd get mismatch --apiAddr 127.0.0.1:7070
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("wrong HTTP address")
		}

		c := api.NewClient(apiAddr)
		res, err := c.GetMismatch()
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", res)
		return nil
	},
}

func init() {
	getCmd.AddCommand(getMismatchCmd)
}
