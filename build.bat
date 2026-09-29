@echo off
setlocal

where i686-w64-mingw32-gcc >nul 2>nul
if errorlevel 1 (
    echo i686-w64-mingw32-gcc was not found in PATH.
    echo Install a MinGW-w64 toolchain with 32-bit support and try again.
    exit /b 1
)

if not exist build mkdir build

i686-w64-mingw32-gcc ^
  -std=c99 ^
  -Os ^
  -s ^
  -mwindows ^
  -DWINVER=0x0501 ^
  -D_WIN32_WINNT=0x0501 ^
  -o build\Aquadelic_GT_Patcher.exe ^
  src\patcher.c ^
  -luser32 -lcomdlg32 -ladvapi32

if errorlevel 1 (
    echo.
    echo Build failed.
    exit /b 1
)

echo.
echo Built: build\Aquadelic_GT_Patcher.exe
echo Target: 32-bit Win32 / Windows XP 5.1+
