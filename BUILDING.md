# Building from source

The Windows and Linux executables are built from the **same Go source code**. For a byte-for-byte match with the published v1.1.0 binaries, install **Go 1.23.2** from [go.dev](https://go.dev/dl/). Other Go versions may produce different files.

## Quick build

- **On Windows:** run `build.bat`.
- **On Linux:** run `./build.sh` (or `bash build.sh`).

Both scripts run `go test ./...` and `go vet ./...`, build the Windows x86 and Linux x64 versions into `dist/`, and verify their SHA-256 hashes.

## Expected SHA-256

| Target | SHA-256 |
|---|---|
| Windows x86 | `30334e6f104c2a2a5272989268896890ee6abbffd024d50e8ef115190b97018a` |
| Linux x64 | `dc0690eaea682a53609097e7df6aa8050ede4d1b62650bf6b203cd1a78df2db9` |

The downloadable asset filenames do not affect their contents or hashes. Windows builds use `GOOS=windows GOARCH=386`; Linux builds use `GOOS=linux GOARCH=amd64`. Both use `CGO_ENABLED=0` and the flags recorded in the build scripts.

Go 1.23.2 is retained solely to reproduce this historical release; it is not a currently supported toolchain. Keep antivirus protection enabled when testing locally built executables.
