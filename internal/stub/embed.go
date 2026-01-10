package stub

// Embedding support has been removed in favor of automatic stub discovery
// and self-stub fallback mode. See cmd/adipo/create.go for the new stub loading logic.

import (
	"fmt"
)

// GetStubBinary always returns an error as embedding is no longer supported
func GetStubBinary() ([]byte, error) {
	return nil, fmt.Errorf("embedded stub is no longer supported - adipo looks for adipo-stub[-{os}-{arch}] next to the adipo binary, or use --stub-path to specify stub location, or use --no-stub")
}
