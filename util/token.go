package util

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func GenerateRandomToken(length int) string {
	data := make([]byte, base64.RawURLEncoding.DecodedLen(length)+1)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)[:length]
}

func WriteStringToWriter(w io.Writer, s string) error {
	n := 0
	d := []byte(s)
	for n < len(d) {
		t, err := w.Write(d[n:])
		if err != nil {
			return err
		}
		n += t
	}
	return nil
}
