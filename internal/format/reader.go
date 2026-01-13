// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package format

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Reader reads fat binary files
type Reader struct {
	input    io.ReadSeeker
	fileSize int64
	header   *FormatHeader
	metadata []*BinaryMetadata
	closer   io.Closer // Optional closer for file handles
}

// NewReader creates a new fat binary reader
func NewReader(input io.ReadSeeker) (*Reader, error) {
	// Get file size
	size, err := input.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to end of file: %w", err)
	}

	reader := &Reader{
		input:    input,
		fileSize: size,
	}

	// Parse the fat binary
	if err := reader.parse(); err != nil {
		return nil, err
	}

	return reader, nil
}

// parse parses the fat binary format
func (r *Reader) parse() error {
	// Find magic marker
	magicOffset, err := r.findMagicMarker()
	if err != nil {
		return err
	}

	// Read header (after standalone magic marker)
	headerOffset := magicOffset + MagicSize
	if _, err := r.input.Seek(headerOffset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to header offset %d: %w", headerOffset, err)
	}

	headerData := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r.input, headerData); err != nil {
		return fmt.Errorf("failed to read header data: %w", err)
	}

	r.header = &FormatHeader{}
	if err := r.header.UnmarshalBinary(headerData); err != nil {
		return err
	}

	// Validate magic
	if r.header.Magic != MagicMarker {
		return ErrInvalidMagic
	}

	// Validate version
	if r.header.Version != FormatVersion {
		return ErrUnsupportedVersion
	}

	// Validate stub size to detect corrupted files
	const minStubSize = 100 * 1024      // 100 KB minimum
	const maxStubSize = 100 * 1024 * 1024 // 100 MB maximum
	if r.header.StubSize < minStubSize {
		return fmt.Errorf("stub size (%d bytes) is too small (minimum: %d bytes)",
			r.header.StubSize, minStubSize)
	}
	if r.header.StubSize > maxStubSize {
		return fmt.Errorf("stub size (%d bytes) exceeds maximum allowed (%d bytes)",
			r.header.StubSize, maxStubSize)
	}

	// Validate number of binaries to prevent excessive memory allocation
	if r.header.NumBinaries == 0 {
		return fmt.Errorf("invalid number of binaries: 0")
	}
	if r.header.NumBinaries > MaxNumBinaries {
		return fmt.Errorf("number of binaries (%d) exceeds maximum allowed (%d)",
			r.header.NumBinaries, MaxNumBinaries)
	}

	// Read metadata table
	if _, err := r.input.Seek(int64(r.header.MetadataOffset), io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to metadata offset %d: %w", r.header.MetadataOffset, err)
	}

	r.metadata = make([]*BinaryMetadata, r.header.NumBinaries)
	for i := uint32(0); i < r.header.NumBinaries; i++ {
		metaData := make([]byte, MetadataEntrySize)
		if _, err := io.ReadFull(r.input, metaData); err != nil {
			return fmt.Errorf("failed to read metadata entry %d: %w", i, err)
		}

		r.metadata[i] = &BinaryMetadata{}
		if err := r.metadata[i].UnmarshalBinary(metaData); err != nil {
			return err
		}

		// Validate sizes to prevent integer overflow and decompression bombs
		meta := r.metadata[i]
		if meta.CompressedSize > MaxCompressedSize {
			return fmt.Errorf("binary %d: compressed size (%d bytes) exceeds maximum allowed (%d bytes)",
				i, meta.CompressedSize, MaxCompressedSize)
		}
		if meta.OriginalSize > MaxOriginalSize {
			return fmt.Errorf("binary %d: original size (%d bytes) exceeds maximum allowed (%d bytes)",
				i, meta.OriginalSize, MaxOriginalSize)
		}
		if meta.CompressedSize == 0 {
			return fmt.Errorf("binary %d: invalid compressed size: 0", i)
		}
		if meta.OriginalSize == 0 {
			return fmt.Errorf("binary %d: invalid original size: 0", i)
		}
	}

	return nil
}

