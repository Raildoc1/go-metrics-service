package middleware

import (
	"bytes"
	"crypto/rand"
	"go-metrics-service/internal/common/hashing"
	"io"
	"testing"
)

func BenchmarkCalculateHash(b *testing.B) {
	hashFactory := hashing.NewHMAC("private key")
	body := make([]byte, 1_024_000)
	_, err := rand.Read(body)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		_, err := calculateHash(body, hashFactory)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadBodyAndRewind(b *testing.B) {
	b.Run("read and rewind body", func(b *testing.B) {
		byteBuffer := bytes.Buffer{}
		byteBuffer.WriteString("test message")
		var readCloser = io.NopCloser(&byteBuffer)
		b.ResetTimer()
		_, err := readBodyAndRewind(&readCloser)
		b.StopTimer()
		if err != nil {
			b.Fatal(err)
		}
	})
}
