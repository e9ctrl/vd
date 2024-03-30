package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
)

var getTypeCmd = &cobra.Command{
	Use:   "type [parameter name]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to get parameter's type",
	Long: `This commands reads type of the parameter. 
It communicates with REST API of the simulator and using HTTP GET it reads specified parameter's types.
Examples:
	vd get type current 				-> get type of the current parameter
	vd get type status 				-> get type of the status parameter
	vd get type mode --apiAddr 127.0.0.1:7070 	-> get type of the mode parameter with not default api addr
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("wrong HTTP address")
		}

		c := api.NewClient(apiAddr)

		t, err := c.GetParameterType(args[0])
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", t)
		return nil
	},
}

func init() {
	getCmd.AddCommand(getTypeCmd)
}
