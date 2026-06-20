// Package config loads server configuration.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tharunn0/blitzdb/pkg/types"
)

// Load reads configuration from environment variables with validation.
func Load() (*types.Config, error) {
	defaultTTLStr := os.Getenv("CACHE_DEFAULT_TTL_MIN")
	comp := os.Getenv("CACHE_COMPRESSION")
	fmt.Println("ttl :", defaultTTLStr, "compresion :", comp)

	var defaultTTL uint64
	var err error

	// Parse defaultTTL if provided
	if defaultTTLStr != "" {
		defaultTTL, err = strconv.ParseUint(defaultTTLStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CACHE_DEFAULT_TTL_MIN: %w", err)
		}
	} else {
		defaultTTL = 60 // Default to 60 minutes
	}

	// Validate compression type
	if comp != "" && comp != "snappy" && comp != "none" {
		return nil, fmt.Errorf("invalid CACHE_COMPRESSION: must be 'snappy' or 'none', got '%s'", comp)
	}

	return &types.Config{
		DefaultTTL:  defaultTTL,
		Compression: comp,
	}, nil
}
