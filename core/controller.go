package core

import (
	"bytes"
	"fmt"
	"regexp"
)

const ControllerOriginalSHA = "397f0754bfbef212807c3ea9fb45d910c8f47183c563cfbaedbae22a85fce56e"
const ControllerFinalSHA = "ae22945d4576a033523aa124989ca2548b38f973e65cebe382ea54924d385344"
const SteamControlsBaseSHA = "4be02e05a3b59fcaf3120b8feffa636f7e9e07dba6c4e8d82976f3aed07a7ae4"
const SteamControlsFinalSHA = "7283966d6ce1891cc2634514d525263689135803baba3fda297c17e5d8cc7409"
const SteamControlsNoMD5BaseSHA = "988ad60b0bb0d95c4c6130b379e478c612fc820cac963a2348a89d601666e1ce"
const SteamControlsNoMD5FinalSHA = "b61fb5d2ef4a365a13494286579203ef4a79568e2758def0a3c89b1d1b8c167c"

var irSettingRX = regexp.MustCompile(`(?s)<IRSetting\b[^>]*>.*?</IRSetting>`)

// For Steam 1.4 only: known original -> tested PowerTriggers + CameraStick mapping.
func PatchControllerXML(b []byte) ([]byte, error) {
	if SHA(b) != ControllerOriginalSHA {
		return nil, fmt.Errorf("ControllerSchemes.xml: unknown source hash %s", SHA(b))
	}
	matches := irSettingRX.FindAllIndex(b, -1)
	if len(matches) != 34 {
		return nil, fmt.Errorf("controller schema changed: %d entries", len(matches))
	}
	replacements := map[int]struct{ axis, reverse, caption string }{
		16: {"2", "True", "RS-U"},
		17: {"2", "False", "RS-D"},
		18: {"3", "True", "RS-L"},
		19: {"3", "False", "RS-R"},
		21: {"5", "False", "RT"},
		22: {"4", "False", "LT"},
	}
	out := append([]byte(nil), b...)
	for j := len(matches) - 1; j >= 0; j-- {
		r, ok := replacements[j]
		if !ok {
			continue
		}
		frag := string(b[matches[j][0]:matches[j][1]])
		axis := regexp.MustCompile(`AxisReverse="(?:True|False)"`)
		key := regexp.MustCompile(`<KeyCode>\d+</KeyCode>`)
		caption := regexp.MustCompile(`<Caption>[^<]+</Caption>`)
		if !axis.MatchString(frag) || !key.MatchString(frag) || !caption.MatchString(frag) {
			return nil, fmt.Errorf("malformed IRSetting %d", j)
		}
		frag = axis.ReplaceAllString(frag, `AxisReverse="`+r.reverse+`"`)
		frag = key.ReplaceAllString(frag, `<KeyCode>`+r.axis+`</KeyCode>`)
		frag = caption.ReplaceAllString(frag, `<Caption>`+r.caption+`</Caption>`)
		out = append(append(append([]byte(nil), out[:matches[j][0]]...), []byte(frag)...), out[matches[j][1]:]...)
	}
	if SHA(out) != ControllerFinalSHA {
		return nil, fmt.Errorf("controller XML target hash mismatch %s", SHA(out))
	}
	return out, nil
}
func PatchSteamControls(b []byte) ([]byte, error)      { return applyDelta(b, deltaControllerAll) }
func PatchSteamControlsNoMD5(b []byte) ([]byte, error) { return applyDelta(b, deltaControllerNoMD5) }
func ControllerReady(b []byte) bool {
	return SHA(b) == SteamControlsBaseSHA || SHA(b) == SteamControlsNoMD5BaseSHA
}
func ControllerPatched(b []byte) bool {
	return SHA(b) == SteamControlsFinalSHA || SHA(b) == SteamControlsNoMD5FinalSHA
}

var _ = bytes.Equal
