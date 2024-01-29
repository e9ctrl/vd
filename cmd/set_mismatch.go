package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
)

var setMismatchCmd = &cobra.Command{
	Use:   "mismatch [value]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to set mismatch message",
	Long: `This command sets value of global mismatch. 
It communicates with REST API of the simulator and using
HTTP POST verb modifies value of the mismatch message.
Examples:
	vd set mismatch "wrong parameter"
	vd set mismatch error
	vd set mismatch test --apiAddr 127.0.0.1:7070
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("wrong HTTP address")
		}

		c := api.NewClient(apiAddr)
		err := c.SetMismatch(args[0])
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), "OK\n")
		return nil
	},
}

func init() {
	getCmd.AddCommand(setMismatchCmd)
}
