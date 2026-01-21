// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package cpu

import (
	"fmt"
	"runtime"
)

// Detect detects the current CPU capabilities
func Detect() (*Capabilities, error) {
	var caps *Capabilities
	var err error

	switch runtime.GOARCH {
	case "amd64":
		caps, err = DetectX86_64()
	case "arm64":
		caps, err = DetectARM64()
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	if err != nil {
		return nil, err
	}

	// Detect CPU model (best effort, non-fatal)
	model, _ := DetectCPUModel()
	caps.CPUModel = model

	return caps, nil
}
