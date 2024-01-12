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
	Use:   "get [parameter name]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to get value of any parameter",
	Long: `This command reads value of any parameter.
It communicates with REST API of the simulator and using HTTP GET it reads value of the specified parameter.
Examples:
	vd get current
	vd get voltage --httpListenAddr 127.0.0.1:7070
`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("apiAddr")

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
	// The default value from here is not used but it is visible in help, that's why it is left here
	getCmd.PersistentFlags().StringP("apiAddr", "a", "127.0.0.1:8080", "VD HTTP API address")
	// Binds viper apiAddr flag to cobra apiAddr pflag
	viper.BindPFlag("apiAddr", getCmd.Flags().Lookup("apiAddr"))
	// Binds viper apiAddr flag to VD_API_ADDR environment variable
	viper.BindEnv("apiAddr", "VD_API_ADDR")
	// Set default flag in viper cause the default one from cobra is not used
	viper.SetDefault("apiAddr", "127.0.0.1:8080")
}
