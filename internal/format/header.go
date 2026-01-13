// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package format

import (
	"encoding/binary"
	"errors"
)

// Magic marker for adipo fat binaries
var MagicMarker = [8]byte{'A', 'D', 'I', 'P', 'O', 'F', 'A', 'T'}

const (
	// FormatVersion is the current format version
	FormatVersion = 1

	// Header field sizes (derived from FormatHeader struct)
	MagicSize              = 8   // [8]byte
	VersionSize            = 4   // uint32
	NumBinariesSize        = 4   // uint32
	StubSizeSize           = 8   // uint64
	MetadataOffsetSize     = 8   // uint64
	MetadataSizeSize       = 8   // uint64
	DataOffsetSize         = 8   // uint64
	FlagsSize              = 8   // uint64
	CompressionAlgoSize    = 4   // uint32
	StubArchitectureSize   = 4   // uint32
	StubArchVersionSize    = 4   // uint32
	ReservedSize           = 924 // [924]byte
	ChecksumSize           = 32  // [32]byte

	// Header field offsets (cumulative)
	MagicOffset              = 0
	VersionOffset            = MagicOffset + MagicSize                                                 // 8
	NumBinariesOffset        = VersionOffset + VersionSize                                             // 12
	StubSizeOffset           = NumBinariesOffset + NumBinariesSize                                     // 16
	MetadataOffsetOffset     = StubSizeOffset + StubSizeSize                                           // 24
	MetadataSizeOffset       = MetadataOffsetOffset + MetadataOffsetSize                               // 32
	DataOffsetOffset         = MetadataSizeOffset + MetadataSizeSize                                   // 40
	FlagsOffset              = DataOffsetOffset + DataOffsetSize                                       // 48
	CompressionAlgoOffset    = FlagsOffset + FlagsSize                                                 // 56
	StubArchitectureOffset   = CompressionAlgoOffset + CompressionAlgoSize                             // 60
	StubArchVersionOffset    = StubArchitectureOffset + StubArchitectureSize                           // 64
	ReservedOffset           = StubArchVersionOffset + StubArchVersionSize                             // 68
	ChecksumOffset           = ReservedOffset + ReservedSize                                           // 992

	// HeaderSize is the fixed size of the format header (derived from all fields)
	HeaderSize = MagicSize + VersionSize + NumBinariesSize + StubSizeSize +
		MetadataOffsetSize + MetadataSizeSize + DataOffsetSize +
		FlagsSize + CompressionAlgoSize + StubArchitectureSize + StubArchVersionSize +
		ReservedSize + ChecksumSize // 1024

	// MetadataEntrySize is the fixed size of each binary metadata entry
	MetadataEntrySize = 512

	// Safety limits to prevent integer overflow and decompression bombs
	MaxNumBinaries    = 10000                // Maximum number of binaries in a fat binary
	MaxCompressedSize = 10 * 1024 * 1024 * 1024 // 10 GB maximum compressed size per binary
	MaxOriginalSize   = 10 * 1024 * 1024 * 1024 // 10 GB maximum uncompressed size per binary
)

// Architecture represents the CPU architecture
type Architecture uint32

const (
	ArchUnknown Architecture = 0
	ArchX86_64  Architecture = 1
	ArchARM64   Architecture = 2
)

func (a Architecture) String() string {
	switch a {
	case ArchX86_64:
		return "x86-64"
	case ArchARM64:
		return "aarch64"
	default:
		return "unknown"
	}
}

// ArchVersion represents the architecture version/level
type ArchVersion uint32

// x86-64 versions
const (
	X86_64_Unknown ArchVersion = 0
	X86_64_V1      ArchVersion = 1
	X86_64_V2      ArchVersion = 2
	X86_64_V3      ArchVersion = 3
	X86_64_V4      ArchVersion = 4
)

// ARM64 versions
const (
	ARM64_Unknown ArchVersion = 0
	ARM64_V8_0    ArchVersion = 1
	ARM64_V8_1    ArchVersion = 2
	ARM64_V8_2    ArchVersion = 3
	ARM64_V8_3    ArchVersion = 4
	ARM64_V8_4    ArchVersion = 5
	ARM64_V8_5    ArchVersion = 8
	ARM64_V8_6    ArchVersion = 9
	ARM64_V8_7    ArchVersion = 10
	ARM64_V8_8    ArchVersion = 11
	ARM64_V8_9    ArchVersion = 12
	ARM64_V9_0    ArchVersion = 6
	ARM64_V9_1    ArchVersion = 7
	ARM64_V9_2    ArchVersion = 13
	ARM64_V9_3    ArchVersion = 14
	ARM64_V9_4    ArchVersion = 15
	ARM64_V9_5    ArchVersion = 16
)

