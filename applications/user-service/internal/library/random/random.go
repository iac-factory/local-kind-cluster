package random

import (
	"math/rand"
	"time"
)

const set = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var total = len(set)

type random = rand.Rand

// entropy - random Instance Constructor
func entropy() *random {
	// Time-Generated Source
	src := rand.NewSource(time.Now().UnixNano())

	// Randomized Instance
	return rand.New(src)
}

func generator(length int) string {
	// Randomized Entropic Seed
	seed := entropy()

	// Pre-Allocated Bytes-Buffer
	buffer := make([]byte, length)

	for character := range buffer {
		// Random Selection from Set
		buffer[character] = set[seed.Intn(total)]
	}

	// String Typecast
	return string(buffer)
}

// Verification - random token generator for generated user account verification tokens [http-api/internal/database/models.Verification]
func Verification() string {
	return generator(32)
}
