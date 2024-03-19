package memory

import (
	"errors"
)

type DataTyp uint8

const (
	DataCoil DataTyp = iota
	DataDiscreteInput
	DataHoldingRegister
	DataInputRegister
	DataWrongTyp
)

var ErrMemoryCollision = errors.New("Memory collision")

type Memory struct {
	Typ     DataTyp
	Addr    uint16
	Length  uint8 // na potem
	DataTyp string
}

func New(addr uint16, regTyp, dataTyp string) Memory {
	return Memory{
		Addr:    addr,
		Typ:     typConvert(regTyp),
		Length:  getLength(dataTyp),
		DataTyp: dataTyp,
	}
}

// do poprawki bo rejestr ma 16 bitów
func IsMemoryValid(mems map[string]Memory) error {
	// Iterate over the map
	for key1, mem1 := range mems {
		for key2, mem2 := range mems {
			if mem2.Typ == mem1.Typ && key1 != key2 {
				// Check for collision
				if (mem1.Addr <= mem2.Addr && mem1.Addr+uint16(mem1.Length) > mem2.Addr) ||
					(mem2.Addr <= mem1.Addr && mem2.Addr+uint16(mem2.Length) > mem1.Addr) {
					return ErrMemoryCollision
				}
			}
		}
	}
	return nil
}

func getLength(typ string) uint8 {
	switch typ {
	case "int16", "uint16":
		return 2 / 2
	case "int32", "uint32", "float32":
		return 4 / 2
	case "int64", "uint64", "float64":
		return 8 / 2
	default:
		return 0
	}
}

func typConvert(s string) DataTyp {
	switch {
	case s == "coil":
		return DataCoil
	case s == "di":
		return DataDiscreteInput
	case s == "holdreg":
		return DataHoldingRegister
	case s == "inreg":
		return DataInputRegister
	default:
		return DataWrongTyp
	}
}
