// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package analyzer

import (
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		want     Instruction
		wantOk   bool
	}{
		{
			name: "GNU objdump x86-64 format",
			line: "  40053b:       mov    %rsp,%rbp",
			want: Instruction{
				Address:  0x40053b,
				Mnemonic: "mov",
				Operands: "%rsp,%rbp",
			},
			wantOk: true,
		},
		{
			name: "LLVM objdump ARM64 SVE format",
			line: "  406630: e5e1fcbf     	st1d	{ z31.d }, p7, [x5, #0x1, mul vl]",
			want: Instruction{
				Address:  0x406630,
				Mnemonic: "st1d",
				Operands: "{ z31.d }, p7, [x5, #0x1, mul vl]",
			},
			wantOk: true,
		},
		{
			name: "LLVM objdump ARM64 ld1w format",
			line: "  409fd0: a5405cbf     	ld1w	{ z31.s }, p7/z, [x5, x0, lsl #2]",
			want: Instruction{
				Address:  0x409fd0,
				Mnemonic: "ld1w",
				Operands: "{ z31.s }, p7/z, [x5, x0, lsl #2]",
			},
			wantOk: true,
		},
		{
			name: "LLVM objdump ARM64 ptrue format",
			line: "  40c52c: 25e30fe7     	whilelo	p7.d, wzr, w3",
			want: Instruction{
				Address:  0x40c52c,
				Mnemonic: "whilelo",
				Operands: "p7.d, wzr, w3",
			},
			wantOk: true,
		},
		{
			name: "GNU objdump x86-64 with comment",
			line: "  401234:       add    $0x10,%rsp   # some comment",
			want: Instruction{
				Address:  0x401234,
				Mnemonic: "add",
				Operands: "$0x10,%rsp",
			},
			wantOk: true,
		},
		{
			name: "LLVM objdump ARM64 with comment",
			line: "  40c4e0: a4035a83     	ld1b	{ z3.b }, p6/z, [x20, x3] # load byte",
			want: Instruction{
				Address:  0x40c4e0,
				Mnemonic: "ld1b",
				Operands: "{ z3.b }, p6/z, [x20, x3]",
			},
			wantOk: true,
		},
		{
			name: "GNU objdump ARM64 NEON format",
			line: "  1234:       fadd   v0.4s, v1.4s, v2.4s",
			want: Instruction{
				Address:  0x1234,
				Mnemonic: "fadd",
				Operands: "v0.4s, v1.4s, v2.4s",
			},
			wantOk: true,
		},
		{
			name: "LLVM objdump with longer raw bytes",
			line: "  40bf88: e4075cbe     	st1b	{ z30.b }, p7, [x5, x7]",
			want: Instruction{
				Address:  0x40bf88,
				Mnemonic: "st1b",
				Operands: "{ z30.b }, p7, [x5, x7]",
			},
			wantOk: true,
		},
		{
			name:   "Empty line",
			line:   "",
			want:   Instruction{},
			wantOk: false,
		},
		{
			name:   "Header line",
			line:   "Disassembly of section .text:",
			want:   Instruction{},
			wantOk: false,
		},
		{
			name:   "Function label",
			line:   "0000000000400530 <main>:",
			want:   Instruction{},
			wantOk: false,
		},
		{
			name:   "Line with no address",
			line:   "   mov    %rsp,%rbp",
			want:   Instruction{},
			wantOk: false,
		},
	}

	// Create a disassembler with dummy values for testing
	d := &Disassembler{
		objdumpPath: "/usr/bin/objdump",
		variant:     ObjdumpLLVM,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := d.parseLine(tt.line)
			if ok != tt.wantOk {
				t.Errorf("parseLine() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if !ok {
				return
			}
			if got.Address != tt.want.Address {
				t.Errorf("parseLine() Address = 0x%x, want 0x%x", got.Address, tt.want.Address)
			}
			if got.Mnemonic != tt.want.Mnemonic {
				t.Errorf("parseLine() Mnemonic = %q, want %q", got.Mnemonic, tt.want.Mnemonic)
			}
			if got.Operands != tt.want.Operands {
				t.Errorf("parseLine() Operands = %q, want %q", got.Operands, tt.want.Operands)
			}
		})
	}
}
