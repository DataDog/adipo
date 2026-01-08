// +build !arm64

package cpu

import (
	"fmt"
)

// DetectARM64 is a stub for non-arm64 architectures
func DetectARM64() (*Capabilities, error) {
	return nil, fmt.Errorf("ARM64 detection not available on this architecture")
}
