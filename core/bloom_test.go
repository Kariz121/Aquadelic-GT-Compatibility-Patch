package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBloomReversible(t *testing.T) {
	p := os.Getenv("AQGT_BLOOM_DATA")
	if p == "" {
		t.Skip("fixture dir required")
	}
	for _, f := range BloomFiles {
		b, err := os.ReadFile(filepath.Join(p, filepath.FromSlash(f)))
		if err != nil {
			t.Fatal(err)
		}
		patched, err := PatchBloomXML(b)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if bytes.Equal(patched, b) {
			t.Fatalf("%s unchanged (expected source original)", f)
		}
		again, err := PatchBloomXML(patched)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(again, patched) {
			t.Fatalf("%s is not idempotent", f)
		}
	}
}
