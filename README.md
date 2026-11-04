# Aquadelic GT Compatibility Patch

Unofficial compatibility patch for **Aquadelic GT** on Windows, Linux/Proton and Steam Deck. It updates files in your existing game installation; no complete game executables are distributed.

## Download

[Download the latest release](https://github.com/Kariz121/Aquadelic-GT-Compatibility-Patch/releases)

## What's new in v1.1.0

- Support for verified Steam 1.4 and retail game installations.
- Windows and Linux/Steam Deck compatibility improvements, including LAN discovery, graphics and language fixes.
- Optional Steam controller modes: **Original**, **Boat Analog** and **Boat Step**.
- File verification, backups, controller-mode switching and Restore.

## Installation

1. Close the game and its launcher.
2. Download the patcher for your operating system from **Releases**.
3. **Windows:** run the `.exe`. **Linux:** make the downloaded file executable and run it from Desktop Mode.
4. Open the local interface, select the game folder if needed, and choose **Check files**.
5. Select the preferred controller mode, then install. Use **Restore original files** to undo changes.

Only known game builds are modified. **Back up your game folder before installing**, and keep the patcher's backups until you have confirmed everything works.

## Known limitations

An interrupted installation may require manual recovery. Multiplayer freezes have been reported on some Linux configurations, and an intermittent OpenAL audio error has been reported on multiple systems, including without the patch. These issues are not yet fully understood.

Read [Known issues](KNOWN_ISSUES.md) before use. **Bug reports are welcome**, including small problems and successful compatibility tests. Please use [GitHub Issues](https://github.com/Kariz121/Aquadelic-GT-Compatibility-Patch/issues).

## Source code

The v1.1.0 implementation is in `cmd/`, `core/` and `engine/`. Build instructions and verified checksums are in [BUILDING.md](BUILDING.md). Earlier releases remain available through their GitHub tags.

Further development notes are in [PATCH_HISTORY.md](PATCH_HISTORY.md).

## Disclaimer

This is an independent community project, not affiliated with the game developer or publisher. A legally obtained copy of the game is required. Use at your own risk.
