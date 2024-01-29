package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
)

var setDelayCmd = &cobra.Command{
	Use:   "delay [command name] [value]",
	Args:  cobra.ExactArgs(2),
	Short: "Command to set value of delays",
	Long: `The command sets value of command delays.
It communicates with REST API of the simulator and using HTTP POST verb modifies value of the specified delay.
Examples:
	vd set delay get_temp 100ms	-> set response delay of get temp command
	vd set delay set_volt 1m	-> set response delay of set volt command
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("wrong HTTP address")
		}

		c := api.NewClient(apiAddr)
		err := c.SetCommandDelay(args[0], args[1])
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), "OK\n")
		return nil
	},
}

func init() {
	setCmd.AddCommand(setDelayCmd)
}
