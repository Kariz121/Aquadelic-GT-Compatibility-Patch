//go:build !windows

package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// On Linux, use an installed desktop environment's native folder picker.
// If neither utility exists, users can paste the folder into the text field.
func pickFolder() string {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if _, err := exec.LookPath("kdialog"); err == nil {
		cmd = exec.CommandContext(ctx, "kdialog", "--getexistingdirectory", os.Getenv("HOME"))
	} else if _, err := exec.LookPath("zenity"); err == nil {
		cmd = exec.CommandContext(ctx, "zenity", "--file-selection", "--directory", "--title=Select Aquadelic GT folder")
	}
	if cmd == nil {
		return ""
	}
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