// ARM64VersionFallbackOrder defines the canonical ordering of ARM64 versions from newest to oldest.
// This is used for version fallback chains (e.g., v9.4 can fall back to v9.0, v8.9, etc.).
// IMPORTANT: When adding new ARM64 versions (e.g., ARM64_V9_6), add them to this list in the correct position.
var ARM64VersionFallbackOrder = []ArchVersion{
	ARM64_V9_5,
	ARM64_V9_4,
	ARM64_V9_3,
	ARM64_V9_2,
	ARM64_V9_1,
	ARM64_V9_0,
	ARM64_V8_9,
	ARM64_V8_8,
	ARM64_V8_7,
	ARM64_V8_6,
	ARM64_V8_5,
	ARM64_V8_4,
	ARM64_V8_3,
	ARM64_V8_2,
	ARM64_V8_1,
	ARM64_V8_0,
}

// String returns the version string for the given architecture.
// IMPORTANT: When adding new ARM64 versions, update both ARM64VersionFallbackOrder and this String() method.
func (v ArchVersion) String(arch Architecture) string {
	switch arch {
	case ArchX86_64:
		switch v {
		case X86_64_V1:
			return "v1"
		case X86_64_V2:
			return "v2"
		case X86_64_V3:
			return "v3"
		case X86_64_V4:
			return "v4"
		default:
			return "unknown"
		}
	case ArchARM64:
		switch v {
		case ARM64_V8_0:
			return "v8.0"
		case ARM64_V8_1:
			return "v8.1"
		case ARM64_V8_2:
			return "v8.2"
		case ARM64_V8_3:
			return "v8.3"
		case ARM64_V8_4:
			return "v8.4"
		case ARM64_V8_5:
			return "v8.5"
		case ARM64_V8_6:
			return "v8.6"
		case ARM64_V8_7:
			return "v8.7"
		case ARM64_V8_8:
			return "v8.8"
		case ARM64_V8_9:
			return "v8.9"
		case ARM64_V9_0:
			return "v9.0"
		case ARM64_V9_1:
			return "v9.1"
		case ARM64_V9_2:
			return "v9.2"
		case ARM64_V9_3:
			return "v9.3"
		case ARM64_V9_4:
			return "v9.4"
		case ARM64_V9_5:
			return "v9.5"
		default:
			return "unknown"
		}
	default:
		return "unknown"
	}
}

// CompressionAlgo represents the compression algorithm
type CompressionAlgo uint32

const (
	CompressionNone CompressionAlgo = 0
	CompressionGzip CompressionAlgo = 1
	CompressionZstd CompressionAlgo = 2
	CompressionLZ4  CompressionAlgo = 3
)

func (c CompressionAlgo) String() string {
	switch c {
	case CompressionNone:
		return "none"
	case CompressionGzip:
		return "gzip"
	case CompressionZstd:
		return "zstd"
	case CompressionLZ4:
		return "lz4"
	default:
		return "unknown"
	}
}

// BinaryFormat represents the binary file format
type BinaryFormat uint32

const (
	FormatUnknown BinaryFormat = 0
	FormatELF     BinaryFormat = 1
	FormatMachO   BinaryFormat = 2
	FormatPE      BinaryFormat = 3
)

func (f BinaryFormat) String() string {
	switch f {
	case FormatELF:
		return "ELF"
	case FormatMachO:
		return "Mach-O"
	case FormatPE:
		return "PE"
	default:
		return "unknown"
	}
}

