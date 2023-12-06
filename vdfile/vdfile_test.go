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
