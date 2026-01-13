// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package compression

// NoneCompressor is a pass-through compressor (no compression)
type NoneCompressor struct{}

// Compress returns the input unchanged
func (c *NoneCompressor) Compress(input []byte, level int) ([]byte, error) {
	return input, nil
}

// Decompress returns the input unchanged
func (c *NoneCompressor) Decompress(input []byte, expectedSize uint64) ([]byte, error) {
	return input, nil
}

// Name returns the algorithm name
func (c *NoneCompressor) Name() string {
	return "none"
}
