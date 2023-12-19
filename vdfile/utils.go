package vdfile

import (
	"fmt"
	"math/rand"
)

func GenerateRandomDelay(config Config) Config {
	for i := 0; i < len(config.Commands); i++ {
		config.Commands[i].Dly = fmt.Sprintf("%ds", rand.Intn(10))
	}

	return config
}
