package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
)

var getDelayCmd = &cobra.Command{
	Use:   "delay [command name]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to get value of delays",
	Long: `This commands reads value command delays. 
It communicates with REST API of the simulator and using HTTP GET it reads specified delays.
Examples:
	vd get delay get_temperature 				-> get response delay of get temperature command
	vd get delay set_voltage 				-> get acknowledge delay of set voltage command
	vd get delay set_voltage --apiAddr 127.0.0.1:7070 	-> get acknowledge delay of set voltage command with not default api addr
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("wrong HTTP address")
		}

		c := api.NewClient(apiAddr)

		t, err := c.GetCommandDelay(args[0])
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", t)
		return nil
	},
}

func init() {
	getCmd.AddCommand(getDelayCmd)
}
