// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.

package integration_test

import (
	"bytes"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
	"github.com/DataDog/adipo/internal/selector"
)

// TestVendorCPUHintArchiveSelection exercises persisted hints through archive
// creation, inspection, runtime selection and payload extraction. Synthetic
// payloads and CPU capabilities make the test independent of the host CPU.
func TestVendorCPUHintArchiveSelection(t *testing.T) {
	for _, hint := range []string{"graviton4", "google-axion", "nvidia-grace"} {
		t.Run(hint, func(t *testing.T) {
			if _, err := cpu.ValidateCPUHint(hint, format.ArchARM64); err != nil {
				t.Fatal(err)
			}
			templates := []string{"/opt/{{.ArchVersion}}/lib", "/opt/{{.CPUAlias}}/lib"}
			generic := &format.BinaryMetadata{
				Architecture: format.ArchARM64,
				ArchVersion:  format.ARM64_V9_0,
				Format:       format.FormatELF,
			}
			tuned := &format.BinaryMetadata{
				Architecture: format.ArchARM64,
				ArchVersion:  format.ARM64_V9_0,
				Format:       format.FormatELF,
			}
			if err := tuned.SetLibraryPathTemplates(templates); err != nil {
				t.Fatal(err)
			}
			if err := tuned.SetCPUHint(hint); err != nil {
				t.Fatal(err)
			}
			genericData := []byte("generic payload")
			// The tuned payload loses on size unless its hint earns the bonus.
			tunedData := bytes.Repeat([]byte{0x42}, 1024*1024)
			entries := []*format.BinaryEntry{
				{Data: genericData, OriginalData: genericData, Metadata: generic},
				{Data: tunedData, OriginalData: tunedData, Metadata: tuned},
			}
			path := filepath.Join(t.TempDir(), "app.fat")
			// The reader requires a stub of at least 100 KiB.
			if err := format.WriteToFile(path, make([]byte, 100*1024), entries,
				format.ArchARM64, format.ARM64_V8_0, 0, "", ""); err != nil {
				t.Fatal(err)
			}
			reader, err := format.OpenFile(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := reader.Close(); err != nil {
					t.Errorf("close archive: %v", err)
				}
			}()
			if err := reader.VerifyChecksum(); err != nil {
				t.Fatal(err)
			}
			metadata := reader.Metadata()
			if len(metadata) != 2 {
				t.Fatalf("archive contains %d entries, want 2", len(metadata))
			}
			if got := metadata[1].GetCPUHint(); got != hint {
				t.Errorf("stored hint = %q, want original vendor name %q", got, hint)
			}
			if got := metadata[1].GetLibraryPathTemplates(); !slices.Equal(got, templates) {
				t.Errorf("stored templates = %v, want %v", got, templates)
			}
			caps := &cpu.Capabilities{
				ArchType: format.ArchARM64,
				Version:  format.ARM64_V9_0,
				CPUModel: &cpu.CPUModel{Implementer: 0x41, PartNum: 0xd4f},
			}
			index, selected, err := selector.NewSelector(caps, metadata).SelectBinary()
			if err != nil {
				t.Fatal(err)
			}
			if index != 1 || selected.GetCPUHint() != hint {
				t.Fatalf("selected entry %d with hint %q, want entry 1 with %q", index, selected.GetCPUHint(), hint)
			}
			data, err := reader.GetBinaryData(index)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, tunedData) {
				t.Error("extracted payload differs from selected tuned payload")
			}
		})
	}
}
