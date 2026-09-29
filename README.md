# Aquadelic GT Compatibility Patch

## Download

**[Download the latest release](https://github.com/Kariz121/Aquadelic-GT-Compatibility-Patch/releases/latest)**

Unofficial compatibility and stability patch for **Aquadelic GT**.

The project focuses on fixing real bugs in the original executable while keeping the original gameplay, assets and game logic as close to the original release as possible.

The patcher modifies the user's existing executable.  
**No complete original or modified Aquadelic GT executable is distributed with this project.**

## What this patch fixes

The final patch is the result of several rounds of reverse engineering, cleanup and regression testing.

### Startup and legacy initialization

- restores required legacy Delphi/unit initialization that the original executable could fail to perform correctly on newer systems
- preserves the clean PE `.itext` mapping required by the compatibility bootstrap
- initializes the Ogg/Vorbis callback table before streamed audio uses it
- improves legacy XML/MSXML initialization behavior

### Profile and game-data compatibility

- fixes saving several Boolean values in `Profiles.xml` in a representation the game can reliably read back
- fixes the play-time display calculation so hours and minutes are calculated correctly

### Scene cleanup / Runtime Error 2

- fixes the main double-free crash occurring while stopping or cleaning up a race/session
- the crash was traced through `TSceneNode` / `TCameraLifter` destruction to conflicting `TObjectList` ownership behavior
- the final root fix removes the object from the owning list without allowing the list to free it prematurely
- old experimental `nil` guards, invalid-free suppression and exception-masking workarounds are not used as the final solution

### Multiplayer

- restores missing original multiplayer/network unit initialization
- fixes crashes caused by uninitialized multiplayer function pointers
- fixes a separate lobby/session cleanup bug where a timer callback could remain active after its session pointer had been cleared
- fixes the crash path that could occur when leaving a lobby while waiting for additional players

### COM / XML — Runtime Error 24

- fixes `Runtime Error 24` on clean/newer Windows installations
- the underlying HRESULT was identified as `0x800401F0 (CO_E_NOTINITIALIZED)`
- the final implementation initializes COM using the game's existing initialization path when XML first requires it
- COM initialization is performed once instead of repeatedly for every XML operation

## Verified after patching

The final executable has been tested with:

- game startup
- profile loading
- singleplayer
- entering and leaving multiplayer
- creating a LAN lobby
- discovering a LAN lobby from another machine
- joining with a second player
- changing boats
- ready-state handling
- starting a multiplayer race
- playing the race
- returning to the lobby
- leaving and re-entering multiplayer
- creating another multiplayer session
- normal game shutdown

LAN multiplayer was tested between a physical host and a separate Windows virtual machine on the same local network.

## What is intentionally NOT included

During development many diagnostic and experimental builds were created. The final patch does **not** keep temporary workarounds such as:

- generic invalid-free suppression
- broad `nil`-list guards used only during diagnosis
- disabling music or menu audio
- forcing an OpenAL default device
- direct-export OpenAL experiments
- diagnostic MessageBoxes and logging hooks
- the experimental 16:9/FHD GUI modifications

Those experiments were useful for finding the real causes, but they are not part of the public compatibility patch.

## Supported Windows versions

The patcher is designed as a 32-bit Win32 application targeting Windows XP-era APIs for broad compatibility with:

- Windows XP
- Windows 7
- Windows 8 / 8.1
- Windows 10
- Windows 11

It is intended to run on both 32-bit Windows and 64-bit Windows through WOW64.

## Supported game version

At the moment, only one specific original Aquadelic GT executable has been verified and is supported.

Original supported `Run.exe` SHA-256:

```text
36fcf10bd35f0ea3844b16d6ecc8ac67752ce878f7a40ace5192d91d2017d89a
```

Patched executable SHA-256:

```text
949768e738e43aa2240176b752ac2aad9d0b66dc442aa87fd21634b58d738aa5
```

The patcher checks the executable before changing anything.

If an unsupported executable is detected, **no changes are made**.

Different retail revisions, language editions or the Steam release may use a different executable and are not automatically assumed to be compatible.

## Installation

1. Close Aquadelic GT if it is running.
2. Download `Aquadelic_GT_Patcher.exe` from the latest GitHub Release.
3. Run the patcher.
4. The patcher will try to locate `Run.exe` automatically.
5. If the executable cannot be found automatically, you will be asked to select it manually.
6. The selected file is identified by its contents/SHA-256, not only by its filename.
7. If the executable is supported, the patch is applied automatically.

Before modifying the game, the patcher creates a backup in:

```text
Original files before patch
```

If the executable has already been patched, the patcher detects this and leaves it unchanged.

## Unsupported versions

If the patcher reports that your executable is not compatible, you may have a different release, language edition, Steam version or game revision.

Please open an issue and include:

- your Aquadelic GT release / language
- where the game was obtained (for example retail or Steam)
- the SHA-256 hash of your original executable
- your Windows version
- the exact message shown by the patcher

Please **do not upload copyrighted game executables publicly** to GitHub issues.

If another executable revision needs to be investigated, contact the maintainer for further instructions.

## How the patcher works

The public patcher uses a binary-difference approach.

It does **not** contain or distribute the complete original or modified `Run.exe`. It verifies the supported original executable and applies only the required byte changes.

The current public patch modifies **354 bytes across 55 regions** of the executable and verifies the SHA-256 of the finished file.

## Backup / restoring the original game

To restore the original executable manually:

1. Close the game.
2. Open the game's `Original files before patch` folder.
3. Copy the backed-up `Run.exe` back into the main Aquadelic GT directory.

## Known limitations

- only the currently verified executable revision is supported
- other retail/language/Steam executables require separate analysis before they can be patched safely
- the old launcher may still have trouble parsing some modern OpenGL version strings; this is separate from the fixes in the game executable
- widescreen/FHD GUI modifications are not part of this public patch

## Development history

A condensed history of the reverse-engineering process and the major development branches is available in [`PATCH_HISTORY.md`](PATCH_HISTORY.md).

## Source code

The patcher source is available in [`src/patcher.c`](src/patcher.c).

The repository contains the patching logic and replacement-byte table only. It does not contain the complete original or patched game executable.

Build instructions are available in [`BUILDING.md`](BUILDING.md).

## License

The original patcher source code and tooling in this repository are released under the MIT License. See [`LICENSE`](LICENSE).

Aquadelic GT itself, its executable, trademarks and game assets are not covered by this license and remain the property of their respective rights holders.

## Disclaimer

- This is an unofficial community compatibility patch.
- A legally obtained copy of Aquadelic GT is required.
- This project is not affiliated with or endorsed by the original developer or publisher.
- Use the patch at your own risk.
