package cmd_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/e9ctrl/vd/api"
	"github.com/e9ctrl/vd/cmd"
	"github.com/e9ctrl/vd/device"
	"github.com/e9ctrl/vd/vdfile"
)

const (
	FILE     = "../vdfile/vdfile"
	API_ADDR = "127.0.0.1:7777"
)

func TestMain(m *testing.M) {
	config, err := vdfile.DecodeVDFile(FILE)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(config.Commands); i++ {
		switch config.Commands[i].Name {
		case "get_psi":
			config.Commands[i].Dly = "3s"
		case "get_temp":
			config.Commands[i].Dly = "1s"
		case "get_mode":
			config.Commands[i].Dly = "5s"
		}
	}
	config.Mismatch = "Wrong query"

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	vdfile, err := vdfile.ReadVDFileFromConfig(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// create stream device
	d, err := device.NewDevice(vdfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// create instance of HTTP server
	a := api.NewHttpApiServer(d)

	go func() {
		// run HTTP server with REST API
		err = a.Serve(ctx, API_ADDR)
		if err != nil {
			fmt.Fprintf(os.Stderr, "HTTP server failed %v", err)
			os.Exit(1)
		}
	}()

	// wait for http server
	<-time.After(time.Second)

	// run tests
	exitCode := m.Run()

	stop()
	os.Exit(exitCode)
}

func execute(args []string) string {
	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)
	cmd.RootCmd.SetArgs(args)
	cmd.RootCmd.SilenceUsage = true
	cmd.RootCmd.Execute()

	return buf.String()
}

func TestGetMismatch(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		api  string
	}{
		{"get mismatch", "Wrong query\n", API_ADDR},
		{"wrong api addr", `Error: Get "http://127.0.0.1:5555/mismatch": dial tcp 127.0.0.1:5555: connect: connection refused` + "\n", "127.0.0.1:5555"},
		{"wrong api addr format", "Error: wrong HTTP address\n", "127.test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := fmt.Sprintf("get mismatch --apiAddr %s", tt.api)
			in := strings.Split(str, " ")
			res := execute(in)
			if res != tt.exp {
				t.Errorf("exp value: %s got %s\n", tt.exp, res)
			}
		})
	}
}

func TestGetParameter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
		api   string
	}{
		{"get version", "version", "version 1.0\n", API_ADDR},
		{"get mode", "mode", "NORM\n", API_ADDR},
		{"wrong api addr", "version", `Error: Get "http://127.0.0.1:5555/version": dial tcp 127.0.0.1:5555: connect: connection refused` + "\n", "127.0.0.1:5555"},
		{"wrong api addr format", "version", "Error: wrong HTTP address\n", "127.test"},
		{"wrong cmd", "test", "Error: API error Error: parameter not found: test\n", API_ADDR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := fmt.Sprintf("get %s --apiAddr %s", tt.input, tt.api)
			in := strings.Split(str, " ")
			res := execute(in)
			if res != tt.exp {
				t.Errorf("exp value: %s got %s\n", tt.exp, res)
			}
		})
	}
}

func TestGetDelay(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
		api   string
	}{
		{"get_psi delay", "get_psi", "3s\n", API_ADDR},
		{"wrong api addr", "get_psi", `Error: Get "http://127.0.0.1:5555/delay/get_psi": dial tcp 127.0.0.1:5555: connect: connection refused` + "\n", "127.0.0.1:5555"},
		{"wrong api addr format", "get_psi", "Error: wrong HTTP address\n", "127.test"},
		{"wrong cmd", "get_test", "Error: API error Error: command not found: get_test\n", API_ADDR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := fmt.Sprintf("get delay %s --apiAddr %s", tt.input, tt.api)
			in := strings.Split(str, " ")
			res := execute(in)
			if res != tt.exp {
				t.Errorf("exp value: %s got %s\n", tt.exp, res)
			}
		})
	}
}

func TestTrigger(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
		api   string
	}{
		{"trig get_current", "get_current", "Error: API error Error: no client available\n", API_ADDR},
		{"wrong api addr", "get_psi", `Error: Post "http://127.0.0.1:5555/trigger/get_psi": dial tcp 127.0.0.1:5555: connect: connection refused` + "\n", "127.0.0.1:5555"},
		{"wrong api addr format", "get_psi", "Error: wrong HTTP address\n", "127.test"},
		{"wrong cmd", "get_test", "Error: API error Error: command not found: get_test\n", API_ADDR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := fmt.Sprintf("trigger %s --apiAddr %s", tt.input, tt.api)
			in := strings.Split(str, " ")
			res := execute(in)
			if res != tt.exp {
				t.Errorf("exp value: %s got %s\n", tt.exp, res)
			}
		})
	}
}

