// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


//go:build !amd64

package cpu

import (
	"fmt"
)

// DetectX86_64 is a stub for non-amd64 architectures
func DetectX86_64() (*Capabilities, error) {
	return nil, fmt.Errorf("x86-64 detection not available on this architecture")
}