// validateMagicMarker checks if the magic marker at the given offset is valid
// by verifying the header that follows it. This reduces false positives when
// the magic bytes appear in legitimate data (e.g., embedded resources).
func (r *Reader) validateMagicMarker(offset int64) bool {
	// Seek to position after standalone magic marker (at start of header)
	if _, err := r.input.Seek(offset+MagicSize, io.SeekStart); err != nil {
		return false
	}

	// Read enough bytes to check critical header fields
	bytesNeeded := MagicSize + VersionSize + NumBinariesSize + StubSizeSize +
		MetadataOffsetSize + MetadataSizeSize
	buf := make([]byte, bytesNeeded)
	if _, err := io.ReadFull(r.input, buf); err != nil {
		return false
	}

	// Check that header also starts with magic marker (double verification)
	if !bytes.Equal(buf[MagicOffset:MagicOffset+MagicSize], MagicMarker[:]) {
		return false
	}

	// Check version (should be 1)
	version := binary.LittleEndian.Uint32(buf[VersionOffset:VersionOffset+VersionSize])
	if version != FormatVersion {
		return false
	}

	// Check NumBinaries (should be > 0 and within limits)
	numBinaries := binary.LittleEndian.Uint32(buf[NumBinariesOffset:NumBinariesOffset+NumBinariesSize])
	if numBinaries == 0 || numBinaries > uint32(MaxNumBinaries) {
		return false
	}

	// Check StubSize matches offset (critical validation for uniqueness)
	stubSize := binary.LittleEndian.Uint64(buf[StubSizeOffset:StubSizeOffset+StubSizeSize])
	if int64(stubSize) != offset {
		return false
	}

	// Validate stub size is within reasonable bounds
	const minStubSize = 100 * 1024
	const maxStubSize = 100 * 1024 * 1024
	if stubSize < minStubSize || stubSize > maxStubSize {
		return false
	}

	// Check MetadataOffset and MetadataSize are sensible
	metadataOffset := binary.LittleEndian.Uint64(buf[MetadataOffsetOffset:MetadataOffsetOffset+MetadataOffsetSize])
	metadataSize := binary.LittleEndian.Uint64(buf[MetadataSizeOffset:MetadataSizeOffset+MetadataSizeSize])

	// Metadata should come after stub + header
	expectedMinMetadataOffset := uint64(offset) + uint64(MagicSize) + uint64(HeaderSize)
	if metadataOffset < expectedMinMetadataOffset {
		return false
	}

	// Metadata size should match number of binaries
	expectedMetadataSize := uint64(numBinaries) * uint64(MetadataEntrySize)
	if metadataSize != expectedMetadataSize {
		return false
	}

	// All checks passed - this is likely the real magic marker
	return true
}

// findMagicMarker searches for the magic marker in the file
// It searches backwards from the header location
func (r *Reader) findMagicMarker() (int64, error) {
	// The magic marker should be right after the stub binary
	// We need to scan through the file to find it
	// Strategy: scan in chunks from different positions

	// Try common stub sizes first (2-4 MB typical for Go binaries)
	commonOffsets := []int64{
		2 * 1024 * 1024,     // 2 MB
		2.5 * 1024 * 1024,   // 2.5 MB
		3 * 1024 * 1024,     // 3 MB
		4 * 1024 * 1024,     // 4 MB
		5 * 1024 * 1024,     // 5 MB
		1 * 1024 * 1024,     // 1 MB
		10 * 1024 * 1024,    // 10 MB
	}

	for _, offset := range commonOffsets {
		if offset >= r.fileSize {
			continue
		}

		// Read a chunk around this offset (larger window for better coverage)
		searchStart := offset - 512*1024 // 512 KB before
		if searchStart < 0 {
			searchStart = 0
		}
		searchEnd := offset + 512*1024 // 512 KB after
		if searchEnd > r.fileSize {
			searchEnd = r.fileSize
		}

		if _, err := r.input.Seek(searchStart, io.SeekStart); err != nil {
			continue
		}

		chunk := make([]byte, searchEnd-searchStart)
		if _, err := io.ReadFull(r.input, chunk); err != nil {
			continue
		}

		// Search for magic marker in chunk and validate each occurrence
		searchOffset := 0
		for {
			idx := bytes.Index(chunk[searchOffset:], MagicMarker[:])
			if idx == -1 {
				break
			}

			magicOffset := searchStart + int64(searchOffset) + int64(idx)
			if r.validateMagicMarker(magicOffset) {
				return magicOffset, nil
			}

			searchOffset += idx + 1
		}
	}

	// Fall back to scanning the entire file if not found
	// This is slower but more reliable
	return r.scanForMagic()
}

