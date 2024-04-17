package modbus

import (
	"fmt"
)

// Exception codes
type Exception uint8

var (
	// Success operation successful.
	Success Exception
	// IllegalFunction function code received in the query is not recognized or allowed by slave.
	IllegalFunction Exception = 1
	// IllegalDataAddress data address of some or all the required entities are not allowed or do not exist in slave.
	IllegalDataAddress Exception = 2
	// IllegalDataValue value is not accepted by slave.
	IllegalDataValue Exception = 3
)

// Method to print exception message
func (e Exception) Error() string {
	return fmt.Sprintf("%d", e)
}

// Stringer method to convert exception to string exception message
func (e Exception) String() string {
	var str string
	switch e {
	case Success:
		str = "Success"
	case IllegalFunction:
		str = "IllegalFunction"
	case IllegalDataAddress:
		str = "IllegalDataAddress"
	case IllegalDataValue:
		str = "IllegalDataValue"
	default:
		str = "unknown"
	}
	return str
}
