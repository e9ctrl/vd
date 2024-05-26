package main

import (
	"bytes"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/e9ctrl/vd/device"
	"github.com/e9ctrl/vd/server"
	"github.com/e9ctrl/vd/vdfile"
)

var (
	vdfileStream         vdfile.ConfigStream
	vdfileStreamDelay    vdfile.ConfigStream
	vdfileStreamMismatch vdfile.ConfigStream
	vdfileModbus         vdfile.ConfigModbus
)

const (
	FILE_STREAM = "vdfile/vdfile_stream"
	FILE_MODBUS = "vdfile/vdfile_modbus"
	ADDR1       = "localhost:3333"
	ADDR2       = "localhost:4444"
	ADDR3       = "localhost:5555"
	ADDR4       = "localhost:6666"
	ADDR5       = "localhost:2222"
	ADDR6       = "localhost:2233"
	ADDR7       = "localhost:8888"
)

func init() {
	config, err := vdfile.DecodeVDFileStream(FILE_STREAM)
	if err != nil {
		panic(err)
	}

	vdfileStream = config

	config1, _ := vdfile.DecodeVDFileStream(FILE_STREAM)
	for i := 0; i < len(config1.Commands); i++ {
		switch config1.Commands[i].Name {
		case "get_psi":
			config1.Commands[i].Dly = "3s"
		case "get_temp":
			config1.Commands[i].Dly = "1s"
		case "set_psi":
			config1.Commands[i].Dly = "3s"
		case "get_current":
			config1.Commands[i].Dly = "2s"
		case "set_current":
			config1.Commands[i].Dly = "2s"
		}
	}
	vdfileStreamDelay = config1

	config2, _ := vdfile.DecodeVDFileStream(FILE_STREAM)

	config2.Mismatch = "Wrong query"
	vdfileStreamMismatch = config2

	configModbus, err := vdfile.DecodeVDFileModbus(FILE_MODBUS)
	if err != nil {
		panic(err)
	}
	vdfileModbus = configModbus
}

func setupStreamTestCase(t *testing.T, addr string, vd vdfile.ConfigStream) func() {
	vdfile, err := vdfile.ReadVDFileStreamFromConfig(vd)
	if err != nil {
		t.Fatal(err)
	}

	//create stream device
	d, err := device.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	// create TCP server
	s, err := server.New(d, addr)
	if err != nil {
		t.Fatalf("error while creating server %v\n", err)
	}

	s.Start()

	return func() {
		s.Stop()
	}
}

func setupModbusTestCase(t *testing.T, addr string, vd vdfile.ConfigModbus) func() {
	vdfile, err := vdfile.ReadVDFileModbusFromConfig(vd)
	if err != nil {
		t.Fatal(err)
	}

	//create stream device
	d, err := device.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	// create TCP server
	s, err := server.New(d, addr)
	if err != nil {
		t.Fatalf("error while creating server %v\n", err)
	}

	s.Start()

	return func() {
		s.Stop()
	}
}

func TestRunStream(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer setupStreamTestCase(t, ADDR1, vdfileStream)()
	// connect to server
	conn, err := net.Dial("tcp", ADDR1)
	if err != nil {
		t.Fatalf("could not connect to to server: %v\n", err)
	}
	defer conn.Close()
	// set timeout for reading data
	conn.SetReadDeadline(time.Now().Add(time.Second))

	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"current check", []byte("CUR?\r\n"), []byte("CUR 300\r\n")},
		{"psi check", []byte("PSI?\r\n"), []byte("PSI 3.30\r\n")},
		{"version check", []byte("VER?\r\n"), []byte("version 1.0\r\n")},
		{"mode check", []byte(":PULSE0:MODE?\r\n"), []byte("NORM\r\n")},
		{"temp check", []byte("TEMP?\r\n"), []byte("TEMP 2.30\r\n")},
		{"ack check", []byte("ACK?\r\n"), []byte("false\r\n")},
		{"current set", []byte("CUR 20\r\n"), []byte("OK\r\n")},
		{"psi set", []byte("PSI 3.46\r\n"), []byte("PSI 3.46 OK\r\n")},
		{"mode set", []byte(":PULSE0:MODE SING\r\n"), []byte("ok\r\n")},
		{"hex check", []byte("HEX?\r\n"), []byte("0x0FF\r\n")},
		{"hex 0 check", []byte("HEX0?\r\n"), []byte("0FF\r\n")},
		{"set hex", []byte("HEX 0x03F\r\n"), []byte("HEX 0x03F\r\n")},
		{"set hex 0", []byte("HEX0 ABC\r\n"), []byte("ABC\r\n")},
		{"get status", []byte("S?\r\n"), []byte("version 1.0 - 2.3\r\n")},
		{"set two params", []byte("set mode BURS psi 4.56\r\n"), []byte("ok\r\n")},
		{"get status ch2", []byte("get status ch 2\r\n"), []byte("mode: BURS psi: 4.56\r\n")},
		{"get status ch3", []byte("get status ch 3\r\n"), []byte("mode: BURS\npsi: 4.56\r\n")},
		{"test command that can be parsed in two ways", []byte("set ch1 tec07F\r\n"), []byte("set ch1 tec 7F\r\r\n")},
		// This case needs to be added when "when" functionality will be added
		// otherwise get speed returns one of two responses randomly
		{"get speed", []byte("get speed?\r\n"), []byte("High speed: 36.600000\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := conn.Write(tt.input); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			out := make([]byte, 128)
			if _, err := conn.Read(out); err == nil {
				trimmed := bytes.Trim(out, "\x00")
				if !bytes.Equal(tt.want, trimmed) {
					t.Errorf("exp resp: %[1]v %[1]s got: %[2]v %[2]s\n", tt.want, trimmed)
				}
			} else {
				t.Error("could not read from connection")
			}
		})
	}
}

