package cpu

import (
	"fmt"
	"runtime"
)

// Detect detects the current CPU capabilities
func Detect() (*Capabilities, error) {
	switch runtime.GOARCH {
	case "amd64":
		return DetectX86_64()
	case "arm64":
		return DetectARM64()
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
}
