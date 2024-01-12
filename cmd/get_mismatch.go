package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getMismatchCmd represents the getMismatch command
var getMismatchCmd = &cobra.Command{
	Use:   "mismatch",
	Args:  cobra.NoArgs,
	Short: "Command to get mismatch message",
	Long: `This command reads value of global mismatch. 
It communicates with REST API of the simulator and using HTTP GET it reads mismatch string.
Examples:
	vd get mismatch
	vd get mismatch --httpListenAddr 127.0.0.1:7070
`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("apiAddr")

		if !verifyIPAddr(addr) {
			fmt.Println("Wrong HTTP address")
			os.Exit(1)
		}

		c := api.NewClient(addr)
		res, err := c.GetMismatch()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("%s\n", res)
	},
}

func init() {
	getCmd.AddCommand(getMismatchCmd)
}