// ValidateBinaryFormat checks if the binary data matches the claimed format
// by examining the magic bytes at the start of the binary.
func ValidateBinaryFormat(data []byte, expectedFormat BinaryFormat) error {
	if len(data) < 4 {
		return errors.New("binary too small to determine format")
	}

	// Check ELF magic: 0x7f 'E' 'L' 'F'
	isELF := len(data) >= 4 && data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F'

	// Check Mach-O magic: 0xfeedface (32-bit), 0xfeedfacf (64-bit), or 0xcafebabe (universal)
	isMachO := false
	if len(data) >= 4 {
		magic := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		isMachO = magic == 0xfeedface || magic == 0xfeedfacf || magic == 0xcafebabe ||
			magic == 0xcefaedfe || magic == 0xcffaedfe // Reverse byte order variants
	}

	// Check PE magic: 'M' 'Z' at start, then PE signature at offset specified by header
	isPE := len(data) >= 2 && data[0] == 'M' && data[1] == 'Z'

	// Validate format matches
	switch expectedFormat {
	case FormatELF:
		if !isELF {
			return errors.New("binary format mismatch: expected ELF but binary has different magic bytes")
		}
	case FormatMachO:
		if !isMachO {
			return errors.New("binary format mismatch: expected Mach-O but binary has different magic bytes")
		}
	case FormatPE:
		if !isPE {
			return errors.New("binary format mismatch: expected PE but binary has different magic bytes")
		}
	case FormatUnknown:
		// Unknown format is not an error, but we can't validate it
		return nil
	default:
		return errors.New("invalid binary format enum value")
	}

	return nil
}

// FormatFlags represents feature flags in the header
type FormatFlags uint64

const (
	FlagContainsX86_64  FormatFlags = 1 << 0
	FlagContainsARM64   FormatFlags = 1 << 1
	FlagMixedArch       FormatFlags = 1 << 2
	FlagStubCompressed  FormatFlags = 1 << 3
)

// StubSettings represents stub execution settings
type StubSettings uint32

const (
	StubSettingVerbose       StubSettings = 1 << 0
	StubSettingCleanupOnExit StubSettings = 1 << 1
	StubSettingPreferDisk    StubSettings = 1 << 2 // Prefer disk extraction over memfd
)

// FormatHeader is the fixed-size header at the start of the fat binary format
// Size: 1024 bytes (1KB)
type FormatHeader struct {
	Magic            [8]byte         // "ADIPOFAT"
	Version          uint32          // Format version (currently 1)
	NumBinaries      uint32          // Number of embedded binaries
	StubSize         uint64          // Size of the stub binary (0 if no stub)
	MetadataOffset   uint64          // Offset to metadata table from start of file
	MetadataSize     uint64          // Size of metadata table in bytes
	DataOffset       uint64          // Offset to first binary data from start of file
	Flags            FormatFlags     // Feature flags
	CompressionAlgo  CompressionAlgo // Default compression algorithm
	StubArchitecture Architecture    // Stub binary architecture (0 if no stub)
	StubArchVersion  ArchVersion     // Stub binary architecture version (0 if no stub)
	Reserved         [924]byte       // Reserved for future use
	Checksum         [32]byte        // SHA-256 of entire fat binary (excluding this field)
}

// MarshalBinary encodes the header to binary format
func (h *FormatHeader) MarshalBinary() ([]byte, error) {
	buf := make([]byte, HeaderSize)

	// Magic
	copy(buf[MagicOffset:MagicOffset+MagicSize], h.Magic[:])

	// Version
	binary.LittleEndian.PutUint32(buf[VersionOffset:VersionOffset+VersionSize], h.Version)

	// NumBinaries
	binary.LittleEndian.PutUint32(buf[NumBinariesOffset:NumBinariesOffset+NumBinariesSize], h.NumBinaries)

	// StubSize
	binary.LittleEndian.PutUint64(buf[StubSizeOffset:StubSizeOffset+StubSizeSize], h.StubSize)

	// MetadataOffset
	binary.LittleEndian.PutUint64(buf[MetadataOffsetOffset:MetadataOffsetOffset+MetadataOffsetSize], h.MetadataOffset)

	// MetadataSize
	binary.LittleEndian.PutUint64(buf[MetadataSizeOffset:MetadataSizeOffset+MetadataSizeSize], h.MetadataSize)

	// DataOffset
	binary.LittleEndian.PutUint64(buf[DataOffsetOffset:DataOffsetOffset+DataOffsetSize], h.DataOffset)

	// Flags
	binary.LittleEndian.PutUint64(buf[FlagsOffset:FlagsOffset+FlagsSize], uint64(h.Flags))

	// CompressionAlgo
	binary.LittleEndian.PutUint32(buf[CompressionAlgoOffset:CompressionAlgoOffset+CompressionAlgoSize], uint32(h.CompressionAlgo))

	// StubArchitecture
	binary.LittleEndian.PutUint32(buf[StubArchitectureOffset:StubArchitectureOffset+StubArchitectureSize], uint32(h.StubArchitecture))

	// StubArchVersion
	binary.LittleEndian.PutUint32(buf[StubArchVersionOffset:StubArchVersionOffset+StubArchVersionSize], uint32(h.StubArchVersion))

	// Reserved
	copy(buf[ReservedOffset:ReservedOffset+ReservedSize], h.Reserved[:])

	// Checksum
	copy(buf[ChecksumOffset:ChecksumOffset+ChecksumSize], h.Checksum[:])

	return buf, nil
}

