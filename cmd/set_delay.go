package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("apiAddr")

		if !verifyIPAddr(addr) {
			fmt.Println("Wrong HTTP address")
			os.Exit(1)
		}

		c := api.NewClient(addr)

		err := c.SetCommandDelay(args[0], args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("OK")
	},
}

func init() {
	setCmd.AddCommand(setDelayCmd)
}
