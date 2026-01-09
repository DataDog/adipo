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
	ReservedSize           = 160 // [160]byte
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
	ChecksumOffset           = ReservedOffset + ReservedSize                                           // 228

	// HeaderSize is the fixed size of the format header (derived from all fields)
	HeaderSize = MagicSize + VersionSize + NumBinariesSize + StubSizeSize +
		MetadataOffsetSize + MetadataSizeSize + DataOffsetSize +
		FlagsSize + CompressionAlgoSize + StubArchitectureSize + StubArchVersionSize +
		ReservedSize + ChecksumSize // 260

	// MetadataEntrySize is the fixed size of each binary metadata entry
	MetadataEntrySize = 256
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

// FormatFlags represents feature flags in the header
type FormatFlags uint64

const (
	FlagContainsX86_64  FormatFlags = 1 << 0
	FlagContainsARM64   FormatFlags = 1 << 1
	FlagMixedArch       FormatFlags = 1 << 2
	FlagStubCompressed  FormatFlags = 1 << 3
)

// FormatHeader is the fixed-size header at the start of the fat binary format
// Size: 260 bytes
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
	Reserved         [160]byte       // Reserved for future use
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

// BinaryMetadata contains metadata for a single embedded binary
// Size: 256 bytes
type BinaryMetadata struct {
	Architecture     Architecture    // CPU architecture (4 bytes)
	ArchVersion      ArchVersion     // Architecture version (4 bytes)
	RequiredFeatures uint64          // Primary feature bitmask (8 bytes)
	ExtFeatures      [4]uint64       // Extended feature bitmasks (32 bytes)
	OriginalSize     uint64          // Uncompressed size (8 bytes)
	CompressedSize   uint64          // Compressed size (8 bytes)
	DataOffset       uint64          // Offset from start of file (8 bytes)
	Compression      CompressionAlgo // Compression algorithm (4 bytes)
	Checksum         [32]byte        // SHA-256 of uncompressed binary (32 bytes)
	Priority         uint32          // Selection priority (4 bytes)
	Format           BinaryFormat    // Binary format (ELF/Mach-O/PE) (4 bytes)
	Reserved         [136]byte       // Reserved for future use (136 bytes)
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

	// Reserved
	copy(m.Reserved[:], data[offset:offset+136])

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