func TestRunModbus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer setupModbusTestCase(t, ADDR5, vdfileModbus)()
	// connect to server
	conn, err := net.Dial("tcp", ADDR5)
	if err != nil {
		t.Fatalf("could not connect to to server: %v\n", err)
	}
	defer conn.Close()
	// set timeout for reading data
	conn.SetReadDeadline(time.Now().Add(time.Second))

	tests := []struct {
		name   string
		input  []byte
		length int
		want   []byte
	}{
		{"read state", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x01, 0x00, 0x01}, 10, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x01, 0x01, 0x01}},
		{"read temp", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x05, 0x00, 0x01}, 11, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x03, 0x02, 0x00, 0x14}},
		{"read temp2", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x07, 0x00, 0x02}, 13, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x07, 0x01, 0x03, 0x04, 0x00, 0x06, 0xEE, 0x43}},
		{"read state and state1", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x01, 0x00, 0x02}, 10, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x01, 0x01, 0x03}},
		{"read voltage", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x0a, 0x00, 0x04}, 17, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x01, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1e}},
		{"read pressure", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x0f, 0x00, 0x04}, 17, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x01, 0x04, 0x08, 0x40, 0x41, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"read mode", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x03, 0x00, 0x01}, 10, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x01, 0x01}},
		{"read coil not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x14, 0x00, 0x01}, 10, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x01, 0x01, 0x00}},
		{"read di not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x01, 0xc8, 0x00, 0x01}, 10, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x01, 0x00}},
		{"read multiple coils not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x08, 0x00, 0x0a}, 11, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x01, 0x02, 0x00, 0x00}},
		{"read multiple dis not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x17, 0x00, 0x05}, 10, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x01, 0x00}},
		{"read holding register not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x03, 0x00, 0xC7, 0x00, 0x01}, 11, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0xff, 0x03, 0x02, 0x00, 0x00}},
		{"read input register not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x04, 0x00, 0x22, 0x00, 0x01}, 11, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0xFF, 0x04, 0x02, 0x00, 0x00}},
		{"write state", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x01, 0x00, 0x00}, 12, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x01, 0x00, 0x00}},
		{"write state2 and state3", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x0f, 0x00, 0x02, 0x00, 0x02, 0x02, 0x00, 0x00}, 12, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x00, 0x0F, 0x00, 0x02, 0x00, 0x02}},
		{"write single holding register temp", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x05, 0x03, 0x09}, 12, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x05, 0x03, 0x09}},
		{"write multiple holding registers temp2", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x00, 0x10, 0x00, 0x07, 0x00, 0x02, 0x04, 0x03, 0x04, 0x01, 0x04}, 12, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x00, 0x10, 0x00, 0x07, 0x00, 0x02}},
		{"write multiple holding registers not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x01, 0x10, 0x00, 0x1e, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14}, 12, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x10, 0x00, 0x1e, 0x00, 0x04}},
		{"write single holding register not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x19, 0x00, 0x32}, 12, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x19, 0x00, 0x32}},
		{"illegal address", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x01, 0x10, 0x27, 0x0d, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x09}, 9, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0x01, 0x90, 0x02}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := conn.Write(tt.input); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			out := make([]byte, tt.length)
			if _, err := conn.Read(out); err == nil {
				if !bytes.Equal(tt.want, out) {
					t.Errorf("exp resp: %v got: %v\n", tt.want, out)
				}
			} else {
				t.Error("could not read from connection")
			}
		})
	}
}

func TestRunWrongQueriesStream(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer setupStreamTestCase(t, ADDR2, vdfileStream)()
	// connect to server
	conn, err := net.Dial("tcp", ADDR2)

	if err != nil {
		t.Fatalf("could not connect to to server: %v\n", err)
	}
	defer conn.Close()
	// set timeout for reading data
	conn.SetReadDeadline(time.Now().Add(time.Second))

	tests := []struct {
		name  string
		input []byte
	}{
		{"current wrong set string", []byte("CUR test\n")},
		{"current wrong set bool", []byte("CUR false\n")},
		{"current wrong set float", []byte("CUR 32.44\n")},
		{"psi wrong set", []byte("PSI test\n")},
		{"psi wrong int", []byte("PSI 24\n")},
		{"mode wrong set", []byte(":PULSE0:MODE TEST\n")},
		{"set wrong hex", []byte("HEX 12.45\n")},
		{"ack wrong set", []byte("ACK test\n")},
		{"unknown param", []byte("TEST?\n")},
		{"only new line", []byte("\n")},
		{"empty", []byte(nil)},
		{"wrong terminator", []byte("\t")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := conn.Write(tt.input); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			out := make([]byte, 128)
			_, err := conn.Read(out)
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				t.Errorf("exp error: %s got: %s\n", os.ErrDeadlineExceeded, err)
			}
		})
	}
}

