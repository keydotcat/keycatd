package util

import (
	"bytes"
	"testing"
)

func TestGenerateRandomToken(t *testing.T) {
	tok := GenerateRandomToken(64)
	if len(tok) != 64 {
		t.Error("Token doe not have expected length")
	}
	if GenerateRandomToken(64) == tok {
		t.Error("Tokens are the same!!")
	}
}

func TestWriteStringtoWriter(t *testing.T) {
	source := GenerateRandomToken(128)
	b := bytes.NewBuffer(nil)
	if err := WriteStringToWriter(b, source); err != nil {
		t.Fatal(err)
	}
	if b.String() != source {
		t.Fatalf("Unexpected data %s vs %s", b.String(), source)
	}
}
