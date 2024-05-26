package memory

import (
	"testing"
)

var ()

func TestGetLength(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		typ  string
		want uint8
	}{
		{"test int16", "int16", uint8(1)},
		{"test uint16", "uint16", uint8(1)},
		{"test int32", "int32", uint8(2)},
		{"test uint32", "uint32", uint8(2)},
		{"test float32", "float32", uint8(2)},
		{"test float64", "float64", uint8(4)},
		{"test int64", "int64", uint8(4)},
		{"test uint64", "uint64", uint8(4)},
		{"wrong type", "test", uint8(0)},
		{"empty type", "", uint8(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLength(tt.typ)
			if got != tt.want {
				t.Errorf("exp length: %d got: %d", tt.want, got)
			}
		})
	}
}

func TestIsMemoryValid(t *testing.T) {
	mapOK1 := make(map[string]Memory, 6)
	mapOK1["di1"] = New(1, "di", "uint8")
	mapOK1["di2"] = New(2, "di", "uint8")
	mapOK1["coil1"] = New(1, "coil", "uint8")
	mapOK1["hr1"] = New(1, "holdreg", "uint16")
	mapOK1["hr2"] = New(3, "holdreg", "int64")
	mapOK1["inreg1"] = New(3, "inreg", "int64")

	mapWrong1 := make(map[string]Memory, 2)
	mapWrong1["di1"] = New(1, "di", "uint8")
	mapWrong1["di2"] = New(1, "di", "uint8")

	mapWrong2 := make(map[string]Memory, 2)
	mapWrong2["hr1"] = New(1, "holdreg", "int64")
	mapWrong2["hr2"] = New(3, "holdreg", "uint32")

	t.Parallel()
	tests := []struct {
		name  string
		input map[string]Memory
		want  error
	}{
		{"ok memory map", mapOK1, nil},
		{"wromg memory map - same addresses", mapWrong1, ErrMemoryCollision},
		{"wromg memory map - covering ranges", mapWrong2, ErrMemoryCollision},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMemoryValid(tt.input)
			if got != tt.want {
				t.Errorf("exp error: %v got: %v", tt.want, got)
			}
		})
	}
}

func TestTypConvert(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		typ  string
		want DataTyp
	}{
		{"test coil", "coil", DataCoil},
		{"test di", "di", DataDiscreteInput},
		{"test holdreg", "holdreg", DataHoldingRegister},
		{"test inreg", "inreg", DataInputRegister},
		{"wrong type", "test", DataWrongTyp},
		{"empty type", "", DataWrongTyp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typConvert(tt.typ)
			if got != tt.want {
				t.Errorf("exp length: %d got: %d", tt.want, got)
			}
		})
	}
}