func TestRunWrongQueriesModbus(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer setupModbusTestCase(t, ADDR6, vdfileModbus)()
	// connect to server
	conn, err := net.Dial("tcp", ADDR6)

	if err != nil {
		t.Fatalf("could not connect to to server: %v\n", err)
	}
	defer conn.Close()
	// set timeout for reading data
	conn.SetReadDeadline(time.Now().Add(time.Second))

	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte(nil)},
		{"wrong frame length", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x01, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := conn.Write(tt.input); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			out := make([]byte, 128)
			_, err := conn.Read(out)
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				t.Errorf("exp error: %s got: %s\n", os.ErrDeadlineExceeded, err)
			}
		})
	}
}

func TestRunWithDelaysStream(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer setupStreamTestCase(t, ADDR3, vdfileStreamDelay)()
	// connect to server
	conn, err := net.Dial("tcp", ADDR3)
	if err != nil {
		t.Fatalf("could not connect to to server: %v\n", err)
	}
	defer conn.Close()
	tests := []struct {
		name  string
		dur   time.Duration
		input []byte
		want  []byte
	}{
		{"current check", 3 * time.Second, []byte("CUR?\r\n"), []byte("CUR 300\r\n")},
		{"psi check", 4 * time.Second, []byte("PSI?\r\n"), []byte("PSI 3.30\r\n")},
		{"temp check", 2 * time.Second, []byte("TEMP?\r\n"), []byte("TEMP 2.30\r\n")},
		{"current set", 3 * time.Second, []byte("CUR 20\r\n"), []byte("OK\r\n")},
		{"psi set", 4 * time.Second, []byte("PSI 3.46\r\n"), []byte("PSI 3.46 OK\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set timeout for reading data
			conn.SetReadDeadline(time.Now().Add(tt.dur))
			if _, err := conn.Write(tt.input); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			out := make([]byte, 128)
			start := time.Now()
			_, err := conn.Read(out)
			elapsed := time.Since(start)
			if err == nil {
				trimmed := bytes.Trim(out, "\x00")
				if !bytes.Equal(tt.want, trimmed) {
					t.Errorf("exp resp: %[1]v %[1]s got: %[2]v %[2]s\n", tt.want, trimmed)
				}
			}
			if elapsed >= tt.dur && elapsed < tt.dur+5*time.Microsecond {
				t.Errorf("exp delay around: %v got: %v\n", tt.dur, elapsed)
			}
		})
	}
}

func TestRunWithMismatchStream(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer setupStreamTestCase(t, ADDR4, vdfileStreamMismatch)()
	// connect to server
	conn, err := net.Dial("tcp", ADDR4)
	if err != nil {
		t.Fatalf("could not connect to to server: %v\n", err)
	}
	defer conn.Close()
	// set timeout for reading data
	conn.SetReadDeadline(time.Now().Add(time.Second))

	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"current check", []byte("CUR?\r\n"), []byte("CUR 300\r\n")},
		{"psi check", []byte("PSI?\r\n"), []byte("PSI 3.30\r\n")},
		{"version check", []byte("VER?\r\n"), []byte("version 1.0\r\n")},
		{"mode check", []byte(":PULSE0:MODE?\r\n"), []byte("NORM\r\n")},
		{"temp check", []byte("TEMP?\r\n"), []byte("TEMP 2.30\r\n")},
		{"ack check", []byte("ACK?\r\n"), []byte("false\r\n")},
		{"current set", []byte("CUR 20\r\n"), []byte("OK\r\n")},
		{"psi set", []byte("PSI 3.46\r\n"), []byte("PSI 3.46 OK\r\n")},
		{"mode set", []byte(":PULSE0:MODE SING\r\n"), []byte("ok\r\n")},
		{"wrong parameter", []byte("test"), []byte("Wrong query\r\n")},
		{"only white characters ", []byte("\t"), []byte("Wrong query\r\n")},
		{"wrong set value", []byte("PSI test\r\n"), []byte("Wrong query\r\n")},
		{"wrong mode", []byte(":PULSE0:MODE TEST\r\n"), []byte("Wrong query\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := conn.Write(tt.input); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			out := make([]byte, 128)
			if _, err := conn.Read(out); err == nil {
				trimmed := bytes.Trim(out, "\x00")
				if !bytes.Equal(tt.want, trimmed) {
					t.Errorf("exp resp: %[1]v %[1]s got: %[2]v %[2]s\n", tt.want, trimmed)
				}
			} else {
				t.Error("could not read from connection")
			}
		})
	}
}
