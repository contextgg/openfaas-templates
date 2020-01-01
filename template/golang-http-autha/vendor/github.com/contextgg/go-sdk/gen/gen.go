package gen

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// RandomString will return a randomised string
func RandomString(length int) string {
	nonceBytes := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, nonceBytes)
	if err != nil {
		panic("Source of randomness unavailable: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(nonceBytes)
}
