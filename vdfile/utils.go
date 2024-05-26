package vdfile

import (
	"fmt"
	"math/rand"
)

// Adds random delays to commands
func GenerateRandomDelay(vd VDFile) VDFile {
	for i := 0; i < len(vd.Commands); i++ {
		vd.Commands[i].Dly = fmt.Sprintf("%ds", rand.Intn(10))
	}

	return vd
}
