//go:build !windows

package engine

import "os"

func swapFile(src, dst string) error { return os.Rename(src, dst) }
