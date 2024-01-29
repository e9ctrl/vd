package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger [command name]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to trigger the sending of the parameter value to the client ",
	Long: `This commands causes sending the current value of the specified parameter to the
connected TCP client. As a argument it is required to pass corresponding getter command name.
Examples:
	vd trigger get_current
	vd trigger get_voltage --apiAddr 127.0.0.1:7070
`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("apiAddr")

		if !verifyIPAddr(addr) {
			fmt.Println("Wrong HTTP address")
			os.Exit(1)
		}

		c := api.NewClient(addr)
		err := c.Trigger(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("OK")
	},
}

func init() {
	rootCmd.AddCommand(triggerCmd)
	triggerCmd.PersistentFlags().StringP("apiAddr", "a", "127.0.0.1:8080", "VD HTTP API address")
	// Binds viper apiAddr flag to cobra apiAddr pflag
	viper.BindPFlag("apiAddr", getCmd.PersistentFlags().Lookup("apiAddr"))
	// Binds viper apiAddr flag to VD_API_ADDR environment variable
	viper.BindEnv("apiAddr", "VD_API_ADDR")
}
