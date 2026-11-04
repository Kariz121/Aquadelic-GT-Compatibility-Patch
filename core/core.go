package core

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	RetailOriginal   = "36fcf10bd35f0ea3844b16d6ecc8ac67752ce878f7a40ace5192d91d2017d89a"
	RetailPatched    = "949768e738e43aa2240176b752ac2aad9d0b66dc442aa87fd21634b58d738aa5"
	SteamOriginal    = "c275d5bde9fdb9fa172450c28e70927d443cc4b43b08a16dce923b4b7e9dd732"
	SteamBoolean     = "ef1bde8a7ff003334cf2c4f04d54bdf5c5aedb8fffc60ce6e880829b91aaed94"
	SteamMultiplayer = "b2804c68d184d084fbf590d68bf3a7f558fd0b3096f7db6032f6445d4e5af782"
	SteamBoth        = "d73ab0a65776343b9043986e7b508008c873bb758917841d6c72ae5ec756e94c"
)

func SHA(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

type Kind string

const (
	Unknown                Kind = "unknown"
	RetailClean            Kind = "retail-clean"
	RetailDone             Kind = "retail-patched"
	SteamClean             Kind = "steam-clean"
	SteamBoolDone          Kind = "steam-boolean"
	SteamMpDone            Kind = "steam-multiplayer"
	SteamBothDone          Kind = "steam-both"
	SteamMD5Done           Kind = "steam-md5"
	SteamMD5BoolDone       Kind = "steam-md5-boolean"
	SteamMD5MpDone         Kind = "steam-md5-multiplayer"
	SteamMD5BothDone       Kind = "steam-md5-both"
	SteamBroadcastDone     Kind = "steam-broadcast-v2"
	SteamMD5BroadcastDone  Kind = "steam-md5-broadcast-v2"
	SteamAllControlsDone   Kind = "steam-md5-broadcast-controls"
	SteamNoMD5ControlsDone Kind = "steam-broadcast-controls-no-md5"
)

func Detect(data []byte) Kind {
	switch SHA(data) {
	case SteamControlsNoMD5FinalSHA:
		return SteamNoMD5ControlsDone
	case SteamControlsFinalSHA:
		return SteamAllControlsDone
	case "988ad60b0bb0d95c4c6130b379e478c612fc820cac963a2348a89d601666e1ce":
		return SteamBroadcastDone
	case "4be02e05a3b59fcaf3120b8feffa636f7e9e07dba6c4e8d82976f3aed07a7ae4":
		return SteamMD5BroadcastDone
	case SteamMD5:
		return SteamMD5Done
	case md5BoolHash:
		return SteamMD5BoolDone
	case md5MpHash:
		return SteamMD5MpDone
	case md5BothHash:
		return SteamMD5BothDone
	case RetailOriginal:
		return RetailClean
	case RetailPatched:
		return RetailDone
	case SteamOriginal:
		return SteamClean
	case SteamBoolean:
		return SteamBoolDone
	case SteamMultiplayer:
		return SteamMpDone
	case SteamBoth:
		return SteamBothDone
	case "679d029cc5440effd16088701467a43bbd9d68eefd03433eb2b41ded2cd29bd9":
		return SteamBoolDone
	case "fbfaee86fa1d5db4b1cf16b5774dccff2a0b8e45f3630de7cdaa59da4eb67602":
		return SteamMpDone
	case "40f3eca8d7cf82d7dd4e0d41771d1ba0cba5373d1dc3b7113a8e22aa69d86534":
		return SteamBothDone
	default:
		return Unknown
	}
}

type patchRegion struct {
	off  int
	data []byte
}

func ApplyRetail(data []byte) ([]byte, error) {
	if SHA(data) != RetailOriginal {
		return nil, errors.New("not the verified Polish retail revision")
	}
	out := append([]byte(nil), data...)
	for _, r := range retailPatches {
		if r.off+len(r.data) > len(out) {
			return nil, errors.New("patch outside executable")
		}
		copy(out[r.off:], r.data)
	}
	if SHA(out) != RetailPatched {
		return nil, errors.New("retail output hash mismatch; refusing patch")
	}
	return out, nil
}

type Mode string

const (
	Bool        Mode = "boolean"
	Multiplayer Mode = "multiplayer"
	Both        Mode = "both"
)

func (m Mode) Hash() string {
	switch m {
	case Bool:
		return SteamBoolean
	case Multiplayer:
		return SteamMultiplayer
	case Both:
		return SteamBoth
	}
	return ""
}

var boolOperands = []int{0x0D3364, 0x0D34CC, 0x0D3A78, 0x0D3BDC, 0x0D3D3C, 0x0D4DD0, 0x0D6FEC, 0x0D7150, 0x0D7414, 0x0D7570, 0x0D78A4, 0x0D8014, 0x0D8170, 0x0D8388, 0x0D83FA}

const (
	steamVarFromBool   = 0x00415DB8
	steamVarClearDeep  = 0x0041176C
	steamMpClearVA     = 0x005370B3
	steamMpClearOffset = 0x1364B3
	sectionAlignment   = 0x1000
	fileAlignment      = 0x200
)

func u16(b []byte, off int) int      { return int(binary.LittleEndian.Uint16(b[off:])) }
func u32(b []byte, off int) int      { return int(binary.LittleEndian.Uint32(b[off:])) }
func put16(b []byte, off int, x int) { binary.LittleEndian.PutUint16(b[off:], uint16(x)) }
func put32(b []byte, off int, x int) { binary.LittleEndian.PutUint32(b[off:], uint32(x)) }
func align(n, n2 int) int            { return (n + n2 - 1) / n2 * n2 }
func relative(insnVA, destVA int) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(int32(destVA-(insnVA+5))))
	return b
}
func bfromhex(s string) []byte {
	b, e := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	if e != nil {
		panic(e)
	}
	return b
}

