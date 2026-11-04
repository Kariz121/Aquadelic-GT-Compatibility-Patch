#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

if ! command -v go >/dev/null 2>&1; then
  echo "ERROR: Go not found. See BUILDING.md" >&2
  exit 1
fi

export GOTOOLCHAIN=local CGO_ENABLED=0
unset GOFLAGS GOEXPERIMENT || true
if [[ "$(go version | awk '{print $3}')" != "go1.23.2" ]]; then
  echo "ERROR: Go 1.23.2 is required for reproducible v1.1.0 builds." >&2
  exit 1
fi

echo '[1/4] Testing source...'
go test ./...
go vet ./...
mkdir -p dist

echo '[2/4] Building Windows x86...'
CGO_ENABLED=0 GOOS=windows GOARCH=386 GO386=sse2 \
  go build -buildvcs=false -trimpath -ldflags='-s -w -H windowsgui' \
  -o dist/Aquadelic_GT_Patcher_v1.1.0_Windows.exe ./cmd/aqgt

echo '[3/4] Building Linux x64...'
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 \
  go build -buildvcs=false -trimpath -ldflags='-s -w' \
  -o dist/Aquadelic_GT_Patcher_v1.1.0_Linux_x64 ./cmd/aqgt

echo '[4/4] Verifying SHA-256...'
printf '%s  %s\n' \
  '30334e6f104c2a2a5272989268896890ee6abbffd024d50e8ef115190b97018a' \
  'dist/Aquadelic_GT_Patcher_v1.1.0_Windows.exe' \
  'dc0690eaea682a53609097e7df6aa8050ede4d1b62650bf6b203cd1a78df2db9' \
  'dist/Aquadelic_GT_Patcher_v1.1.0_Linux_x64' | sha256sum --check --status

echo 'SUCCESS: Both executables match the published SHA-256 hashes.'
