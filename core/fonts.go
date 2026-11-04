package core

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Technically raised to 1000.0 solely for Wine/Proton compatibility. Not a
// statement about the upstream DejaVu font version, and not a font update.
const (
	FontOriginalRevision uint32 = 144179     // original game font, approx 2.20
	FontProbeRevision    uint32 = 4 << 16    // earlier Deck PROBE v2/v3
	FontCompatRevision   uint32 = 1000 << 16 // integrated compatibility revision
	fontChecksumMagic    uint32 = 0xB1B0AFBA
)

var FontNames = []string{
	"DejaVuSerifCondensed-BoldOblique.ttf",
	"DejaVuSansCondensed.ttf",
	"DejaVuSansCondensed-Bold.ttf",
	"DejaVuSansCondensed-Oblique.ttf",
	"DejaVuSansCondensed-BoldOblique.ttf",
}

func be16(b []byte, at int) uint16       { return binary.BigEndian.Uint16(b[at:]) }
func be32(b []byte, at int) uint32       { return binary.BigEndian.Uint32(b[at:]) }
func putBE32(b []byte, at int, x uint32) { binary.BigEndian.PutUint32(b[at:], x) }
func fontHead(b []byte) (entry, head, length int, err error) {
	if len(b) < 12 || !(string(b[:4]) == "\x00\x01\x00\x00" || string(b[:4]) == "true") {
		return 0, 0, 0, errors.New("unsupported TTF signature")
	}
	count := int(be16(b, 4))
	if count < 1 || count > 100 || 12+count*16 > len(b) {
		return 0, 0, 0, errors.New("invalid TTF table directory")
	}
	for j := 0; j < count; j++ {
		at := 12 + j*16
		if string(b[at:at+4]) != "head" {
			continue
		}
		h, l := int(be32(b, at+8)), int(be32(b, at+12))
		if h < 0 || l < 54 || h > len(b) || l > len(b)-h || h%4 != 0 || be32(b, h+12) != 0x5f0f3cf5 {
			return 0, 0, 0, errors.New("invalid TTF head table")
		}
		return at, h, l, nil
	}
	return 0, 0, 0, errors.New("TTF head table missing")
}
func FontRevision(b []byte) (uint32, error) {
	_, h, _, e := fontHead(b)
	if e != nil {
		return 0, e
	}
	return be32(b, h+4), nil
}
func fontSum(b []byte) uint32 {
	var s uint32
	for at := 0; at < len(b); at += 4 {
		var w [4]byte
		copy(w[:], b[at:])
		s += binary.BigEndian.Uint32(w[:])
	}
	return s
}
func SetFontRevision(in []byte, revision uint32) ([]byte, error) {
	entry, head, length, e := fontHead(in)
	if e != nil {
		return nil, e
	}
	current := be32(in, head+4)
	if current != FontOriginalRevision && current != FontProbeRevision && current != FontCompatRevision {
		return nil, fmt.Errorf("unknown TTF fontRevision %d", current)
	}
	if revision != FontOriginalRevision && revision != FontProbeRevision && revision != FontCompatRevision {
		return nil, errors.New("unsupported target revision")
	}
	out := append([]byte(nil), in...)
	if current == revision {
		return out, nil
	}
	putBE32(out, head+4, revision)
	putBE32(out, head+8, 0)
	putBE32(out, entry+4, fontSum(out[head:head+length]))
	putBE32(out, head+8, fontChecksumMagic-fontSum(out))
	if fontSum(out) != fontChecksumMagic {
		return nil, errors.New("invalid TTF checksum after patch")
	}
	return out, nil
}
