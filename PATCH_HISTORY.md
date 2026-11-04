# Patch history

This document summarizes the major changes to the public compatibility patch. Detailed diagnostic and experimental builds are not part of the release.

## v1.1.0 — Cross-platform compatibility update

### Platform support

The patcher was rebuilt in Go to support Windows and Linux from one source tree. The Linux version also supports use with Proton on Steam Deck. It identifies supported game installations and checks file hashes before making changes.

The Steam 1.4 update includes platform-specific compatibility adjustments: launcher and Mesa compatibility, Bloom and language fixes, and improvements to UDP broadcast-based LAN discovery. Linux-specific changes cover Steam profile integrity, launcher integrity and certain fonts. Windows does not apply those Linux-only file changes. Support for the previously verified retail installation is retained.

### Controls and installation

Added three controller choices for supported Steam installations: Original, Boat Analog and Boat Step. Installation state and file backups are stored in the game directory. The patcher supports changing modes and restoring the original files after relaunching it. The local browser interface includes installation discovery and a folder picker.

### Verification

The Windows x86 and Linux x64 executables have been independently reproduced byte-for-byte from the published Go sources using Go 1.23.2. Tests were also conducted on the supported game versions, including multiplayer on different systems. Successful tests do not guarantee compatibility on every configuration.

### Security review and limitations

A source review found no confirmed Critical or High severity vulnerabilities within its stated threat model, but identified several medium and lower-severity issues. The main practical concern is recovery after an unexpected interruption during installation or a mode change; additional issues involve local HTTP validation and handling untrusted filesystem paths. These issues remain in the published v1.1.0 build and are summarized in [KNOWN_ISSUES.md](KNOWN_ISSUES.md).

Microsoft reviewed a submitted build and removed its associated antivirus detection. This is a determination about that submission, not a security certification or a guarantee for every security product.

*Build note:* The original executable retains an internal release-candidate label from development. The public release version is **v1.1.0**; keeping the original executable bytes allows the published checksums to remain reproducible.

## v1.0.0 — Initial Windows release

The first public version supported a specific verified retail installation and included fixes for game startup, legacy initialization, XML/COM behavior, profile data, play-time calculations and multiplayer cleanup. Compatibility fixes were applied only after checking the original executable's SHA-256. The original v1.0.0 release and its history remain accessible through the corresponding Git tag.
