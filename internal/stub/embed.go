package stub

//go:generate sh -c "cd ../../stub && go build -trimpath -ldflags='-s -w' -o /tmp/adipo-stub-build.bin . && printf 'package stub\\n\\nfunc init() {\\n\\tStubBinary = []byte{' > ../internal/stub/stub_generated.go && od -An -tx1 -v /tmp/adipo-stub-build.bin | tr -d ' \\n' | sed 's/../0x&,/g' | sed 's/,$//' >> ../internal/stub/stub_generated.go && printf '}\\n}\\n' >> ../internal/stub/stub_generated.go && rm /tmp/adipo-stub-build.bin"

import (
	"fmt"
)

// StubBinary contains the stub binary
// This is populated by go:generate which creates stub_generated.go
// If not generated, StubBinary will be nil
var StubBinary []byte

// GetStubBinary returns the stub binary
func GetStubBinary() ([]byte, error) {
	if len(StubBinary) == 0 {
		return nil, fmt.Errorf("stub binary not available - use --stub-path to provide external stub, --no-stub to skip, or run 'go generate ./internal/stub' to build embedded stub")
	}
	return StubBinary, nil
}

// StubSize returns the size of the stub
func StubSize() int {
	return len(StubBinary)
}
