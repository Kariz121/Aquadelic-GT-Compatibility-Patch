# Building the patcher

The source is written as a small Win32 C application.

## Requirements

A MinGW-w64 toolchain capable of producing 32-bit Windows executables.

The build script expects:

```text
i686-w64-mingw32-gcc
```

## Build

From the repository root, run:

```bat
build.bat
```

The resulting executable will be written to:

```text
build\Aquadelic_GT_Patcher.exe
```

The source targets Windows XP / Server 2003-era Win32 APIs (`WINVER=0x0501`)
so that the resulting 32-bit executable can also run on later 32-bit Windows
and under WOW64 on later 64-bit Windows.

## Important

The repository does not contain the complete original or modified Aquadelic GT
`Run.exe`.

The patch table contains only the replacement bytes required by the supported
compatibility patch. Before changing anything, the patcher identifies the
supported executable by SHA-256.
