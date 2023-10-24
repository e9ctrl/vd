package main

import (
	"bytes"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/e9ctrl/vd/server"
	"github.com/e9ctrl/vd/stream"
)

const ADDR = "localhost:3333"

func setupTestCase(t *testing.T) func() {
	//read file
	vdfile, err := stream.ReadVDFile("stream/vdfile")
	if err != nil {
		t.Fatal(err)
	}

	//read file
	d, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	s, err := server.New(d, ADDR)
	if err != nil {
		t.Fatalf("error while creating server %v\n", err)
	}

	s.Start()

	return func() {
		s.Stop()
	}
}

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer setupTestCase(t)()
	// connect to server
	conn, err := net.Dial("tcp", ADDR)

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

func TestRunWrongQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer setupTestCase(t)()
	// connect to server
	conn, err := net.Dial("tcp", ADDR)

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
