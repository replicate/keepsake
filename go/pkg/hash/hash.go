package hash

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letters = "1234567890abcdef"

// Random generates a random hash
//
// intended to be the same as python/keepsake/hash.py
func Random() string {
	n := 64
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
