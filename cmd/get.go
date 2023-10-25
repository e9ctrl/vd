package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.ExactArgs(1),
	Short: "Command to get value of any parameter",
	Long: `This command reads value of any parameter. It communicates with REST API of the simulator and using HTTP GET it reads value of the specified parameter.
Usage:
	vd get current
	vd get voltage --httpListenAddr 127.0.0.1:7070
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
		res, err := c.GetParameter(args[0])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("%s\n", res)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().StringP("apiAddr", "", "127.0.0.1:8080", "VD HTTP API address")
	viper.AutomaticEnv()
	viper.BindPFlag("apiAddr", getCmd.Flags().Lookup("apiAddr"))
	viper.BindEnv("apiAddr", "VD_API_ADDR")
}
