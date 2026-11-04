package core

import (
	"os"
	"testing"
)

func TestBroadcastPayloadABI(t *testing.T) {
	b, err := broadcastPayload(0x800000)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) > 0x200 {
		t.Fatal("code too large")
	}
	if !BroadcastMarkerPresent(b) {
		t.Fatal("marker missing")
	}
}
func BroadcastMarkerPresent(b []byte) bool {
	for i := 0; i+len(BroadcastMarker) <= len(b); i++ {
		eq := true
		for j := range BroadcastMarker {
			if b[i+j] != BroadcastMarker[j] {
				eq = false
				break
			}
		}
		if eq {
			return true
		}
	}
	return false
}
func TestBroadcastKnownSteamBinaries(t *testing.T) {
	samples := []struct{ env, source, out string }{
		{"AQGT_ORIGINAL_EXE", SteamOriginal, "988ad60b0bb0d95c4c6130b379e478c612fc820cac963a2348a89d601666e1ce"},
		{"AQGT_MD5_EXE", SteamMD5, "4be02e05a3b59fcaf3120b8feffa636f7e9e07dba6c4e8d82976f3aed07a7ae4"},
	}
	for _, s := range samples {
		path := os.Getenv(s.env)
		if path == "" {
			t.Log("optional test fixture missing: " + s.env)
			continue
		}
		data, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if SHA(data) != s.source {
			t.Fatalf("%s source hash mismatch: %s", s.env, SHA(data))
		}
		out, e := PatchBroadcastV2(data)
		if e != nil {
			t.Fatal(e)
		}
		if SHA(out) != s.out {
			t.Fatalf("%s produced %s, expected %s", s.env, SHA(out), s.out)
		}
		if !BroadcastPatched(out) {
			t.Fatal("section verification failed")
		}
		if _, e = PatchBroadcastV2(out); e == nil {
			t.Fatal("double patch was not refused")
		}
	}
}
