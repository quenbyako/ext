package tensor

import (
	"crypto/rand"
	"encoding/binary"
)

func GetRandInt(min, max int) int {
	var buf [8]byte
	rand.Read(buf[:])

	n := binary.LittleEndian.Uint64(buf[:]) % uint64(max-min)
	return int(n) + min
}

func GetRandFloat[float float32 | float64](min, max float, precision int) float {
	minInt := int(min) * precision
	maxInt := int(max) * precision

	return float(GetRandInt(minInt, maxInt)) / float(precision)
}
