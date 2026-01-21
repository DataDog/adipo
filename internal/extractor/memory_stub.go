// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


//go:build !linux

package extractor

import (
	"fmt"
)

// MemoryExtractor is a stub for non-Linux platforms
type MemoryExtractor struct{}

// Extract returns an error on non-Linux platforms
func (m *MemoryExtractor) Extract(data []byte, name string) (string, func(), error) {
	return "", nil, fmt.Errorf("memory extraction not supported on this platform")
}

// Name returns the extractor name
func (m *MemoryExtractor) Name() string {
	return "memory"
}
