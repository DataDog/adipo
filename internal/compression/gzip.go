package compression

import (
	"bytes"
	"compress/gzip"
	"io"
)

// GzipCompressor implements gzip compression
type GzipCompressor struct{}

// Compress compresses data using gzip
func (c *GzipCompressor) Compress(input []byte, level int) ([]byte, error) {
	// Default to best compression if level not specified
	if level <= 0 {
		level = gzip.BestCompression
	}
	if level > gzip.BestCompression {
		level = gzip.BestCompression
	}

	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
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

// Decompress decompresses gzip data
func (c *GzipCompressor) Decompress(input []byte, expectedSize uint64) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	// Pre-allocate based on expected size
	output := make([]byte, 0, expectedSize)
	buf := bytes.NewBuffer(output)

	if _, err := io.Copy(buf, reader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Name returns the algorithm name
func (c *GzipCompressor) Name() string {
	return "gzip"
}
