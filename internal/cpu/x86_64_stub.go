//go:build !amd64

package cpu

import (
	"fmt"
)

// DetectX86_64 is a stub for non-amd64 architectures
func DetectX86_64() (*Capabilities, error) {
	return nil, fmt.Errorf("x86-64 detection not available on this architecture")
}
