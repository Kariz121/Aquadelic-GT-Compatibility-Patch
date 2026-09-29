# Aquadelic GT Compatibility Patch

Unofficial compatibility patch for **Aquadelic GT** focused on fixing crashes and initialization problems on modern Windows while keeping the original game behavior intact.

The patcher modifies the user's existing game executable.  
**No original Aquadelic GT executable is distributed with this project.**

## What this patch fixes

The current patch includes fixes for several issues found in the original game executable:

- fixes a double-free crash related to `TCameraLifter` / `TObjectList`
- restores missing multiplayer initialization
- fixes a multiplayer lobby/session timer issue
- fixes COM/XML initialization that can cause `Runtime Error 24` on newer or clean Windows installations

The goal of the project is to keep changes to the original executable as small as possible and fix the underlying causes rather than simply suppressing errors.

## Supported Windows versions

The patcher is designed as a 32-bit Win32 application for broad compatibility with:

- Windows XP
- Windows 7
- Windows 8 / 8.1
- Windows 10
- Windows 11

It is intended to work on both 32-bit Windows and 64-bit Windows through WOW64.

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

Different language releases, retail revisions or the Steam version may use a different executable and therefore may not currently be supported.

## Installation

1. Close Aquadelic GT if it is running.
2. Download `Aquadelic_GT_Patcher.exe` from the latest GitHub Release.
3. Run the patcher.
4. The patcher will try to locate `Run.exe` automatically.
5. If the game executable cannot be found automatically, you will be asked to select it manually.
6. If the executable is supported, the patch will be applied automatically.

Before modifying the game, the patcher creates a backup in:

```text
Original files before patch
```

If the executable has already been patched, the patcher will detect this and leave it unchanged.

## Unsupported versions

If the patcher reports that your executable is not compatible, you may have a different release, language edition, Steam version or game revision.

Please open an issue and include:

- your Aquadelic GT release / language
- where the game was obtained (for example retail or Steam)
- the SHA-256 hash of your original executable
- your Windows version
- the exact error or message shown by the patcher

Please **do not upload copyrighted game executables publicly** to GitHub issues.

If another executable revision needs to be investigated, contact the maintainer for further instructions.

## What the patcher changes

The patcher uses a binary-difference approach.

It does **not** contain or distribute the complete original or modified `Run.exe`. It verifies the supported original executable and applies only the required byte changes.

The current supported patch changes **354 bytes across 55 regions** of the executable.

## Backup / restoring the original game

The original executable is backed up before patching.

To restore it manually:

1. Close the game.
2. Open the game's `Original files before patch` folder.
3. Copy the backed-up executable back into the main Aquadelic GT directory.

## Important notes

- This is an unofficial community compatibility patch.
- A legally obtained copy of Aquadelic GT is required.
- This project is not affiliated with or endorsed by the original developer or publisher.
- Aquadelic GT and all original game assets remain the property of their respective rights holders.
- Use the patch at your own risk.

## Reporting problems

When reporting a problem, please include:

- Windows version
- game release / language if known
- whether the game is retail or Steam
- SHA-256 of the original `Run.exe`
- exact error message
- whether singleplayer and multiplayer are affected

## Credits

Compatibility investigation, reverse engineering and patch development were performed as a community preservation / compatibility effort for Aquadelic GT.
