package core

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var BloomFiles = []string{
	"Generic/Effects/PostProcessing.xml",
	"Russia/effects/PostProcessing.xml",
	"Greece/effects/PostProcessing.xml",
	"Caribbean/effects/PostProcessing.xml",
}
var rxPass = regexp.MustCompile(`(?is)<Pass\b[^>]*>.*?</Pass\s*>`)
var rxFrag = regexp.MustCompile(`(?is)<FragmentProgram\s*>\s*([^<>]+?)\s*</FragmentProgram\s*>`)
var rxBlur = regexp.MustCompile(`(?is)<Uniformf\b[^>]*\bName\s*=\s*["']blurVec["'][^>]*/>`)
var rxX = regexp.MustCompile(`(?i)\bX\s*=\s*(["'])([^"']+)['"]`)
var rxY = regexp.MustCompile(`(?i)\bY\s*=\s*(["'])([^"']+)['"]`)
var rxName = regexp.MustCompile(`(?i)\bName\s*=\s*["']bloom["']`)

func replaceXmlAttr(tag []byte, rx *regexp.Regexp, val string) ([]byte, error) {
	m := rx.FindSubmatchIndex(tag)
	if m == nil {
		return nil, fmt.Errorf("blurVec missing X/Y attribute: %q", tag)
	}
	return append(append([]byte(nil), tag[:m[4]]...), append([]byte(val), tag[m[5]:]...)...), nil
}
func decodeXY(tag []byte) (string, string, error) {
	x, y := rxX.FindSubmatch(tag), rxY.FindSubmatch(tag)
	if x == nil || y == nil {
		return "", "", fmt.Errorf("missing blurVec X/Y in %q", tag)
	}
	return string(x[2]), string(y[2]), nil
}

// PatchBloomXML updates only the two standard bloom passes, never mini camera blur.
// It refuses unknown values and returns the exact input when it is already fixed.
func validateBloomXML(data []byte) error {
	// Game files declare windows-1250 but reference files use ASCII-only markup.
	// Do not misinterpret arbitrary non-ASCII input when no charset decoder exists.
	for _, v := range data {
		if v >= 128 {
			return fmt.Errorf("non-ASCII XML with windows-1250 cannot be verified safely")
		}
	}
	copyData := bytes.ReplaceAll(data, []byte(`encoding="windows-1250"`), []byte(`encoding="utf-8"`))
	dec := xml.NewDecoder(bytes.NewReader(copyData))
	for {
		_, e := dec.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return fmt.Errorf("invalid XML: %w", e)
		}
	}
	return nil
}

func PatchBloomXML(data []byte) ([]byte, error) {
	if err := validateBloomXML(data); err != nil {
		return nil, err
	}
	/* dec:=xml.NewDecoder(bytes.NewReader(data))
	for {_,e:=dec.Token();if e==io.EOF {break};if e!=nil{return nil,fmt.Errorf("invalid XML: %w",e)}} */
	passes := rxPass.FindAllIndex(data, -1)
	output := append([]byte(nil), data...)
	type change struct {
		a, b int
		v    []byte
	}
	var changes []change
	found := map[string]bool{}
	for _, pos := range passes {
		p := data[pos[0]:pos[1]]
		headerEnd := bytes.IndexByte(p, '>')
		if headerEnd < 0 || !rxName.Match(p[:headerEnd+1]) {
			continue
		}
		fm := rxFrag.FindSubmatch(p)
		if len(fm) < 2 {
			continue
		}
		shader := strings.TrimSpace(string(fm[1]))
		if shader != "Blur.fs" && shader != "Bloom.fs" {
			continue
		}
		if found[shader] {
			return nil, fmt.Errorf("multiple standard bloom passes for %s", shader)
		}
		found[shader] = true
		hits := rxBlur.FindAllIndex(p, -1)
		if len(hits) != 1 {
			return nil, fmt.Errorf("%s has %d blurVec uniforms", shader, len(hits))
		}
		u := hits[0]
		tag := p[u[0]:u[1]]
		x, y, e := decodeXY(tag)
		if e != nil {
			return nil, e
		}
		wantX, wantY := "0", "0.001"
		oldX, oldY := "0", "0.003"
		if shader == "Bloom.fs" {
			wantX, wantY = "0.0005", "0"
			oldX, oldY = "0.003", "0"
		}
		if x != wantX || y != wantY {
			if x != oldX || y != oldY {
				return nil, fmt.Errorf("%s has custom blurVec X=%s Y=%s; not overwriting", shader, x, y)
			}
			t, e := replaceXmlAttr(tag, rxX, wantX)
			if e != nil {
				return nil, e
			}
			t, e = replaceXmlAttr(t, rxY, wantY)
			if e != nil {
				return nil, e
			}
			changes = append(changes, change{pos[0] + u[0], pos[0] + u[1], t})
		}
	}
	if !found["Blur.fs"] || !found["Bloom.fs"] {
		return nil, fmt.Errorf("missing standard bloom Blur/Bloom passes")
	}
	for i := len(changes) - 1; i >= 0; i-- {
		r := changes[i]
		output = append(append(append([]byte(nil), output[:r.a]...), r.v...), output[r.b:]...)
	}
	if err := validateBloomXML(output); err != nil {
		return nil, err
	}
	return output, nil
}