// UnmarshalBinary decodes the header from binary format
func (h *FormatHeader) UnmarshalBinary(data []byte) error {
	if len(data) < HeaderSize {
		return errors.New("insufficient data for header")
	}

	// Magic
	copy(h.Magic[:], data[MagicOffset:MagicOffset+MagicSize])

	// Version
	h.Version = binary.LittleEndian.Uint32(data[VersionOffset : VersionOffset+VersionSize])

	// NumBinaries
	h.NumBinaries = binary.LittleEndian.Uint32(data[NumBinariesOffset : NumBinariesOffset+NumBinariesSize])

	// StubSize
	h.StubSize = binary.LittleEndian.Uint64(data[StubSizeOffset : StubSizeOffset+StubSizeSize])

	// MetadataOffset
	h.MetadataOffset = binary.LittleEndian.Uint64(data[MetadataOffsetOffset : MetadataOffsetOffset+MetadataOffsetSize])

	// MetadataSize
	h.MetadataSize = binary.LittleEndian.Uint64(data[MetadataSizeOffset : MetadataSizeOffset+MetadataSizeSize])

	// DataOffset
	h.DataOffset = binary.LittleEndian.Uint64(data[DataOffsetOffset : DataOffsetOffset+DataOffsetSize])

	// Flags
	h.Flags = FormatFlags(binary.LittleEndian.Uint64(data[FlagsOffset : FlagsOffset+FlagsSize]))

	// CompressionAlgo
	h.CompressionAlgo = CompressionAlgo(binary.LittleEndian.Uint32(data[CompressionAlgoOffset : CompressionAlgoOffset+CompressionAlgoSize]))

	// StubArchitecture
	h.StubArchitecture = Architecture(binary.LittleEndian.Uint32(data[StubArchitectureOffset : StubArchitectureOffset+StubArchitectureSize]))

	// StubArchVersion
	h.StubArchVersion = ArchVersion(binary.LittleEndian.Uint32(data[StubArchVersionOffset : StubArchVersionOffset+StubArchVersionSize]))

	// Reserved
	copy(h.Reserved[:], data[ReservedOffset:ReservedOffset+ReservedSize])

	// Checksum
	copy(h.Checksum[:], data[ChecksumOffset:ChecksumOffset+ChecksumSize])

	return nil
}

// GetStubSettings returns the stub settings from the reserved space
func (h *FormatHeader) GetStubSettings() StubSettings {
	return StubSettings(binary.LittleEndian.Uint32(h.Reserved[0:4]))
}

// SetStubSettings sets the stub settings in the reserved space
func (h *FormatHeader) SetStubSettings(settings StubSettings) {
	binary.LittleEndian.PutUint32(h.Reserved[0:4], uint32(settings))
}

// GetDefaultExtractDir returns the default extraction directory from the reserved space
func (h *FormatHeader) GetDefaultExtractDir() string {
	// Extract null-terminated string from Reserved[4:516] (512 bytes)
	end := 4
	for i := 4; i < 516 && h.Reserved[i] != 0; i++ {
		end = i + 1
	}
	return string(h.Reserved[4:end])
}

// SetDefaultExtractDir sets the default extraction directory in the reserved space
func (h *FormatHeader) SetDefaultExtractDir(dir string) error {
	// Clear the extraction dir area
	for i := 4; i < 516; i++ {
		h.Reserved[i] = 0
	}

	// Check if the directory path fits (512 bytes including null terminator)
	if len(dir) > 511 {
		return errors.New("extraction directory path too long (max 511 bytes)")
	}

	// Copy the directory path
	copy(h.Reserved[4:], []byte(dir))
	return nil
}

