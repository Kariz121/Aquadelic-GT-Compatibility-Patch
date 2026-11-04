package core

import (
	"encoding/binary"
	"os"
	"testing"
)

func TestTTFFontRevisions(t *testing.T) {
	path := os.Getenv("AQGT_TTF_FIXTURE")
	if path == "" {
		t.Log("TTF fixture not specified")
		return
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	ent, h, n, e := fontHead(raw)
	if e != nil {
		t.Fatal(e)
	}
	_ = ent
	_ = n
	original := append([]byte(nil), raw...)
	binary.BigEndian.PutUint32(original[h+4:], FontOriginalRevision)
	// Normalize the test input TTF to the game revision and valid checksums.
	binary.BigEndian.PutUint32(original[h+8:], 0)
	binary.BigEndian.PutUint32(original[ent+4:], fontSum(original[h:h+n]))
	binary.BigEndian.PutUint32(original[h+8:], fontChecksumMagic-fontSum(original))
	v4, e := SetFontRevision(original, FontProbeRevision)
	if e != nil {
		t.Fatal(e)
	}
	v1000, e := SetFontRevision(v4, FontCompatRevision)
	if e != nil {
		t.Fatal(e)
	}
	if fontSum(v1000) != fontChecksumMagic {
		t.Fatal("font checksum bad")
	}
	reverted, e := SetFontRevision(v1000, FontOriginalRevision)
	if e != nil {
		t.Fatal(e)
	}
	if SHA(reverted) != SHA(original) {
		t.Fatal("TTF patch not reversible byte-for-byte")
	}
	if SHA(v1000) == SHA(v4) {
		t.Fatal("TTF unchanged")
	}
}
