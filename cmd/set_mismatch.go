package cmd

import (
	"fmt"
	"os"

	"github.com/e9ctrl/vd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setMismatchCmd represents the setMismatch command
var setMismatchCmd = &cobra.Command{
	Use:   "mismatch [value]",
	Args:  cobra.ExactArgs(1),
	Short: "Command to set mismatch message",
	Long: `This command sets value of global mismatch. 
It communicates with REST API of the simulator and using
HTTP POST verb modifies value of the mismatch message.
Examples:
	vd set mismatch "wrong parameter"
	vd set mismatch error
`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("apiAddr")

		if !verifyIPAddr(addr) {
			fmt.Println("Wrong HTTP address")
			os.Exit(1)
		}

		c := api.NewClient(addr)
		err = c.SetMismatch(args[0])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("OK")
	},
}

func init() {
	getCmd.AddCommand(setMismatchCmd)
}
