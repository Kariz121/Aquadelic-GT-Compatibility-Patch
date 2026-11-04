package core

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Reference hashes are exclusively for the exact 2026 supplied revisions.
const (
	SteamMD5                 = "099e1b7f2f85f4987921f43d51d4b150705f49809999a285ea7c500bc3b2e62d"
	Launcher140Original      = "f33d1e405c1c539063bce68ba5b2226c3daaf3a4e97268c122e62481a4b1db4c"
	Launcher140Mesa          = "e890db5f3dedf869765a119039917e6fe0bb5ac2bb84ee17cfcaf49e591e694f"
	Launcher140Full          = "087282c4db7465e52ce61eb2e4ca034b70d343e5649b556e11528b5cb51d71e5"
	LauncherRetailUnverified = "e251f6b76b5719f90de0ae1a678c241a062eebf5a33e7f25f45195eeb980e382"
)

func applyDelta(input []byte, d deltaPatch) ([]byte, error) {
	if SHA(input) != d.Source {
		return nil, fmt.Errorf("expected exact source SHA-256 %s, got %s", d.Source, SHA(input))
	}
	result := append([]byte(nil), input...)
	for _, r := range d.Regions {
		if r.Offset < 0 || r.Offset+len(r.Old) > len(input) || len(r.Old) != len(r.New) {
			return nil, errors.New("patch region out of bounds")
		}
		for i, v := range r.Old {
			if input[r.Offset+i] != v {
				return nil, fmt.Errorf("unexpected source at 0x%x", r.Offset+i)
			}
		}
		copy(result[r.Offset:], r.New)
	}
	result = append(result, d.Suffix...)
	if SHA(result) != d.Target {
		return nil, errors.New("full output SHA-256 verification failed")
	}
	return result, nil
}

// ApplyLauncher140 patches both the OpenGL Mesa version parser and the XML
// Integrity normalization. It rejects the DIFFERENT retail Launcher.exe.
func ApplyLauncher140(in []byte) ([]byte, error) {
	switch SHA(in) {
	case Launcher140Original:
		return applyDelta(in, deltaLauncherFull)
	case Launcher140Mesa:
		return applyDelta(in, deltaLauncherUpgrade)
	case Launcher140Full:
		return nil, errors.New("Launcher fix already installed")
	case LauncherRetailUnverified:
		return nil, errors.New("Retail Launcher.exe: different Delphi build; not patched until offset mapping is verified")
	default:
		return nil, fmt.Errorf("unknown Launcher.exe SHA-256 %s", SHA(in))
	}
}

// ApplyMD5Steam reconstructs only a small verified code-injection delta,
// never embedding or replacing the game executable from a copyrighted copy.
func ApplyMD5Steam(in []byte) ([]byte, error) { return applyDelta(in, deltaRunMD5) }

// ApplySteamMD5Options adds *experimental* Boolean and MP cleanup options
// to the verified MD5-only patched Steam build. This does not fix Winsock 10013.
func ApplySteamMD5Options(in []byte, useBool, useMP bool) ([]byte, error) {
	if SHA(in) != SteamMD5 {
		return nil, fmt.Errorf("options require exact MD5-only build SHA-256 %s", SteamMD5)
	}
	if !useBool && !useMP {
		return append([]byte(nil), in...), nil
	}
	// .agfix is a 512-byte added section in the verified Steam MD5 patch.
	// 0x000..0x0f9: verified MD5-normalizer. Unused 0x100..0x1ff: helpers.
	if len(in) != 8124416 {
		return nil, errors.New("unexpected MD5 build length")
	}
	out := append([]byte(nil), in...)
	secOff := len(out) - 0x200
	hdr := 0x388 // 11th section header, verified as '.agfix' in the known build.
	if string(out[hdr:hdr+6]) != ".agfix" || binary.LittleEndian.Uint32(out[hdr+16:]) != 0x200 || int(binary.LittleEndian.Uint32(out[hdr+20:])) != secOff {
		return nil, errors.New("unknown .agfix section layout")
	}
	for _, v := range out[secOff+0x100:] {
		if v != 0 {
			return nil, errors.New("helper area not empty")
		}
	}
	va := 0x400000 + int(binary.LittleEndian.Uint32(out[hdr+12:]))
	if useMP {
		if string(out[steamMpClearOffset:steamMpClearOffset+5]) != string(bfromhex("33 c0 89 43 20")) {
			return nil, errors.New("MP opcode mismatch")
		}
		copy(out[secOff+0x100:], buildMpHelper())
		out[steamMpClearOffset] = 0xe8
		copy(out[steamMpClearOffset+1:], relative(steamMpClearVA, va+0x100))
	}
	if useBool {
		helperVA := va + 0x120
		copy(out[secOff+0x120:], buildBoolHelper(helperVA))
		for _, off := range boolOperands {
			if in[off-1] != 0xe8 {
				return nil, fmt.Errorf("invalid Boolean opcode at %x", off)
			}
			disp := int(int32(binary.LittleEndian.Uint32(in[off:])))
			oldInsn := off + 0x400c00 - 1
			if oldInsn+5+disp != steamVarFromBool {
				return nil, fmt.Errorf("unexpected Boolean target at 0x%x", off)
			}
			copy(out[off:], relative(oldInsn, helperVA))
		}
	}
	binary.LittleEndian.PutUint32(out[hdr+8:], 0x180) // virtual size; safely within 0x200 raw
	pe := int(binary.LittleEndian.Uint32(out[0x3c:]))
	opt := pe + 24
	checkOff := opt + 64
	binary.LittleEndian.PutUint32(out[checkOff:], 0)
	binary.LittleEndian.PutUint32(out[checkOff:], checksumPE(out, checkOff))
	return out, nil
}

const (
	md5BoolHash = "2d4eadeea7488bed9e25c3ce276a7216f68502d920db944ed80ac3f7e6c40c68"
	md5MpHash   = "01c40599b50ee7f5a275c2724e7fba55adbb8d6dd46c7d49c441362397bb6f18"
	md5BothHash = "d5597ca5886664394b97eeb05a97ae3e4feea19e61ccd070a3c6efa17123e750"
)
