//go:build windows

package engine

import (
	"fmt"
	"syscall"
	"unsafe"
)

var movefileex = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW")

func swapFile(src, dst string) error {
	a, e := syscall.UTF16PtrFromString(src)
	if e != nil {
		return e
	}
	b, e := syscall.UTF16PtrFromString(dst)
	if e != nil {
		return e
	}
	r, _, err := movefileex.Call(uintptr(unsafe.Pointer(a)), uintptr(unsafe.Pointer(b)), uintptr(1|8))
	if r == 0 {
		return fmt.Errorf("MoveFileExW failed: %w", err)
	}
	return nil
}
