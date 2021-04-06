package util

import (
	"math/rand"
	"strings"
	"time"
)

const (
	runeChars = "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
)

func RandomString(n int) string {
	var output strings.Builder

	rand.Seed(time.Now().Unix())

	for i := 0; i < n; i++ {
		random := rand.Intn(len(runeChars))
		randomChar := runeChars[random]
		output.WriteString(string(randomChar))
	}

	return output.String()
}

