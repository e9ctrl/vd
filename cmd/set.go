package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set [parameter name] [value]",
	Args:  cobra.ExactArgs(2),
	Short: "Command to set value of any parameter",
	Long: `The command sets value of any parameter.
It communicates with REST API of the simulator and using
HTTP POST verb modifies value of the specified parameter inside the simulator.
Examples:
	vd set current 20
	vd set voltage 3.5 --httpListenAddr 192.168.56.100:9999
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
		err = c.SetParameter(args[0], args[1])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("OK")
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.PersistentFlags().StringP("apiAddr", "", "127.0.0.1:8080", "VD HTTP API address")
	viper.AutomaticEnv()
	viper.BindPFlag("apiAddr", setCmd.Flags().Lookup("apiAddr"))
	viper.BindEnv("apiAddr", "VD_API_ADDR")
}
