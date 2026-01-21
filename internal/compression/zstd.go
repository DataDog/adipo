// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package compression

import (
	"fmt"

	"github.com/DataDog/adipo/internal/format"
	"github.com/klauspost/compress/zstd"
)

// ZstdCompressor implements zstd compression
type ZstdCompressor struct{}

// Compress compresses data using zstd
func (c *ZstdCompressor) Compress(input []byte, level int) ([]byte, error) {
	// Default to level 3 if not specified
	if level <= 0 {
		level = 3
	}
	// zstd levels: 1 (fastest) to 22 (best compression, very slow)
	if level > 22 {
		level = 22
	}

	encoder, err := zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
	if err != nil {
		return nil, err
	}
	defer func() { _ = encoder.Close() }()

	compressed := encoder.EncodeAll(input, make([]byte, 0, len(input)))
	return compressed, nil
}

// Decompress decompresses zstd data
func (c *ZstdCompressor) Decompress(input []byte, expectedSize uint64) ([]byte, error) {
	// Validate expected size to prevent decompression bombs
	if expectedSize > format.MaxOriginalSize {
		return nil, fmt.Errorf("expected decompression size (%d bytes) exceeds maximum allowed (%d bytes)",
			expectedSize, format.MaxOriginalSize)
	}
	if expectedSize == 0 {
		return nil, fmt.Errorf("invalid expected size: 0")
	}

	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	decompressed, err := decoder.DecodeAll(input, make([]byte, 0, expectedSize))
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// Name returns the algorithm name
func (c *ZstdCompressor) Name() string {
	return "zstd"
}