// GetDefaultExtractFile returns the default extraction file template from the reserved space
func (h *FormatHeader) GetDefaultExtractFile() string {
	// Extract null-terminated string from Reserved[516:772] (256 bytes)
	end := 516
	for i := 516; i < 772 && h.Reserved[i] != 0; i++ {
		end = i + 1
	}
	return string(h.Reserved[516:end])
}

// SetDefaultExtractFile sets the default extraction file template in the reserved space
func (h *FormatHeader) SetDefaultExtractFile(file string) error {
	// Clear the extraction file area
	for i := 516; i < 772; i++ {
		h.Reserved[i] = 0
	}

	// Check if the file template fits (256 bytes including null terminator)
	if len(file) > 255 {
		return errors.New("extraction file template too long (max 255 bytes)")
	}

	// Copy the file template
	copy(h.Reserved[516:], []byte(file))
	return nil
}

// Metadata version constants
const (
	MetadataVersionV1 = 1 // Template-based library paths
)

// BinaryMetadata contains metadata for a single embedded binary
// Size: 512 bytes
type BinaryMetadata struct {
	Architecture     Architecture    // CPU architecture (4 bytes)
	ArchVersion      ArchVersion     // Architecture version (4 bytes)
	MetadataVersion  uint32          // Metadata format version (4 bytes)
	RequiredFeatures uint64          // Primary feature bitmask (8 bytes)
	ExtFeatures      [4]uint64       // Extended feature bitmasks (32 bytes)
	OriginalSize     uint64          // Uncompressed size (8 bytes)
	CompressedSize   uint64          // Compressed size (8 bytes)
	DataOffset       uint64          // Offset from start of file (8 bytes)
	Compression      CompressionAlgo // Compression algorithm (4 bytes)
	Checksum         [32]byte        // SHA-256 of uncompressed binary (32 bytes)
	Priority         uint32          // Selection priority (4 bytes)
	Format           BinaryFormat    // Binary format (ELF/Mach-O/PE) (4 bytes)
	LibPathFlags     uint32          // Library path flags (4 bytes)
	Reserved         [388]byte       // Reserved for future use (388 bytes)
}

// MarshalBinary encodes the metadata to binary format
func (m *BinaryMetadata) MarshalBinary() ([]byte, error) {
	buf := make([]byte, MetadataEntrySize)
	offset := 0

	// Architecture
	binary.LittleEndian.PutUint32(buf[offset:], uint32(m.Architecture))
	offset += 4

	// ArchVersion
	binary.LittleEndian.PutUint32(buf[offset:], uint32(m.ArchVersion))
	offset += 4

	// MetadataVersion
	binary.LittleEndian.PutUint32(buf[offset:], m.MetadataVersion)
	offset += 4

	// RequiredFeatures
	binary.LittleEndian.PutUint64(buf[offset:], m.RequiredFeatures)
	offset += 8

	// ExtFeatures
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(buf[offset:], m.ExtFeatures[i])
		offset += 8
	}

	// OriginalSize
	binary.LittleEndian.PutUint64(buf[offset:], m.OriginalSize)
	offset += 8

	// CompressedSize
	binary.LittleEndian.PutUint64(buf[offset:], m.CompressedSize)
	offset += 8

	// DataOffset
	binary.LittleEndian.PutUint64(buf[offset:], m.DataOffset)
	offset += 8

	// Compression
	binary.LittleEndian.PutUint32(buf[offset:], uint32(m.Compression))
	offset += 4

	// Checksum
	copy(buf[offset:], m.Checksum[:])
	offset += 32

	// Priority
	binary.LittleEndian.PutUint32(buf[offset:], m.Priority)
	offset += 4

	// Format
	binary.LittleEndian.PutUint32(buf[offset:], uint32(m.Format))
	offset += 4

	// LibPathFlags
	binary.LittleEndian.PutUint32(buf[offset:], m.LibPathFlags)
	offset += 4

	// Reserved
	copy(buf[offset:], m.Reserved[:])

	return buf, nil
}

