// Package middleware provides reusable HTTP middleware for the chi router.
// Add new middleware as separate files in this package (e.g., auth.go).
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
)

// maxDecompressedSize is the upper bound on decompressed request body size.
// This prevents gzip bombs from exhausting server memory.
const maxDecompressedSize = 10 << 20 // 10 MB

// Decompress transparently decompresses request bodies that arrive with
// Content-Encoding: gzip. The decompressed stream is capped at
// maxDecompressedSize to guard against decompression bombs. After unwrapping,
// it removes the header so downstream handlers see a plain body.
func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = io.NopCloser(io.LimitReader(gz, maxDecompressedSize))
			r.Header.Del("Content-Encoding")
		}
		next.ServeHTTP(w, r)
	})
}
