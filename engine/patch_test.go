package engine

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func testAQPT(before, after []byte) []byte {
	out := make([]byte, 16)
	copy(out, "AQPT")
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(before)))
	binary.LittleEndian.PutUint32(out[8:12], uint32(len(after)))
	binary.LittleEndian.PutUint32(out[12:16], 1)
	rec := make([]byte, 8)
	binary.LittleEndian.PutUint32(rec[:4], 2)
	binary.LittleEndian.PutUint32(rec[4:8], 2)
	return append(append(out, rec...), after[2:4]...)
}
func TestAQPTExactHashes(t *testing.T) {
	old := []byte("abcdefgh")
	target := []byte("abXYefgh")
	patch := testAQPT(old, target)
	got, err := ApplyAQPT(old, patch, SHA(old), SHA(target))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, target) {
		t.Fatalf("output differs: %q", got)
	}
	if _, err = ApplyAQPT([]byte("ABCDEFGH"), patch, SHA(old), SHA(target)); err == nil {
		t.Fatal("unknown source accepted")
	}
	if _, err = ApplyAQPT(old, patch, SHA(old), SHA(old)); err == nil {
		t.Fatal("wrong target accepted")
	}
	if _, err = ApplyAQPT(old, patch[:len(patch)-1], SHA(old), SHA(target)); err == nil {
		t.Fatal("truncated patch accepted")
	}
	malicious := append([]byte(nil), patch...)
	binary.LittleEndian.PutUint32(malicious[16:20], uint32(len(old)+1))
	if _, err = ApplyAQPT(old, malicious, SHA(old), SHA(target)); err == nil {
		t.Fatal("out-of-bounds patch accepted")
	}
}
func TestPatchSteamRefusesUnknown(t *testing.T) {
	for _, platform := range []string{"windows", "linux"} {
		for _, mode := range []string{"none", "analog", "step"} {
			if _, err := PatchSteamRun([]byte("Not the original game"), platform, mode); err == nil {
				t.Fatalf("unknown build accepted: %s %s", platform, mode)
			}
		}
	}
	if ModeValid("bad-mode") {
		t.Fatal("unknown mode accepted")
	}
}
