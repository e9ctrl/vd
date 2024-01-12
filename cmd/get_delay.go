package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/e9ctrl/vd/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getDelayCmd = &cobra.Command{
	Use:   "delay\n  vd get delay [command name]",
	Args:  cobra.RangeArgs(1, 2),
	Short: "Command to get value of delays",
	Long: `This commands reads value command delays. 
	It communicates with REST API of the simulator and using HTTP GET ir reads specified delays. 
Examples:
	vd get delay get_temperature 				-> get response delay of get temperature command
	vd get delay set_voltage 				-> get acknowledge delay of set voltage command
	vd get delay set_voltage --apiAddr 127.0.0.1:7070 	-> get acknowledge delay of set voltage command with not default api addr
`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("apiAddr")

		if !verifyIPAddr(addr) {
			fmt.Println("Wrong HTTP address")
			os.Exit(1)
		}

		c := api.NewClient(addr)

		var t time.Duration
		switch len(args) {
		case 1:
			t, err = c.GetCommandDelay(args[0])
		default:
			fmt.Println("Invalid number of arguments.")
			os.Exit(1)
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", t)
	},
}

func init() {
	getCmd.AddCommand(getDelayCmd)
}
