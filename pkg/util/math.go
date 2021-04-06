package util

import "math/rand"

func Between(min, max int) int {
	return rand.Intn(max-min) + min
}
