package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/e9ctrl/vd/api"
	"github.com/e9ctrl/vd/device"
	"github.com/e9ctrl/vd/server"
	"github.com/e9ctrl/vd/vdfile"
	"github.com/jwalton/gchalk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var apiAddr string

const version = "0.1.0"
const website = "https://vd.e9controls.com"

// based on https://github.com/labstack/echo/blob/4bc3e475e3137b6402933eec5e6fde641e0d2320/echo.go#L264C5-L264C71
const banner = `
          __
 _  _____/ /
| |/ / _  / 
|___/\_,_/ v%s
Easy to use device simulator
%s
____________________________________O/_______
                                    O\
`

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "vd",
	Args:  cobra.ExactArgs(1),
	Short: "vd is a easy to use device simulator",
	Long: `Virtual Device is an open source program that can be used to simulate lab device communication streams. 
It is useful for testing and debugging software that communicates with lab devices, 
as well as for creating virtual lab environments for education and research.
To run simulator create .toml file with device description and run it using following commands:

	vd vdfile.toml
	vd vdfile.toml --listenAddr 127.0.0.1:6666

By default, vd is listenning on 127.0.0.1:9999.`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf(banner, version, website)
		// parse config file
		vdfileMod, err := vdfile.ReadVDFileMod(args[0])
		if err != nil {
			fmt.Printf("Config loading failed %v", err)
			os.Exit(1)
		}

		vdfile, err := vdfile.ReadVDFile(args[0])
		if err != nil {
			fmt.Printf("Config loading failed %v", err)
			os.Exit(1)
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// create device instance using loaded vdfile
		str, err := device.NewDevice(vdfile, vdfileMod)
		if err != nil {
			fmt.Printf("Device creation failed %v", err)
			os.Exit(1)
		}

		ip := viper.GetString("listenAddr")
		if !verifyIPAddr(ip) {
			fmt.Println("Wrong TCP address")
			os.Exit(1)
		}

		// create instance of TCP simulator server
		srv, err := server.New(str, ip)
		if err != nil {
			fmt.Printf("TCP server creation failed %v", err)
			os.Exit(1)
		}

		// run TCP simulator server
		go srv.Start()
		fmt.Println("vd running on ", gchalk.BrightYellow(ip))

		addr := viper.GetString("httpListenAddr")
		if !verifyIPAddr(addr) {
			fmt.Println("Wrong HTTP address")
			os.Exit(1)
		}

		// create instance of HTTP server
		a := api.NewHttpApiServer(str)

		go func() {
			// run HTTP server with REST API
			err = a.Serve(ctx, addr)
			if err != nil {
				fmt.Printf("HTTP server failed %v", err)
				os.Exit(1)
			}
		}()

		<-ctx.Done()
		srv.Stop()
		fmt.Println("vd stopped")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// This is the main method that start Cobra CLI.
func Execute(f fs.FS) {
	vdTemplate = f
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vd.yaml)")

	// The default value from here is not used but it is visible in help, that's why it is left here
	RootCmd.Flags().StringP("httpListenAddr", "", "127.0.0.1:8080", "Address of the HTTP simulator server")
	// Binds viper httpListenAddr flag to cobra httpListenAddr pflag
	viper.BindPFlag("httpListenAddr", RootCmd.Flags().Lookup("httpListenAddr"))
	// Binds viper apiAddr flag to VD_HTTP_LISTEN_ADDR environment variable
	viper.BindEnv("httpListenAddr", "VD_HTTP_LISTEN_ADDR")
	// Set default flag in viper cause the default one from cobra is not used
	viper.SetDefault("httpListenAddr", "127.0.0.1:8080")

	// The default value from here is not used but it is visible in help, that's why it is left here
	RootCmd.Flags().StringP("listenAddr", "", "127.0.0.1:9999", "Virtual device address")
	// Binds viper listenAddr flag to cobra listenAddr pflag
	viper.BindPFlag("listenAddr", RootCmd.Flags().Lookup("listenAddr"))
	// Binds viper apiAddr flag to VD_LISTEN_ADDR environment variable
	viper.BindEnv("listenAddr", "VD_LISTEN_ADDR")
	// Set default flag in viper cause the default one from cobra is not used
	viper.SetDefault("listenAddr", "127.0.0.1:9999")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".vd" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".vd")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