// scanForMagic scans the entire file for the magic marker
func (r *Reader) scanForMagic() (int64, error) {
	const chunkSize = 64 * 1024 // 64 KB chunks
	overlap := len(MagicMarker) - 1

	if _, err := r.input.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	buffer := make([]byte, chunkSize+overlap)
	offset := int64(0)
	prevChunk := make([]byte, 0, overlap)

	for {
		n, err := r.input.Read(buffer[len(prevChunk):])
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}

		// Combine with previous chunk overlap
		searchBuf := buffer[:len(prevChunk)+n]

		// Search for magic and validate each occurrence
		searchOffset := 0
		for {
			idx := bytes.Index(searchBuf[searchOffset:], MagicMarker[:])
			if idx == -1 {
				break
			}

			magicOffset := offset + int64(searchOffset) + int64(idx) - int64(len(prevChunk))
			if r.validateMagicMarker(magicOffset) {
				return magicOffset, nil
			}

			searchOffset += idx + 1
		}

		// Prepare for next iteration
		if len(searchBuf) >= overlap {
			copy(prevChunk, searchBuf[len(searchBuf)-overlap:])
			prevChunk = prevChunk[:overlap]
		}
		offset += int64(n)

		if err == io.EOF {
			break
		}
	}

	return 0, ErrInvalidMagic
}

// Header returns the format header
func (r *Reader) Header() *FormatHeader {
	return r.header
}

// Metadata returns all binary metadata entries
func (r *Reader) Metadata() []*BinaryMetadata {
	return r.metadata
}

// Close closes the reader and any associated resources
func (r *Reader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

// GetBinaryData reads and returns the compressed data for a specific binary
func (r *Reader) GetBinaryData(index int) ([]byte, error) {
	if index < 0 || index >= len(r.metadata) {
		return nil, ErrInvalidMetadata
	}

	meta := r.metadata[index]

	// Double-check size limits before allocation (defense in depth)
	if meta.CompressedSize > MaxCompressedSize {
		return nil, fmt.Errorf("compressed size (%d bytes) exceeds maximum allowed (%d bytes)",
			meta.CompressedSize, MaxCompressedSize)
	}
	if meta.CompressedSize == 0 {
		return nil, fmt.Errorf("invalid compressed size: 0")
	}

	// Validate file offset to prevent reading beyond file boundaries
	if meta.DataOffset >= uint64(r.fileSize) {
		return nil, fmt.Errorf("data offset (%d) is beyond file size (%d)",
			meta.DataOffset, r.fileSize)
	}

	// Check that offset + size doesn't exceed file size (with overflow protection)
	endOffset := meta.DataOffset + meta.CompressedSize
	if endOffset < meta.DataOffset { // Integer overflow check
		return nil, fmt.Errorf("data offset + compressed size overflows")
	}
	if endOffset > uint64(r.fileSize) {
		return nil, fmt.Errorf("data range [%d, %d) exceeds file size (%d)",
			meta.DataOffset, endOffset, r.fileSize)
	}

	// Seek to binary data
	if _, err := r.input.Seek(int64(meta.DataOffset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to binary data offset %d: %w", meta.DataOffset, err)
	}

	// Read compressed data
	data := make([]byte, meta.CompressedSize)
	if _, err := io.ReadFull(r.input, data); err != nil {
		return nil, fmt.Errorf("failed to read compressed binary data: %w", err)
	}

	return data, nil
}

// VerifyChecksum verifies the integrity of the fat binary
func (r *Reader) VerifyChecksum() error {
	// Seek to start
	if _, err := r.input.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Find magic marker offset
	magicOffset, err := r.findMagicMarker()
	if err != nil {
		return err
	}

	// Calculate checksum of everything except the checksum field
	hasher := sha256.New()

	// Hash up to checksum field (standalone magic + header up to checksum field)
	checksumFieldOffset := magicOffset + MagicSize + ChecksumOffset
	if _, err := r.input.Seek(0, io.SeekStart); err != nil {
		return err
	}

	buf := make([]byte, checksumFieldOffset)
	if _, err := io.ReadFull(r.input, buf); err != nil {
		return err
	}
	hasher.Write(buf)

	// Skip checksum field
	if _, err := r.input.Seek(ChecksumSize, io.SeekCurrent); err != nil {
		return err
	}

	// Hash rest of file
	if _, err := io.Copy(hasher, r.input); err != nil {
		return err
	}

	// Compare checksums
	calculated := hasher.Sum(nil)
	var expected [32]byte
	copy(expected[:], calculated)

	if expected != r.header.Checksum {
		return ErrInvalidChecksum
	}

	return nil
}

// OpenFile is a convenience method to open and parse a fat binary file
func OpenFile(path string) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader, err := NewReader(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	reader.closer = file
	return reader, nil
}
