//go:build windows

package main

import (
	"runtime"
	"syscall"
	"unsafe"
)

// browseInfoW is the Win32 BROWSEINFOW structure. This invokes Windows'
// built-in folder chooser directly; no PowerShell, shell script, or child
// console window is involved.
type browseInfoW struct {
	owner        uintptr
	root         uintptr
	displayName  *uint16
	title        *uint16
	flags        uint32
	callback     uintptr
	callbackData uintptr
	image        int32
}

const (
	coinitApartmentThreaded = 0x2
	bifReturnOnlyFSDirs     = 0x0001
	bifNewDialogStyle       = 0x0040
)

var (
	shell32             = syscall.NewLazyDLL("shell32.dll")
	ole32               = syscall.NewLazyDLL("ole32.dll")
	shBrowseForFolder   = shell32.NewProc("SHBrowseForFolderW")
	shGetPathFromIDList = shell32.NewProc("SHGetPathFromIDListW")
	coInitializeEx      = ole32.NewProc("CoInitializeEx")
	coUninitialize      = ole32.NewProc("CoUninitialize")
	coTaskMemFree       = ole32.NewProc("CoTaskMemFree")
)

func pickFolder() string {
	// SHBrowseForFolderW's modern dialog style requires an STA thread.
	// Keep the goroutine on the same OS thread until COM is uninitialised.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hr, _, _ := coInitializeEx.Call(0, coinitApartmentThreaded)
	if hr != 0 && hr != 1 { // S_OK and S_FALSE both require CoUninitialize.
		return ""
	}
	defer coUninitialize.Call()

	var display [260]uint16
	title, err := syscall.UTF16PtrFromString("Select Aquadelic GT folder (containing Run.exe)")
	if err != nil {
		return ""
	}
	bi := browseInfoW{
		displayName: &display[0],
		title:       title,
		flags:       bifReturnOnlyFSDirs | bifNewDialogStyle,
	}
	pidl, _, _ := shBrowseForFolder.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 { // User pressed Cancel.
		return ""
	}
	defer coTaskMemFree.Call(pidl)

	// Win32 SHGetPathFromIDListW writes at most MAX_PATH UTF-16 code units.
	var result [260]uint16
	ok, _, _ := shGetPathFromIDList.Call(pidl, uintptr(unsafe.Pointer(&result[0])))
	if ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(result[:])
}
