package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:   "get [parameter name]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to get value of any parameter",
	Long: `This command reads value of any parameter.
It communicates with REST API of the simulator and using HTTP GET it reads value of the specified parameter.
Examples:
	vd get current
	vd get voltage --apiAddr 127.0.0.1:7070
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("Wrong HTTP address")
		}

		c := api.NewClient(apiAddr)
		res, err := c.GetParameter(args[0])
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", res)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
	// The default value from here is not used but it is visible in help, that's why it is left here
	getCmd.PersistentFlags().StringVarP(&apiAddr, "apiAddr", "a", "127.0.0.1:8080", "VD HTTP API address")
	// Binds viper apiAddr flag to cobra apiAddr pflag
	viper.BindPFlag("apiAddr", getCmd.PersistentFlags().Lookup("apiAddr"))
	// Binds viper apiAddr flag to VD_API_ADDR environment variable
	viper.BindEnv("apiAddr", "VD_API_ADDR")
}
