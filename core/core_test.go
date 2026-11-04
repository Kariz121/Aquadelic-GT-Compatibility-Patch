package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReferenceExecutables(t *testing.T) {
	p := os.Getenv("AQGT_TEST_DATA")
	if p == "" {
		t.Skip("Set AQGT_TEST_DATA to directory with private original EXEs")
	}
	retail, e := os.ReadFile(filepath.Join(p, "RunOrginal(1).exe"))
	if e != nil {
		t.Fatal(e)
	}
	if Detect(retail) != RetailClean {
		t.Fatal("wrong retail hash")
	}
	patched, e := ApplyRetail(retail)
	if e != nil {
		t.Fatal(e)
	}
	if Detect(patched) != RetailDone {
		t.Fatal("retail binary mismatch")
	}
	steam, e := os.ReadFile(filepath.Join(p, "Run(6).exe"))
	if e != nil {
		t.Fatal(e)
	}
	if Detect(steam) != SteamClean {
		t.Fatal("wrong Steam hash")
	}
	for _, mode := range []Mode{Bool, Multiplayer, Both} {
		b, e := BuildSteam(steam, mode)
		if e != nil {
			t.Fatalf("%s: %v", mode, e)
		}
		if SHA(b) != mode.Hash() {
			t.Fatalf("%s incorrect result SHA", mode)
		}
		if _, e := BuildSteam(b, mode); e == nil {
			t.Fatal("accepted already modified executable")
		}
	}
}
