package compression

import (
	"fmt"

	"github.com/DataDog/adipo/internal/format"
)

// Compressor is the interface for compression algorithms
type Compressor interface {
	// Compress compresses the input data
	Compress(input []byte, level int) ([]byte, error)

	// Decompress decompresses the input data
	Decompress(input []byte, expectedSize uint64) ([]byte, error)

	// Name returns the algorithm name
	Name() string
}

// Get returns a compressor for the given algorithm
func Get(algo format.CompressionAlgo) (Compressor, error) {
	switch algo {
	case format.CompressionNone:
		return &NoneCompressor{}, nil
	case format.CompressionGzip:
		return &GzipCompressor{}, nil
	case format.CompressionZstd:
		return &ZstdCompressor{}, nil
	case format.CompressionLZ4:
		return &LZ4Compressor{}, nil
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %v", algo)
	}
}

// Compress is a convenience function to compress data
func Compress(data []byte, algo format.CompressionAlgo, level int) ([]byte, error) {
	compressor, err := Get(algo)
	if err != nil {
		return nil, err
	}
	return compressor.Compress(data, level)
}

// Decompress is a convenience function to decompress data
func Decompress(data []byte, algo format.CompressionAlgo, expectedSize uint64) ([]byte, error) {
	compressor, err := Get(algo)
	if err != nil {
		return nil, err
	}
	return compressor.Decompress(data, expectedSize)
}
