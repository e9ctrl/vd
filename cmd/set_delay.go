package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
)

var setDelayCmd = &cobra.Command{
	Use:   "delay [delay_type] [value]\n  vd set delay [delay_type] [parameter name] [value]",
	Args:  cobra.RangeArgs(2, 3),
	Short: "Command to set value of delays",
	Long: `The command sets value of global and parameter delays of two types: reponse and acknowledge.
	It communicates with REST API of the simulator and using HTTP POSRT verb modifies value of the specified delay. 
Examples:
	vd set delay res 10s		-> set global response delay
	vd set delay ack 10s		-> set global acknowledge delay
	vd set delay res temp 100ms	-> set response delay of temp parameter
	vd set delay ack volt 1m	-> set acknowledge delay of volt parameter
`,
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := cmd.Flags().GetString("apiAddr")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if !verifyIPAddr(addr) {
			fmt.Println("Wrong HTTP address")
			os.Exit(1)
		}

		c := api.NewClient(addr)
		switch len(args) {
		case 2:
			err = c.SetGlobalDelay(args[0], args[1])
		case 3:
			err = c.SetParamDelay(args[0], args[1], args[2])
		}

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("OK")
	},
}

func init() {
	setCmd.AddCommand(setDelayCmd)
}
