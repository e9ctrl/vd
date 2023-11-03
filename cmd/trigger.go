package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// triggerCmd represents the trigger command
var triggerCmd = &cobra.Command{
	Use:   "trigger [parameter name]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to trigger the sending of the parameter value to the client ",
	Long: `This commands causes sending the current value of the specified in the
argument parameter name is sent to the connected TCP client."
Examples:
	vd trigger current
	vd trigger voltage --httpListenAddr 127.0.0.1:7070
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
		err = c.TriggerParam(args[0])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("OK")
	},
}

func init() {
	rootCmd.AddCommand(triggerCmd)
	triggerCmd.PersistentFlags().StringP("apiAddr", "", "127.0.0.1:8080", "VD HTTP API address")
	viper.AutomaticEnv()
	viper.BindPFlag("apiAddr", getCmd.Flags().Lookup("apiAddr"))
	viper.BindEnv("apiAddr", "VD_API_ADDR")
}
