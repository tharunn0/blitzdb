// Package types defines shared types for the cache server.
package types

type Entry struct {
	Value     []byte
	ExpireTick uint64
}

type Config struct {
	DefaultTTL  uint64 // minutes
	Compression string // "gzip" or "snappy"
}