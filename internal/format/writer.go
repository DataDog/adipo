package format

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// BinaryEntry represents a binary to be included in the fat binary
type BinaryEntry struct {
	Data         []byte          // The actual binary data (can be compressed or uncompressed)
	Metadata     *BinaryMetadata // Metadata for this binary
	OriginalData []byte          // Original uncompressed data (for checksum calculation)
}

// Writer creates fat binary files
type Writer struct {
	output    io.ReadWriteSeeker
	stubData  []byte
	entries   []*BinaryEntry
	header    *FormatHeader
	writePos  int64
}

// NewWriter creates a new fat binary writer
func NewWriter(output io.ReadWriteSeeker, stubData []byte) *Writer {
	return &Writer{
		output:   output,
		stubData: stubData,
		entries:  make([]*BinaryEntry, 0),
		header: &FormatHeader{
			Magic:           MagicMarker,
			Version:         FormatVersion,
			CompressionAlgo: CompressionZstd, // default
		},
	}
}

// SetDefaultCompression sets the default compression algorithm
func (w *Writer) SetDefaultCompression(algo CompressionAlgo) {
	w.header.CompressionAlgo = algo
}

// SetStubArchitecture sets the stub architecture information
func (w *Writer) SetStubArchitecture(arch Architecture, version ArchVersion) {
	w.header.StubArchitecture = arch
	w.header.StubArchVersion = version
}

// AddBinary adds a binary to the fat binary
func (w *Writer) AddBinary(entry *BinaryEntry) error {
	// Calculate checksum of original data
	checksum := sha256.Sum256(entry.OriginalData)
	entry.Metadata.Checksum = checksum

	// Set original size
	entry.Metadata.OriginalSize = uint64(len(entry.OriginalData))

	// Set compressed size (actual data size)
	entry.Metadata.CompressedSize = uint64(len(entry.Data))

	w.entries = append(w.entries, entry)
	return nil
}

// Write writes the complete fat binary to the output
func (w *Writer) Write() error {
	// Calculate sizes and offsets
	if err := w.calculateLayout(); err != nil {
		return err
	}

	// Write stub binary
	if _, err := w.output.Write(w.stubData); err != nil {
		return fmt.Errorf("failed to write stub binary: %w", err)
	}
	w.writePos = int64(len(w.stubData))

	// Write magic marker
	if _, err := w.output.Write(MagicMarker[:]); err != nil {
		return fmt.Errorf("failed to write magic marker: %w", err)
	}
	w.writePos += MagicSize

	// Write header (without checksum)
	headerData, err := w.header.MarshalBinary()
	if err != nil {
		return err
	}
	if _, err := w.output.Write(headerData); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	w.writePos += HeaderSize

	// Write metadata table
	for i, entry := range w.entries {
		metaData, err := entry.Metadata.MarshalBinary()
		if err != nil {
			return err
		}
		if _, err := w.output.Write(metaData); err != nil {
			return fmt.Errorf("failed to write metadata entry %d: %w", i, err)
		}
		w.writePos += MetadataEntrySize
	}

	// Write binary data
	for i, entry := range w.entries {
		if _, err := w.output.Write(entry.Data); err != nil {
			return fmt.Errorf("failed to write binary data %d: %w", i, err)
		}
		w.writePos += int64(len(entry.Data))
	}

	// Calculate and write checksum
	if err := w.calculateAndWriteChecksum(); err != nil {
		return err
	}

	return nil
}

// calculateLayout calculates sizes and offsets for all components
func (w *Writer) calculateLayout() error {
	w.header.StubSize = uint64(len(w.stubData))
	w.header.NumBinaries = uint32(len(w.entries))

	// Calculate offsets
	stubSize := uint64(len(w.stubData))
	magicSize := uint64(MagicSize)
	headerSize := uint64(HeaderSize)

	w.header.MetadataOffset = stubSize + magicSize + headerSize
	w.header.MetadataSize = uint64(len(w.entries)) * MetadataEntrySize
	w.header.DataOffset = w.header.MetadataOffset + w.header.MetadataSize

	// Calculate flags
	w.header.Flags = 0
	hasX86 := false
	hasARM := false

	for _, entry := range w.entries {
		switch entry.Metadata.Architecture {
		case ArchX86_64:
			hasX86 = true
		case ArchARM64:
			hasARM = true
		}
	}

	if hasX86 {
		w.header.Flags |= FlagContainsX86_64
	}
	if hasARM {
		w.header.Flags |= FlagContainsARM64
	}
	if hasX86 && hasARM {
		w.header.Flags |= FlagMixedArch
	}

	// Set data offsets for each binary
	currentOffset := w.header.DataOffset
	for _, entry := range w.entries {
		entry.Metadata.DataOffset = currentOffset
		currentOffset += entry.Metadata.CompressedSize
	}

	return nil
}

// calculateAndWriteChecksum calculates SHA-256 of the entire file and updates the header
func (w *Writer) calculateAndWriteChecksum() error {
	// Seek to start of file
	if _, err := w.output.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to file start for checksum calculation: %w", err)
	}

	// Create a hash of everything except the checksum field
	hasher := sha256.New()

	// Read and hash up to checksum field
	stubSize := len(w.stubData)
	checksumFieldOffset := int64(stubSize) + MagicSize + ChecksumOffset

	// Read from start to just before checksum
	buf := make([]byte, checksumFieldOffset)
	if _, err := io.ReadFull(w.output, buf); err != nil {
		return fmt.Errorf("failed to read file data for checksum calculation: %w", err)
	}
	hasher.Write(buf)

	// Skip checksum field
	if _, err := w.output.Seek(ChecksumSize, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek past checksum field: %w", err)
	}

	// Hash the rest of the file
	if _, err := io.Copy(hasher, w.output); err != nil {
		return fmt.Errorf("failed to read remaining file data for checksum: %w", err)
	}

	// Calculate checksum
	checksum := hasher.Sum(nil)

	// Write checksum to header
	if _, err := w.output.Seek(checksumFieldOffset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to checksum field offset: %w", err)
	}
	if _, err := w.output.Write(checksum); err != nil {
		return fmt.Errorf("failed to write checksum: %w", err)
	}

	return nil
}

// WriteToFile is a convenience method to write to a file path
func WriteToFile(path string, stubData []byte, entries []*BinaryEntry, stubArch Architecture, stubArchVer ArchVersion) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := NewWriter(file, stubData)
	writer.SetStubArchitecture(stubArch, stubArchVer)

	for _, entry := range entries {
		if err := writer.AddBinary(entry); err != nil {
			return err
		}
	}

	if err := writer.Write(); err != nil {
		return err
	}

	// Make executable
	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	return nil
}
