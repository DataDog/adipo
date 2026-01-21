// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package compression

import (
	"bytes"
	"fmt"
	"io"

	"github.com/DataDog/adipo/internal/format"
	"github.com/pierrec/lz4/v4"
)

// LZ4Compressor implements lz4 compression
type LZ4Compressor struct{}

// Compress compresses data using lz4
func (c *LZ4Compressor) Compress(input []byte, level int) ([]byte, error) {
	// LZ4 compression level (0-16, where 0 is fastest)
	// For compatibility with other algorithms, we'll map level to lz4 level
	// level 1-3 -> lz4 level 0 (fast)
	// level 4-9 -> lz4 level 9 (default)
	// level 10+ -> lz4 level 12 (high compression)
	lz4Level := 9 // default
	if level > 0 && level <= 3 {
		lz4Level = 0
	} else if level > 9 {
		lz4Level = 12
	}

	var buf bytes.Buffer
	writer := lz4.NewWriter(&buf)

	// Set compression level
	if err := writer.Apply(lz4.CompressionLevelOption(lz4.CompressionLevel(lz4Level))); err != nil {
		return nil, err
	}

	if _, err := writer.Write(input); err != nil {
		_ = writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress decompresses lz4 data
func (c *LZ4Compressor) Decompress(input []byte, expectedSize uint64) ([]byte, error) {
	// Validate expected size to prevent decompression bombs
	if expectedSize > format.MaxOriginalSize {
		return nil, fmt.Errorf("expected decompression size (%d bytes) exceeds maximum allowed (%d bytes)",
			expectedSize, format.MaxOriginalSize)
	}
	if expectedSize == 0 {
		return nil, fmt.Errorf("invalid expected size: 0")
	}

	reader := lz4.NewReader(bytes.NewReader(input))

	// Pre-allocate based on expected size
	output := make([]byte, 0, expectedSize)
	buf := bytes.NewBuffer(output)

	if _, err := io.Copy(buf, reader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Name returns the algorithm name
func (c *LZ4Compressor) Name() string {
	return "lz4"
}