func checksumPE(data []byte, checkOff int) uint32 {
	var sum uint64
	for i := 0; i < len(data); i += 2 {
		var word uint32
		if i != checkOff && i != checkOff+2 {
			word = uint32(data[i])
			if i+1 < len(data) {
				word |= uint32(data[i+1]) << 8
			}
		}
		sum += uint64(word)
		sum = (sum & 0xffff) + (sum >> 16)
	}
	sum = (sum & 0xffff) + (sum >> 16)
	return uint32(sum + uint64(len(data)))
}
func buildBoolHelper(va int) []byte {
	b := bfromhex("53 56 8b da 8b f0 66 f7 06 e8 bf 74 07 8b c6 e8 00 00 00 00 c7 06 03 00 00 00 c7 46 04 00 00 00 00 33 c0 84 db 0f 95 c0 89 46 08 c7 46 0c 00 00 00 00 5e 5b c3")
	copy(b[0x10:0x14], relative(va+0x0f, steamVarClearDeep))
	if len(b) != 0x35 {
		panic("unexpected helper size")
	}
	return b
}
func buildMpHelper() []byte { return bfromhex("33 c0 89 43 20 c6 43 24 00 c3") }

func BuildSteam(original []byte, mode Mode) ([]byte, error) {
	if mode.Hash() == "" {
		return nil, errors.New("invalid Steam patch mode")
	}
	if SHA(original) != SteamOriginal {
		return nil, errors.New("Steam original SHA-256 does not match the verified v1.4.0 EXE")
	}
	if len(original) < 0x400 || string(original[:2]) != "MZ" {
		return nil, errors.New("invalid PE file")
	}
	pe := u32(original, 0x3c)
	if pe+248+40*11 > len(original) || string(original[pe:pe+4]) != "PE\x00\x00" {
		return nil, errors.New("invalid PE header")
	}
	if u16(original, pe+4) != 0x14c || u16(original, pe+24) != 0x10b {
		return nil, errors.New("not i386 PE32")
	}
	count := u16(original, pe+6)
	opt := pe + 24
	table := opt + u16(original, pe+20)
	if count != 10 || table+40*(count+1) > u32(original, opt+60) {
		return nil, errors.New("unexpected section table")
	}
	if u32(original, opt+36) != fileAlignment || u32(original, opt+32) != sectionAlignment {
		return nil, errors.New("unexpected PE alignment")
	}
	last := table + 40*(count-1)
	if string(original[table:table+5]) != ".text" || string(original[last:last+6]) != ".debug" {
		return nil, errors.New("unexpected section names")
	}
	va := u32(original, last+12)
	vs := u32(original, last+8)
	rawSize := u32(original, last+16)
	rawPtr := u32(original, last+20)
	if rawPtr+rawSize != len(original) {
		return nil, errors.New("unexpected overlay at end of executable")
	}
	newVA := align(va+vs, sectionAlignment)
	if u32(original, opt+56) != newVA {
		return nil, errors.New("unexpected SizeOfImage")
	}
	for _, off := range boolOperands {
		if original[off-1] != 0xe8 {
			return nil, fmt.Errorf("Boolean call opcode mismatch at %X", off)
		}
		disp := int(int32(binary.LittleEndian.Uint32(original[off:])))
		insnVA := off + 0x400c00 - 1
		dest := insnVA + 5 + disp
		if dest != steamVarFromBool {
			return nil, fmt.Errorf("Boolean target mismatch at %X", off)
		}
	}
	if !strings.EqualFold(hex.EncodeToString(original[steamMpClearOffset:steamMpClearOffset+5]), "33c0894320") {
		return nil, errors.New("multiplayer opcode mismatch")
	}
	newRawOffset := align(len(original), fileAlignment)
	out := make([]byte, newRawOffset+fileAlignment)
	copy(out, original)
	payload := make([]byte, fileAlignment)
	copy(payload, []byte("AQGT STEAM PROTON TEST 0.1\x00"))
	if mode == Bool || mode == Both {
		helperVA := 0x400000 + newVA + 0x40
		copy(payload[0x40:], buildBoolHelper(helperVA))
		for _, off := range boolOperands {
			insnVA := off + 0x400c00 - 1
			copy(out[off:off+4], relative(insnVA, helperVA))
		}
	}
	if mode == Multiplayer || mode == Both {
		helperVA := 0x400000 + newVA + 0x80
		copy(payload[0x80:], buildMpHelper())
		out[steamMpClearOffset] = 0xe8
		copy(out[steamMpClearOffset+1:], relative(steamMpClearVA, helperVA))
	}
	copy(out[newRawOffset:], payload)
	hdr := table + 40*count
	copy(out[hdr:hdr+8], []byte{'.', 'a', 'q', 'f', 'i', 'x', 0, 0})
	put32(out, hdr+8, 0x100)
	put32(out, hdr+12, newVA)
	put32(out, hdr+16, fileAlignment)
	put32(out, hdr+20, newRawOffset)
	put32(out, hdr+36, 0x60000020)
	put16(out, pe+6, count+1)
	put32(out, opt+4, u32(out, opt+4)+fileAlignment)
	put32(out, opt+56, newVA+sectionAlignment)
	checkOff := opt + 64
	put32(out, checkOff, int(checksumPE(out, checkOff)))

	if SHA(out) != mode.Hash() {
		return nil, fmt.Errorf("Steam patch verification failed: %s", SHA(out))
	}
	return out, nil
}
