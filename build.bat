@echo off
setlocal EnableExtensions
cd /d "%~dp0"

set "EXPECTED_WINDOWS=30334e6f104c2a2a5272989268896890ee6abbffd024d50e8ef115190b97018a"
set "EXPECTED_LINUX=dc0690eaea682a53609097e7df6aa8050ede4d1b62650bf6b203cd1a78df2db9"
set "WINDOWS_FILE=dist\Aquadelic_GT_Patcher_v1.1.0_Windows.exe"
set "LINUX_FILE=dist\Aquadelic_GT_Patcher_v1.1.0_Linux_x64"

where go.exe >nul 2>&1
if errorlevel 1 goto :nogo
set "GOTOOLCHAIN=local"
set "GOFLAGS="
set "GOEXPERIMENT="
for /f "tokens=3" %%V in ('go version') do set "GO_VERSION=%%V"
if /I not "%GO_VERSION%"=="go1.23.2" goto :wrongversion

set "CGO_ENABLED=0"
echo [1/4] Testing source...
go test ./...
if errorlevel 1 goto :failed
go vet ./...
if errorlevel 1 goto :failed

if not exist dist mkdir dist
if errorlevel 1 goto :failed

echo [2/4] Building Windows x86...
set "GOOS=windows"
set "GOARCH=386"
set "GO386=sse2"
go build -buildvcs=false -trimpath -ldflags="-s -w -H windowsgui" -o "%WINDOWS_FILE%" ./cmd/aqgt
if errorlevel 1 goto :failed

echo [3/4] Building Linux x64...
set "GOOS=linux"
set "GOARCH=amd64"
set "GOAMD64=v1"
go build -buildvcs=false -trimpath -ldflags="-s -w" -o "%LINUX_FILE%" ./cmd/aqgt
if errorlevel 1 goto :failed

echo [4/4] Verifying SHA-256...
for /f "delims=" %%H in ('powershell.exe -NoProfile -NonInteractive -Command "(Get-FileHash -LiteralPath '%WINDOWS_FILE%' -Algorithm SHA256).Hash"') do set "WINDOWS_HASH=%%H"
for /f "delims=" %%H in ('powershell.exe -NoProfile -NonInteractive -Command "(Get-FileHash -LiteralPath '%LINUX_FILE%' -Algorithm SHA256).Hash"') do set "LINUX_HASH=%%H"
if not defined WINDOWS_HASH goto :failed
if not defined LINUX_HASH goto :failed
if /I not "%WINDOWS_HASH%"=="%EXPECTED_WINDOWS%" goto :mismatch
if /I not "%LINUX_HASH%"=="%EXPECTED_LINUX%" goto :mismatch

echo.
echo SUCCESS: Both executables match the published SHA-256 hashes.
echo Output: dist\
goto :done

:nogo
echo ERROR: Go was not found. See BUILDING.md.
goto :error
:wrongversion
echo ERROR: Go 1.23.2 is required for a reproducible v1.1.0 build.
echo Installed: %GO_VERSION%
goto :error
:mismatch
echo ERROR: Built file SHA-256 differs from the published release.
echo Windows: %WINDOWS_HASH%
echo Linux:   %LINUX_HASH%
goto :error
:failed
echo ERROR: Test, build or checksum verification failed.
:error
pause
exit /b 1
:done
pause
exit /b 0
