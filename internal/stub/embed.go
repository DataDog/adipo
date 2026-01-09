package stub

//go:generate sh -c "cd ../../stub && go build -trimpath -ldflags='-s -w' -o ../internal/stub/stub.bin ."

import (
	_ "embed"
	"fmt"
)

// StubBinary contains the embedded stub binary
// This will be automatically built for the current host OS/arch via go:generate
//
//go:embed stub.bin
var StubBinary []byte

// GetStubBinary returns the embedded stub binary
func GetStubBinary() ([]byte, error) {
	if len(StubBinary) == 0 {
		return nil, fmt.Errorf("stub binary not embedded (run 'make' to build)")
	}
	return StubBinary, nil
}

// StubSize returns the size of the embedded stub
func StubSize() int {
	return len(StubBinary)
}
