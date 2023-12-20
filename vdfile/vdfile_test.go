package vdfile

import (
	"bytes"
	"testing"
)

func TestParseTerminator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		line string
		exp  []byte
	}{

		{"valid line", "CR LF", []byte{0x0D, 0x0A}},
		{"empty line", "", []byte(nil)},
		{"unknown terminator", "CR TEST", []byte{0x0D, 0x54, 0x45, 0x53, 0x54}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := parseTerminator(tt.line)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp value: %v got %v\n", tt.name, tt.exp, res)
			}
		})
	}
}

func TestGenerateRandomDelayVDFile(t *testing.T) {
	config, err := DecodeVDFile("vdfile")
	if err != nil {
		t.Error(err)
		return
	}

	config = GenerateRandomDelay(config)
	err = WriteVDFile("vdfile_random_delay", config)
	if err != nil {
		t.Error(err)
	}
}

func TestParseDelays(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		line string
		exp  time.Duration
	}{

		{"valid line 5s", "5s", 5 * time.Second},
		{"valid line 1m", "1m", time.Minute},
		{"empty line", "", 0},
		{"wrong format", "5test", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := parseDelays(tt.line)
			if res != tt.exp {
				t.Errorf("%s: exp value: %v got %v\n", tt.name, tt.exp, res)
			}
		})
	}
}
