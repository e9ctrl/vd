package memory

type DataTyp uint8

const (
	DataCoil DataTyp = iota
	DataDiscreteInput
	DataHoldingRegister
	DataInputRegister
	DataWrongTyp
)

type Memory struct {
	Typ    DataTyp
	Addr   uint16
	Length uint8 // na potem
}

func New(addr uint16, typ string) Memory {
	return Memory{
		Addr: addr,
		Typ:  typConvert(typ),
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
