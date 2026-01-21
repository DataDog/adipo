// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.

//go:build !capstone

package analyzer

import (
	"errors"

	"github.com/DataDog/adipo/internal/format"
)

// CapstoneDisassembler is a stub when Capstone is not available
type CapstoneDisassembler struct{}

// NewCapstoneDisassembler returns an error when Capstone is not available
func NewCapstoneDisassembler(arch format.Architecture) (*CapstoneDisassembler, error) {
	return nil, errors.New("Capstone support not compiled in. Rebuild with: go build -tags capstone")
}

// DisassembleBytes is a stub
func (d *CapstoneDisassembler) DisassembleBytes(data []byte, maxInstructions int) ([]Instruction, error) {
	return nil, errors.New("Capstone not available")
}

// MapCapstoneGroupsToFeatures is a stub
func MapCapstoneGroupsToFeatures(arch format.Architecture, groups []uint8) uint64 {
	return 0
}

// IsCapstoneAvailable returns false when Capstone is not compiled in
func IsCapstoneAvailable() bool {
	return false
}
