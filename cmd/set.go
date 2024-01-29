package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setCmd = &cobra.Command{
	Use:   "set [parameter name] [value]",
	Args:  cobra.ExactArgs(2),
	Short: "Command to set value of any parameter",
	Long: `The command sets value of any parameter.
It communicates with REST API of the simulator and using
HTTP POST verb modifies value of the specified parameter inside the simulator.
Examples:
	vd set current 20
	vd set voltage 3.5 --apiAddr 192.168.56.100:9999
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("Wrong HTTP address")
		}

		c := api.NewClient(apiAddr)
		err := c.SetParameter(args[0], args[1])
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), "OK\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.PersistentFlags().StringP("apiAddr", "a", "127.0.0.1:8080", "VD HTTP API address")
	// Binds viper apiAddr flag to cobra apiAddr pflag
	viper.BindPFlag("apiAddr", setCmd.PersistentFlags().Lookup("apiAddr"))
	// Binds viper apiAddr flag to VD_API_ADDR environment variable
	viper.BindEnv("apiAddr", "VD_API_ADDR")
}
