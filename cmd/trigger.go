package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger [parameter name]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to trigger the sending of the parameter value to the client ",
	Long: `This commands causes sending the current value of the specified in the
argument parameter name is sent to the connected TCP client."
Examples:
	vd trigger current
	vd trigger voltage --apiAddr 127.0.0.1:7070
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("Wrong HTTP address")
		}

		c := api.NewClient(apiAddr)
		err := c.Trigger(args[0])
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), "OK\n")
		return nil
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