// UnmarshalBinary decodes the metadata from binary format
func (m *BinaryMetadata) UnmarshalBinary(data []byte) error {
	if len(data) < MetadataEntrySize {
		return errors.New("insufficient data for metadata")
	}

	offset := 0

	// Architecture
	m.Architecture = Architecture(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	// ArchVersion
	m.ArchVersion = ArchVersion(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	// MetadataVersion
	m.MetadataVersion = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// RequiredFeatures
	m.RequiredFeatures = binary.LittleEndian.Uint64(data[offset:])
	offset += 8

	// ExtFeatures
	for i := 0; i < 4; i++ {
		m.ExtFeatures[i] = binary.LittleEndian.Uint64(data[offset:])
		offset += 8
	}

	// OriginalSize
	m.OriginalSize = binary.LittleEndian.Uint64(data[offset:])
	offset += 8

	// CompressedSize
	m.CompressedSize = binary.LittleEndian.Uint64(data[offset:])
	offset += 8

	// DataOffset
	m.DataOffset = binary.LittleEndian.Uint64(data[offset:])
	offset += 8

	// Compression
	m.Compression = CompressionAlgo(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	// Checksum
	copy(m.Checksum[:], data[offset:offset+32])
	offset += 32

	// Priority
	m.Priority = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// Format
	m.Format = BinaryFormat(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	// LibPathFlags
	m.LibPathFlags = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// Reserved
	copy(m.Reserved[:], data[offset:])

	return nil
}

// GetLibraryPathTemplates retrieves library path templates from metadata
// Returns empty slice if not set
//
// Storage format in Reserved:
// Bytes 0-1:   TemplateCount (uint16)
// Bytes 2-3:   Template1Length (uint16)
// Bytes 4-N:   Template1 string
// Bytes N+1-N+2: Template2Length (uint16)
// Bytes N+3-M:   Template2 string
// ...
func (m *BinaryMetadata) GetLibraryPathTemplates() []string {
	// Maximum reasonable template length (paths are typically < 256 chars)
	const maxTemplateLen = 512

	// Read template count
	templateCount := binary.LittleEndian.Uint16(m.Reserved[0:2])
	if templateCount == 0 || templateCount > 10 { // Sanity check
		return nil
	}

	templates := make([]string, 0, templateCount)
	offset := 2 // Start after count

	for i := 0; i < int(templateCount); i++ {
		// Check we have room for length field
		if offset+2 > len(m.Reserved) {
			break
		}

		// Read template length
		templateLen := binary.LittleEndian.Uint16(m.Reserved[offset : offset+2])
		offset += 2

		// Validate template length
		if templateLen == 0 || templateLen > maxTemplateLen {
			break // Invalid length, stop parsing
		}

		// Check we have room for template data (with overflow protection)
		if templateLen > uint16(len(m.Reserved)-offset) {
			break // Would overflow
		}

		// Additional check: verify offset + templateLen won't exceed bounds
		endOffset := offset + int(templateLen)
		if endOffset > len(m.Reserved) || endOffset < offset { // Check for integer overflow
			break
		}

		// Read template string
		templateBytes := m.Reserved[offset:endOffset]
		templates = append(templates, string(templateBytes))
		offset = endOffset
	}

	return templates
}

// SetLibraryPathTemplates stores library path templates in metadata
// Returns error if templates don't fit in Reserved space
func (m *BinaryMetadata) SetLibraryPathTemplates(templates []string) error {
	if len(templates) == 0 {
		return errors.New("no templates provided")
	}

	if len(templates) > 10 {
		return errors.New("too many templates (max 10)")
	}

	// Calculate required space
	requiredSpace := 2 // Template count
	for _, template := range templates {
		requiredSpace += 2 + len(template) // Length + string
	}

	if requiredSpace > len(m.Reserved) {
		return errors.New("templates too large for metadata storage")
	}

	// Set metadata version
	m.MetadataVersion = MetadataVersionV1

	// Clear Reserved
	for i := range m.Reserved {
		m.Reserved[i] = 0
	}

	// Write template count
	binary.LittleEndian.PutUint16(m.Reserved[0:2], uint16(len(templates)))
	offset := 2

	// Write each template
	for _, template := range templates {
		// Write length
		binary.LittleEndian.PutUint16(m.Reserved[offset:offset+2], uint16(len(template)))
		offset += 2

		// Write template string
		copy(m.Reserved[offset:], []byte(template))
		offset += len(template)
	}

	return nil
}

// Errors
var (
	ErrInvalidMagic       = errors.New("invalid magic marker")
	ErrUnsupportedVersion = errors.New("unsupported format version")
	ErrInvalidChecksum    = errors.New("checksum verification failed")
	ErrNoCompatibleBinary = errors.New("no compatible binary found")
	ErrInvalidMetadata    = errors.New("invalid metadata")
	ErrCorruptedBinary    = errors.New("corrupted fat binary")
)
