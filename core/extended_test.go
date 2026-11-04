package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPatchVectors(t *testing.T) {
	p := os.Getenv("AQGT_TEST_DATA")
	if p == "" {
		t.Skip("AQGT_TEST_DATA required")
	}
	orig, err := os.ReadFile(filepath.Join(p, "AquadelicGT-Run-original-for-patch.exe"))
	if err != nil {
		t.Fatal(err)
	}
	md5, err := ApplyMD5Steam(orig)
	if err != nil {
		t.Fatal(err)
	}
	if SHA(md5) != SteamMD5 {
		t.Fatal("Steam MD5 output mismatch")
	}
	for _, pair := range [][2]bool{{true, false}, {false, true}, {true, true}} {
		out, err := ApplySteamMD5Options(md5, pair[0], pair[1])
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("MD5 combo bool=%v mp=%v sha=%s", pair[0], pair[1], SHA(out))
	}
	launch, err := os.ReadFile(filepath.Join(p, "Launcher(1).exe"))
	if err != nil {
		t.Fatal(err)
	}
	patched, err := ApplyLauncher140(launch)
	if err != nil {
		t.Fatal(err)
	}
	if SHA(patched) != Launcher140Full {
		t.Fatal("launcher output mismatch")
	}
	retail, err := os.ReadFile(filepath.Join(p, "RunOrginal(1).exe"))
	if err != nil {
		t.Fatal(err)
	}
	full, err := ApplyRetail(retail)
	if err != nil {
		t.Fatal(err)
	}
	if SHA(full) != RetailPatched {
		t.Fatal("Retail output mismatch")
	}
}
