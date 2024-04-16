package cmd

import (
	"fmt"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
)

var getProtocolCmd = &cobra.Command{
	Use:   "protocol",
	Args:  cobra.NoArgs,
	Short: "Command to get protocol's type",
	Long: `This commands reads communication protocol type. 
It communicates with REST API of the simulator and using HTTP GET it reads protocol's types.
Examples:
	vd get protocol 				-> get protocol type
	vd get protocol --apiAddr 127.0.0.1:7070 	-> get protocol type of with not default api addr
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !verifyIPAddr(apiAddr) {
			return fmt.Errorf("wrong HTTP address")
		}

		c := api.NewClient(apiAddr)

		t, err := c.GetProtocol()
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", t)
		return nil
	},
}

func init() {
	getCmd.AddCommand(getProtocolCmd)
}
