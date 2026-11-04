# Known issues — v1.1.0

## Installation and Restore

An unexpected shutdown during **Install** or **Change mode** may leave some game files modified before the backup manifest is completed. Automatic Restore can then fail, even if the backup files remain available. **Make a separate backup of your game folder before patching**, and never delete the patcher's backups after an error.

The original v1.1.0 build also has known security-hardening limitations involving its local browser interface and handling of specially crafted backup paths or symbolic links. Only select a game folder you trust, and use the interface's **Exit** action to close the patcher.

## Gameplay reports

- **Multiplayer:** Freezes have been reported on one CachyOS desktop configuration, including tests without the patch. A separate CachyOS laptop successfully joined an older Windows host. The cause remains unknown.
- **Audio:** `TSoundManager.Initialize: can't open preferred device` has been reported on Windows and Steam Deck, including unpatched installations. A connection to the chosen resolution has not been established.

These reports do **not** establish that the patch causes the problems, or that they affect all installations.

## Report a problem

Open a [GitHub Issue](https://github.com/Kariz121/Aquadelic-GT-Compatibility-Patch/issues) with your game version, operating system, hardware, Wine/Proton version (if relevant), steps to reproduce and any error messages. Please also say whether the issue occurs **without the patch**.

Do not upload copyrighted game executables, private data or unredacted personal information.
