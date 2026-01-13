// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


//go:build !arm64

package cpu

import (
	"fmt"
)

// DetectARM64 is a stub for non-arm64 architectures
func DetectARM64() (*Capabilities, error) {
	return nil, fmt.Errorf("ARM64 detection not available on this architecture")
}
