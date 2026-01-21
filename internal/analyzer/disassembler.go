// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package analyzer

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ObjdumpVariant represents the type of objdump binary being used
type ObjdumpVariant int

const (
	// ObjdumpUnknown indicates we couldn't determine the objdump variant
	ObjdumpUnknown ObjdumpVariant = iota
	// ObjdumpGNU indicates GNU binutils objdump (typically on Linux)
	ObjdumpGNU
	// ObjdumpLLVM indicates LLVM objdump (typically on macOS)
	ObjdumpLLVM
)

// String returns a string representation of the objdump variant
func (v ObjdumpVariant) String() string {
	switch v {
	case ObjdumpGNU:
		return "GNU"
	case ObjdumpLLVM:
		return "LLVM"
	default:
		return "Unknown"
	}
}

// Disassembler handles binary disassembly using objdump
type Disassembler struct {
	objdumpPath string
	variant     ObjdumpVariant
}

// Instruction represents a single disassembled instruction
type Instruction struct {
	Address  uint64
	Mnemonic string
	Operands string
}

// NewDisassembler creates a new disassembler with the specified objdump path
// If objdumpPath is empty, it will search for objdump in PATH
func NewDisassembler(objdumpPath string) (*Disassembler, error) {
	// If no path specified, search in PATH
	if objdumpPath == "" {
		path, err := exec.LookPath("objdump")
		if err != nil {
			return nil, fmt.Errorf("objdump not found in PATH: %w\nInstall binutils (Linux) or Xcode Command Line Tools (macOS)", err)
		}
		objdumpPath = path
	}

	// Detect variant
	variant, err := detectObjdumpVariant(objdumpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect objdump variant: %w", err)
	}

	return &Disassembler{
		objdumpPath: objdumpPath,
		variant:     variant,
	}, nil
}

// detectObjdumpVariant detects whether this is GNU or LLVM objdump
func detectObjdumpVariant(objdumpPath string) (ObjdumpVariant, error) {
	cmd := exec.Command(objdumpPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return ObjdumpUnknown, fmt.Errorf("failed to run objdump --version: %w", err)
	}

	versionStr := string(output)
	if strings.Contains(versionStr, "GNU") {
		return ObjdumpGNU, nil
	}
	if strings.Contains(versionStr, "LLVM") {
		return ObjdumpLLVM, nil
	}

	return ObjdumpUnknown, fmt.Errorf("unknown objdump variant: %s", versionStr)
}

// Disassemble disassembles the binary at the given path and returns instructions
func (d *Disassembler) Disassemble(binaryPath string, maxInstructions int) ([]Instruction, error) {
	// Build command arguments based on variant
	args := []string{"-d"}
	if d.variant == ObjdumpGNU {
		args = append(args, "--no-show-raw-insn")
	}
	args = append(args, binaryPath)

	cmd := exec.Command(d.objdumpPath, args...)

	// Get stdout pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start objdump: %w", err)
	}

	// Parse output line by line
	var instructions []Instruction
	scanner := bufio.NewScanner(stdout)

	// Increase buffer size for long lines
	const maxScanTokenSize = 1024 * 1024 // 1MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	for scanner.Scan() {
		line := scanner.Text()

		if insn, ok := d.parseLine(line); ok {
			instructions = append(instructions, insn)

			// Check if we've reached the limit
			if maxInstructions > 0 && len(instructions) >= maxInstructions {
				// Kill the process since we have enough
				_ = cmd.Process.Kill()
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// If we killed the process, ignore broken pipe errors
		if maxInstructions > 0 && len(instructions) >= maxInstructions {
			_ = cmd.Wait()
			return instructions, nil
		}
		return nil, fmt.Errorf("error reading objdump output: %w", err)
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		// If we killed the process, ignore the exit error
		if maxInstructions > 0 && len(instructions) >= maxInstructions {
			return instructions, nil
		}
		return nil, fmt.Errorf("objdump failed: %w", err)
	}

	if len(instructions) == 0 {
		return nil, fmt.Errorf("no instructions found (binary may be stripped or invalid)")
	}

	return instructions, nil
}

// parseLine parses a single line of objdump output
// Expected format: "  address: <tab> [raw_bytes] mnemonic <tab> operands"
// GNU:  "  40053b:       mov    %rsp,%rbp"
// LLVM: "  406630: e5e1fcbf     	st1d	{ z31.d }, p7, [x5, #0x1, mul vl]"
// The raw bytes are 8 hex digits (32-bit instruction encoding) followed by whitespace
var insnRegex = regexp.MustCompile(`^\s*([0-9a-fA-F]+):\s+(?:[0-9a-fA-F]{8}\s+)?([a-zA-Z][a-zA-Z0-9.]*)(?:\s+(.*))?$`)

func (d *Disassembler) parseLine(line string) (Instruction, bool) {
	// Skip empty lines and header lines
	line = strings.TrimSpace(line)
	if line == "" || !strings.Contains(line, ":") {
		return Instruction{}, false
	}

	// Match instruction line
	matches := insnRegex.FindStringSubmatch(line)
	if matches == nil {
		return Instruction{}, false
	}

	// Parse address
	address, err := strconv.ParseUint(matches[1], 16, 64)
	if err != nil {
		return Instruction{}, false
	}

	// Extract mnemonic and operands
	mnemonic := strings.ToLower(matches[2])
	operands := ""
	if len(matches) > 3 {
		operands = strings.TrimSpace(matches[3])
	}

	// Remove comments (everything after "# " with space after hash)
	// This preserves ARM immediate values like #0x1, #2 while removing "# comment"
	// Comments have a space after #, immediates don't
	for _, sep := range []string{" # ", "\t# "} {
		if idx := strings.Index(operands, sep); idx != -1 {
			operands = strings.TrimSpace(operands[:idx])
			break
		}
	}

	return Instruction{
		Address:  address,
		Mnemonic: mnemonic,
		Operands: operands,
	}, true
}

// Variant returns the detected objdump variant
func (d *Disassembler) Variant() ObjdumpVariant {
	return d.variant
}