func TestSetParameter(t *testing.T) {
	res := execute([]string{"set", "current", "20", "--apiAddr", API_ADDR})

	expected := "OK\n"
	if res != expected {
		t.Errorf("exp value: %s got %s\n", expected, res)
	}

	res = execute([]string{"get", "current", "--apiAddr", API_ADDR})
	expected = "20\n"
	if res != expected {
		t.Errorf("exp value: %s got %s\n", expected, res)
	}
}

func TestSetParameterWrong(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
		api   string
	}{
		{"wrong api addr", "current 30", `Error: Post "http://127.0.0.1:5555/current/30": dial tcp 127.0.0.1:5555: connect: connection refused` + "\n", "127.0.0.1:5555"},
		{"wrong api addr format", "current 30", "Error: wrong HTTP address\n", "127.test"},
		{"wrong set value", "current test", "Error: API error Error: received param type that cannot be converted to int\n", API_ADDR},
		{"wrong param", "test 20", "Error: API error Error: parameter not found: test\n", API_ADDR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := fmt.Sprintf("set %s --apiAddr %s", tt.input, tt.api)
			in := strings.Split(str, " ")
			res := execute(in)
			if res != tt.exp {
				t.Errorf("exp value: %s got %s\n", tt.exp, res)
			}
		})
	}
}

func TestSetMismatch(t *testing.T) {
	res := execute([]string{"set", "mismatch", "Error", "--apiAddr", API_ADDR})

	expected := "OK\n"
	if res != expected {
		t.Errorf("exp value: %s got %s\n", expected, res)
	}

	res = execute([]string{"get", "mismatch", "--apiAddr", API_ADDR})

	expected = "Error\n"
	if res != expected {
		t.Errorf("exp value: %s got %s\n", expected, res)
	}
}

func TestSetMismatchWrong(t *testing.T) {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 256)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	mis := string(b)

	tests := []struct {
		name  string
		input string
		exp   string
		api   string
	}{
		{"wrong api addr", "error", `Error: Post "http://127.0.0.1:5555/mismatch/error": dial tcp 127.0.0.1:5555: connect: connection refused` + "\n", "127.0.0.1:5555"},
		{"wrong api addr format", "error", "Error: wrong HTTP address\n", "127.test"},
		{"too long message", mis, "Error: API error Error: new mismatch message exceeded 255 characters limit: " + mis + "\n", API_ADDR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := fmt.Sprintf("set mismatch %s --apiAddr %s", tt.input, tt.api)
			in := strings.Split(str, " ")
			res := execute(in)
			if res != tt.exp {
				t.Errorf("exp value: %s got %s\n", tt.exp, res)
			}
		})
	}
}

func TestSetDelay(t *testing.T) {
	res := execute([]string{"set", "delay", "get_temp", "5s", "--apiAddr", API_ADDR})

	expected := "OK\n"
	if res != expected {
		t.Errorf("exp value: %s got %s\n", expected, res)
	}

	res = execute([]string{"get", "delay", "get_temp", "--apiAddr", API_ADDR})

	expected = "5s\n"
	if res != expected {
		t.Errorf("exp value: %s got %s\n", expected, res)
	}
}

func TestSetDelayWrong(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
		api   string
	}{
		{"wrong api addr", "get_psi 10s", `Error: Post "http://127.0.0.1:5555/delay/get_psi/10s": dial tcp 127.0.0.1:5555: connect: connection refused` + "\n", "127.0.0.1:5555"},
		{"wrong api addr format", "get_psi 10s", "Error: wrong HTTP address\n", "127.test"},
		{"wrong set value", "get_psi test", "Error: API error Error: time: invalid duration \"test\"\n", API_ADDR},
		{"wrong cmd", "get_test 10s", "Error: API error Error: command not found: get_test\n", API_ADDR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := fmt.Sprintf("set delay %s --apiAddr %s", tt.input, tt.api)
			in := strings.Split(str, " ")
			res := execute(in)
			if res != tt.exp {
				t.Errorf("exp value: %s got %s\n", tt.exp, res)
			}
		})
	}
}

func TestCLIEnvVars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
	}{
		{"get param", "get hex", "255\n"},
		{"set param", "set hex0 240", "OK\n"},
		{"get delay", "get delay get_mode", "5s\n"},
		{"set delay", "set delay get_mode 10s", "OK\n"},
		{"get mismatch", "get mismatch", "Error\n"},
		{"set mismtach", "set mismatch ERR", "OK\n"},
		{"trigger", "trigger get_temp", "Error: API error Error: no client available\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("VD_API_ADDR", API_ADDR)
			in := strings.Split(tt.input, " ")
			res := execute(in)
			os.Unsetenv("VD_API_ADDR")
			if res != tt.exp {
				t.Errorf("exp value: %s got %s\n", tt.exp, res)
			}
		})
	}
}
