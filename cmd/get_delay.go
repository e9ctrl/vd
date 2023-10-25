package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"
	"github.com/spf13/cobra"
)

var getDelayCmd = &cobra.Command{
	Use:   "delay [delay_type]\n  vd get delay [delay_type] [parameter name]",
	Args:  cobra.RangeArgs(1, 2),
	Short: "Command to get value of delays",
	Long: `This commands reads value of global and parameter delays of two types: reponse and acknowledge. 
	It communicates with REST API of the simulator and using HTTP GET ir reads specified delays. 
Examples:
	vd get delay res 		-> get global response delay
	vd get delay ack 		-> get global acknowledge delay
	vd get delay res temperature 	-> get response delay of temperature parameter
	vd get dleay ack voltage 	-> get acknowledge delay of voltage parameter
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
		var param string
		if len(args) == 1 {
			param = fmt.Sprintf("delay/%s", args[0])
		} else {
			param = fmt.Sprintf("delay/%s/%s", args[0], args[1])
		}

		res, err := c.GetParameter(param)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("%s\n", res)
	},
}

func init() {
	getCmd.AddCommand(getDelayCmd)
}
