# Development and Patch History

This document explains how the Aquadelic GT Compatibility Patch developed, which problems each release addresses, and why the changes were made. It covers the investigation and testing behind the published patches, not just the final list of features.

The project is an **unofficial compatibility patch**. It modifies selected files in an existing, legally obtained game installation. It does **not** distribute a complete original or modified copy of the game.

**Version numbers below refer to the patcher, not the game.** The currently supported game executables include the verified Polish retail 1.2 release and Steam 1.4. These are different game builds and require different patches.

## Project goals

Aquadelic GT was designed around older Windows components, graphics drivers, input devices and game-engine behavior. Some faults are only visible on newer Windows versions; others become apparent when the Windows game runs through Wine or Proton on Linux and Steam Deck.

The compatibility work has followed four principles:

1. **Fix the specific cause rather than hide the error.** A patch should correct a verified code path or file value, not suppress every exception that happens to occur nearby.
2. **Preserve the original game.** Gameplay, content and unrelated engine behavior should remain as close as possible to the original release.
3. **Treat game editions separately.** A fix for one `Run.exe`, launcher or operating system is not automatically appropriate for another.
4. **Make changes verifiable and reversible.** Recognize known files, validate results, create backups and allow restoration. Unknown or previously modified files should not be patched blindly.

The two public releases represent different stages of that work: v1.0.0 established the initial Windows compatibility fixes for a verified retail edition; v1.1.0 expanded the project into a platform-aware patcher for retail and Steam installations.

---

## v1.0.0 — Initial Windows compatibility release

The original release focused on a particular Windows retail executable. Investigation combined crash reproduction, analysis of the compiled game, examination of Delphi runtime behavior and comparisons of original and patched instructions. The resulting patch changed a small number of verified byte ranges instead of replacing the game executable.

### Legacy startup and component initialization

**Problem:** On newer systems, some engine components could be used before their required initialization had completed. The failures were not necessarily caused by missing game files: the required initialization path was present but could be skipped or reached in an unsuitable order.

**Change:** The patch restores necessary startup and unit-initialization behavior, including initialization used by the game's audio streaming and multiplayer systems. One investigated case concerned the Ogg/Vorbis callback table being needed before streamed audio accessed it.

**Why:** Ensuring that engine components are initialized before use addresses the cause of the crash. It avoids disabling audio or introducing a broad exception handler as a substitute for correct initialization.

### XML and COM — Runtime Error 24

**Problem:** The game could report **Runtime Error 24** when XML-related functionality ran on a clean or newer Windows installation. Debugging linked the error to COM being used before the required initialization (`CO_E_NOTINITIALIZED`, HRESULT `0x800401F0`).

**Change:** The patch initializes COM through the game's existing initialization flow when the XML component first requires it. Initialization is performed once for the relevant path rather than repeatedly around every XML operation.

**Why:** This fixes the missing prerequisite while retaining the original XML functionality. The aim is to make the engine's existing behavior work reliably, not to replace its XML subsystem.

### Profile data and play-time calculations

**Problem:** Some Boolean values saved in `Profiles.xml` were written in a representation that the game did not reliably read back. Separately, the display of accumulated play time could calculate hours and minutes incorrectly.

**Change:** The first release includes targeted corrections to profile-value handling and the play-time calculation.

**Why:** These are data-compatibility problems: the game should be able to read information it previously wrote and display recorded time consistently. The patch does not introduce a new profile format or replace players' save files.

### Scene cleanup — Runtime Error 2

**Problem:** Stopping or cleaning up a race could lead to **Runtime Error 2** or a crash associated with freeing an object twice. The investigated path involved scene/camera teardown and conflicting ownership between `TCameraLifter`, `TSceneNode` and a `TObjectList`.

**Change:** The final correction addresses the ownership conflict so the affected object is not freed again by the owning list.

**Why:** Removing the second deletion is safer than ignoring invalid-free exceptions. Earlier diagnostic ideas such as general invalid-free suppression and broad null guards were not adopted as the final fix.

### Multiplayer startup and lobby cleanup

**Problem:** Multiplayer could fail because parts of the networking system were not initialized when called. A separate failure could occur when leaving a lobby: a timer callback might still reference a session object after the session pointer had been cleared.

**Change:** The patch restores required multiplayer initialization and corrects the affected lobby/session timer cleanup path.

**Why:** A clean transition into and out of multiplayer should not depend on uninitialized function pointers or a callback that outlives its session. These corrections target stability; they are not a replacement for the game's network protocol.

### First-release patching approach

The v1.0.0 patcher applied a fixed set of changes **only to a known retail executable**. The release documentation recorded **55 modified regions totaling 354 bytes**, with checks of the original and resulting game-file SHA-256. The patcher did not bundle the full game executable.

The initial release was designed for legacy Windows compatibility. It did not yet provide the integrated Steam 1.4 feature set, Linux-specific adjustments, controller-mode selection or the newer cross-platform install/restore workflow.

**Historical source note:** Later readable C sources were added to the repository as a functional implementation of the same retail patching logic. They should not be mistaken for a byte-for-byte reproduction of the small original v1.0.0 executable, which was built from an earlier, separate C/assembly implementation. The original release remains preserved under its Git tag, and the later source history is retained on the `legacy-v1.0.0` branch. By contrast, the v1.1.0 Windows and Linux binaries have been reproduced byte-for-byte from their published Go sources.

---

## Why v1.1.0 required a new approach

The Steam 1.4 game is **not** the same executable as the earlier retail release. Its code layout, resources and checksums differ. Reusing addresses or binary changes from the retail edition would risk modifying unrelated instructions.

Linux and Steam Deck introduce another distinction: the **game itself remains a Windows program running through Proton/Wine**, while the patcher can run as a native Linux application. A native Linux patcher is not a native Linux port of Aquadelic GT.

The development work therefore moved from a single-edition Windows patcher to a system that can:

- Detect which verified game edition is installed.
- Select the correct changes for Windows or Linux/Proton.
- Apply multiple related fixes in a controlled order.
- Keep unsupported or already-modified files untouched.
- Preserve original file bytes for restoration and optional controller-mode changes.

Go was chosen for the integrated patcher so the same main source tree could produce a Windows x86 GUI executable and a Linux x86-64 executable, with small platform-specific implementations where operating-system behavior differs.

## v1.1.0 — Steam 1.4 and cross-platform compatibility work

### 1. Steam profile integrity on Linux and Proton

**Observed failure:** The Steam edition could report messages such as **“Game state file was corrupted”** or **“List index out of bounds (0)”** while loading or saving profile data through Proton. The behavior was investigated by comparing profile XML, saved integrity values, buffer contents and the game's checksum routines.

**Investigation:** Profile integrity involves the XML representation of the data and an MD5-based check. Differences in how the game receives or serializes XML under the compatibility layer were a significant lead. The precise underlying Wine/MSXML difference was not fully established in every environment; changing text formatting alone was not a reliable fix.

**Final changes:** For the verified Steam 1.4 installation **on Linux/Proton**, v1.1.0 applies a targeted `Run.exe` integrity correction together with the corresponding launcher integrity adjustment.

**Why Linux only:** The same changes were unnecessary for the tested Windows path and could interfere with profile behavior there. The Windows build deliberately **does not** apply the Steam MD5/integrity modifications to either `Run.exe` or the launcher.

The patcher does **not** rewrite `Profiles.xml` or automatically convert saves from one game edition to another. Save migration was used separately during testing, but it is not a published feature.

### 2. Launcher compatibility with modern OpenGL and Mesa

**Observed failure:** The older launcher made assumptions about the format of OpenGL version information. Modern Mesa/OpenGL driver strings could cause incorrect parsing or launcher compatibility problems under Linux/Proton.

**Final changes:** The supported Steam launcher receives a verified correction to its OpenGL/Mesa version handling. On Linux it also receives the profile-integrity adjustment described above. On Windows it receives **only** the launcher compatibility change, not the integrity modification.

**Why:** The launcher and the running game have different responsibilities. A fix that is needed to read profile integrity under Proton should not be carried into a Windows executable merely because it exists in another build.

### 3. Bloom and post-processing configuration

**Observed failure:** Several game environments used inconsistent values for the blur vectors in the standard Bloom passes. This affected the appearance of post-processing on some graphics setups. Separately, hardware/driver combinations were observed to produce Bloom artifacts under Linux; those cases helped guide testing but did not establish a universal driver defect.

**Final changes:** v1.1.0 adjusts only the known **horizontal and vertical Bloom passes** in four verified `PostProcessing.xml` files:

| Resource | Purpose of the change |
| --- | --- |
| `Generic/Effects/PostProcessing.xml` | Standard effects configuration |
| `Russia/effects/PostProcessing.xml` | Russia environment |
| `Greece/effects/PostProcessing.xml` | Greece environment |
| `Caribbean/effects/PostProcessing.xml` | Caribbean environment |

The verified values are `X=0, Y=0.001` for the vertical `Blur.fs` pass and `X=0.0005, Y=0` for the horizontal `Bloom.fs` pass, replacing the known older values of approximately `0.003` on the affected axis.

**Why:** The correction is narrow. It does not change unrelated camera effects, Nitro effects, or the user's general graphics settings. Unexpected XML structures and unknown values are rejected rather than overwritten.

Using **Zink** as an alternative OpenGL implementation helped with a separate Radeon/Linux artifact report. Zink is **not bundled or forced** by the patcher, and it is not presented as a guaranteed solution for every graphics driver.

### 4. UDP broadcast and LAN multiplayer discovery

**Observed failure:** On Proton, creating or discovering a LAN lobby could fail while the game attempted to send a UDP broadcast packet. Debugging identified a Windows sockets error **10013** on the broadcast send path when the relevant socket option had not been enabled.

**Final change:** The Steam 1.4 compatibility fix enables the socket option **`SO_BROADCAST`** before the affected UDP broadcast operation. The necessary small machine-code addition is placed in a new PE section and connected to the existing game call path.

**Why:** This allows the game's existing LAN-discovery mechanism to function under the tested compatibility environment without replacing the protocol, adding an external service or opening unrestricted network connections from the patcher.

An early experimental implementation exposed a calling-convention error and could crash. The final implementation was corrected and retested before integration. Successful LAN tests included communication between differing game versions and operating-system environments.

This is **not** a direct-IP or VPN multiplayer feature. Broadcast-based discovery may still need additional work across routed VPNs such as Tailscale, and multiplayer can still behave differently on particular Wine/Proton configurations.

### 5. Font resource cleanup on Steam Deck and Linux

**Observed failure:** The game could run successfully but display an error such as **“Error remove font resource”** when closing on Steam Deck. Examination of the font-loading and removal path pointed to how older game-bundled TrueType font resources were handled by the compatibility layer.

**Investigation:** Several alternatives were examined, including file-path and case handling, font registration and exception suppression. Simply hiding the error did not address the underlying resource behavior.

**Final change:** On **Linux/Proton only**, the patcher changes the internal `head.fontRevision` metadata of five verified DejaVu TrueType files to a compatibility revision (`1000.0`). The actual letter shapes and font content are not replaced.

**Why:** This workaround addressed the observed cleanup failure in testing without disabling the game's fonts. The same change is **not performed on Windows**, where modified font resources could cause a different compatibility problem. Existing unexpected font changes are treated as a reason to stop rather than patch over them.

### 6. Language files and dynamic text placeholders

**Observed failure:** Some Steam 1.4 localization files contained translated or damaged *placeholders* used by the engine to insert dynamic values, such as a player's name or a payment amount. In-game messages could consequently display a literal `@`, an incorrect variable name, or malformed text.

**Investigation:** Known localization files were compared with older retail resources and with the variable names expected by the game. The work distinguished ordinary translated prose from special placeholders that must retain their exact spelling.

**Final changes:** Verified corrections are provided for the supported **Polish, Spanish, French and Hungarian** language files. The Polish corrections retain Steam-specific entries while repairing affected placeholders. Equivalent inconsistencies in other supported language files are corrected only where a known input and expected output are available.

**Why:** A user-facing sentence may be translated, but an engine token such as `@player` must remain recognizable to the game. The patcher does not attempt to translate every remaining English controller label or rewrite other localization files that have not been verified.

### 7. Controller support — separating boats from hydroplanes

**Observed problem:** The Steam edition's existing input arrangement did not provide the preferred analog-trigger behavior for modern controllers. Some mappings conflicted when a controller and keyboard were active together. The hydroplane used different internal actions from the boat, so a straightforward button remap was not sufficient.

The solution required both **controller configuration changes** and selected fixes to the game's runtime input logic. Editing `ControllerSchemes.xml` alone could not solve the interaction between keyboard, controller, camera and on-screen prompts.

#### Boat controls

The controller changes are **optional** for supported Steam installations:

| Mode | Boat throttle behavior |
| --- | --- |
| **No controller changes** | Keep the original controller behavior; apply the other compatibility fixes only |
| **Boat Analog** | R2/L2 provide proportional forward/reverse throttle input |
| **Boat Step** | R2 increases and L2 decreases a retained throttle setting, including reverse |

In both controller-enabled modes, the left stick steers the boat. The right stick controls the boat's side/rear camera views. Separating the throttle triggers permits more natural simultaneous steering and power control.

**Why two throttle modes:** Some players prefer continuously proportional triggers, while others prefer increments that remain set after releasing a trigger. Neither preference should be forced on users who want the original controls.

#### Hydroplane controls

The hydroplane has a separate flight-control path. For both Boat Analog and Boat Step modes, the final mapping uses:

- **Left stick:** pitch and roll in flight, with corrected steering behavior on water.
- **Right stick (horizontal):** directional/yaw control, **not** camera movement during flight.
- **R2/L2:** increase/decrease a **retained, step-based engine power** setting.
- **Keyboard:** preserve the game's separate flight actions, including arrow-key steering, Page Up/Down power controls, Z/X directional input and A/S/D camera controls.

**Why this differs from the boat:** The game already distinguishes its aircraft-power and rudder actions from ordinary boat throttle and steering. A direct copy of the boat mapping could reverse a control, select the wrong action or prevent the hydroplane from taking off.

#### Keyboard and controller together

The input logic was also adjusted so that a connected controller does not prevent normal keyboard steering and throttle actions. Earlier tests exposed cases in which analog input overrode keyboard input or the keyboard only responded during key release. The final design preserves physical keyboard controls alongside gamepad input while keeping flight and boat behavior separate.

#### Camera and on-screen prompts

The original boat camera mapping could interfere with the hydroplane's directional controls. The final solution keeps right-stick camera behavior for boats while preventing that same stick from simultaneously moving the camera during hydroplane flight. Keyboard camera controls remain available.

A separate display issue affected **Q/X quick-menu prompts**. Changing text in the HUD XML did not fully solve it because the game generated some captions at runtime. Two on-screen caption paths were corrected so the prompt follows keyboard versus controller input, without changing the underlying quick-menu action.

### 8. Why controller development required multiple test rounds

The controller work was not a single replacement of button numbers. Testing uncovered several distinct causes of regressions:

1. **Action numbering:** Internal game actions and controller XML indices did not use the same numbering convention. Initial mappings could assign a trigger to a steering action or leave a stick ineffective. The indices were corrected and tested in the game.
2. **Direction and state:** Flight controls and on-water hydroplane steering required different sign handling. Reversing the entire steering value caused unwanted movement; the successful correction affected only the relevant stick contribution.
3. **Keyboard priority:** An active gamepad could change which input path the engine used. Separate physical-keyboard handling was needed to preserve normal keyboard controls.
4. **Camera conflicts:** A first attempt to disable flight camera movement also blocked keyboard camera controls. Later revisions separated those paths.
5. **Windows code layout:** Controller fixes first examined under Proton could not simply reuse raw machine-code addresses on the Windows variant. Relocated references and jumps were corrected after testing exposed crashes in the initial Windows integration.
6. **Display versus behavior:** The Q/X prompt had to be fixed in the runtime caption logic, not just in a static XML label.

Each of these investigations produced temporary test variants. **Only the validated result of those iterations is included in v1.1.0.** The temporary experiments, crash-inducing revisions, debug message boxes and broad workarounds are not part of the released patcher.

---

## From separate experiments to one patcher

Earlier work used small command-line programs and isolated test patches. That made individual problems easier to diagnose, but it also made installation error-prone: users could accidentally combine incompatible builds, apply a Linux-specific adjustment on Windows or lose track of which files had been modified.

Version 1.1.0 integrated the accepted corrections into one common implementation with two native patcher executables. The game-file changes remain conditional on the detected edition, operating system and selected controller mode.

### What v1.1.0 changes in the game installation

| File or resource | Purpose | Where applied |
| --- | --- | --- |
| `Run.exe` (verified retail 1.2) | Preserve the original v1.0.0 compatibility corrections | Windows retail support |
| `Run.exe` (verified Steam 1.4) | LAN broadcast fix; Steam profile integrity adjustment where required; optional controller runtime changes | Windows / Linux, with platform-specific differences |
| `Launcher.exe` | OpenGL/Mesa compatibility; Linux-only integrity correction | Steam 1.4 |
| Four `PostProcessing.xml` files | Correct known Bloom blur vectors | Steam 1.4, both platforms |
| `Language.pl`, `Language.esp`, `Language.fr`, `Language.hu` | Repair verified localization placeholders and text inconsistencies | Steam 1.4, both platforms |
| `Configs/ControllerSchemes.xml` | Optional trigger/stick and hydroplane mappings | Steam 1.4 when controller mode is selected |
| Five known DejaVu `.ttf` files | Adjust font revision metadata to address cleanup errors | Steam 1.4 on Linux/Proton only |

The patcher does not modify saves as part of these fixes. Existing game assets and resources outside the verified patch plan are not replaced.

### Edition and platform rules

**Retail 1.2:** Uses the original Windows compatibility corrections. Steam-specific controllers, language files, Bloom changes and launcher/integrity fixes are not applied to this edition.

**Steam 1.4 on Windows:** Receives the verified Steam compatibility work, including launcher OpenGL handling, LAN broadcast, Bloom, language and optional controller changes. It deliberately **does not** receive the Linux-only profile MD5/integrity changes or the Linux font revision workaround.

**Steam 1.4 on Linux, Proton or Steam Deck:** Receives the compatible Steam patch set including the profile-integrity and font adjustments. The patcher is a native Linux ELF, but the patched game still runs through Proton/Wine.

**Unknown versions:** They are rejected. The patcher is not a general-purpose converter for arbitrary editions or previously modded executables.

### File recognition and integrity checks

The installer examines file contents and expected **SHA-256 hashes**, rather than trusting a filename alone. Known inputs are converted to known outputs. Small embedded binary-difference files (`.aqpt`) carry only the necessary patches, not complete game executables. The patch format verifies its source, structure and final output hash.

**Why:** A machine-code change intended for one precise executable layout could be unsafe on another build. Refusing unfamiliar inputs is a deliberate compatibility and data-protection decision, not an installation error to bypass.

### Installation, backups and Restore

The integrated engine checks the selected installation, determines a patch plan and saves copies of the original files in the game directory under `AquadelicGT-Patcher-Backups/v1.1.0/`. It records file names, original and patched hashes, the installed edition, platform and selected controller mode in an installation manifest.

When changing between **No controller changes**, **Boat Analog** and **Boat Step**, the patcher uses the backed-up originals to calculate the requested result. It updates only the controller-related files required for the mode change rather than repeatedly modifying graphics, fonts, localization and launcher resources.

**Restore original files** checks that the current files still correspond to recognized states before restoring the saved originals. It preserves backup history. The system also recognizes certain earlier backup manifests so that an old active patch is not silently overwritten.

**Why:** The original files should be recoverable, and repeated mode changes should not accumulate binary modifications. Installation and Restore should stop if the file on disk has changed unexpectedly.

**Important limitation:** The published v1.1.0 implementation saves the active installation manifest after applying file changes. A power loss, forced termination or crash at the wrong moment can therefore leave modified files without a completed manifest. The backup copies may remain available, but automatic Restore may no longer know how to locate them. Make a separate backup of the entire game directory before installing, and see [Known issues](KNOWN_ISSUES.md).

### Game-folder discovery and user interface

The patcher looks for familiar game installation locations and Steam libraries, including additional libraries listed by Steam. A user can manually select a folder if automatic discovery does not find the game.

The interface is served locally in the browser on `127.0.0.1`; it does not require an online account or remote patching service. Windows uses a native folder chooser rather than spawning a PowerShell window. Linux can use an installed desktop folder chooser or accept a manually entered path.

The integration work went through several rounds of practical installation and recovery testing. Each round exposed a different problem unrelated to the underlying gameplay patches:

| Problem found during integration | What was changed and why |
| --- | --- |
| An older backup manifest could block a fresh installation even after the original files had been restored | The patcher now considers the manifest **and** the real file hashes, rather than assuming that the existence of a manifest proves a patch is still active. |
| The Windows folder chooser opened an unwanted console/PowerShell window | A native Windows folder dialog replaced the helper process, and the Windows build uses the GUI subsystem. |
| Restore or the active game selection could be lost after closing and reopening the patcher | Recent game-folder selection and active-installation state are checked when the program starts again. |
| SteamOS exposed the same game directory through different symlinked paths | Directory identity is compared where appropriate, so two valid aliases do not automatically appear to be two unrelated installations. Unrelated game folders are still rejected. |
| A pre-install file check reset the user's requested controller mode to the already installed mode | The installed mode and selected mode are handled separately, allowing a deliberate switch between the three choices without first restoring everything. |
| Some status text and controller explanations repeated information or were unclear | The public interface was simplified and standardized in English, with a clearer exit/closed state. |

These adjustments improved installation and day-to-day usability **without changing the previously validated Steam game-patch bytes**. They also explain why the integrated patcher had to be tested as a program, not just as a collection of correct binary differences.

The release retains a browser-based interface for both operating systems. It is **not** a native Windows rendering interface or a separately network-hosted application.

---

## Verification and release process

### Gameplay and regression testing

Testing has included supported retail and Steam versions, Windows systems, Linux desktop environments and Steam Deck. The major areas examined were:

- Game startup, menu navigation, loading and saving profiles.
- Single-player operation and the corrected graphics/language resources.
- LAN lobby creation, discovery, joining and leaving, including cross-version/cross-system tests.
- Boat and hydroplane steering, throttle, camera and on-screen controller prompts.
- Keyboard operation with and without an active controller.
- Installing, switching controller modes, restarting the patcher and restoring files.

An external tester successfully connected a patched Steam 1.4 game running on a CachyOS laptop to an older, unpatched Windows XP game. A separate CachyOS desktop continued to freeze during multiplayer, including in tests without the patch. That issue remains under investigation and is not considered resolved by the successful laptop test.

### Reproducible Windows and Linux builds

The v1.1.0 Windows x86 and Linux x86-64 patcher executables were built from the **same published Go source tree** using **Go 1.23.2** and the respective target/build settings. Independent rebuilds produced identical binary hashes:

| Patcher executable | SHA-256 |
| --- | --- |
| Windows x86 | `30334e6f104c2a2a5272989268896890ee6abbffd024d50e8ef115190b97018a` |
| Linux x86-64 | `dc0690eaea682a53609097e7df6aa8050ede4d1b62650bf6b203cd1a78df2db9` |

See [Building from source](BUILDING.md) for the required Go version and commands. Byte-for-byte reproducibility establishes that the supplied source tree matches these released patcher binaries; it is **not** a guarantee of error-free behavior on every machine.

### Security review and Microsoft Defender

An independent source-code review identified **no confirmed Critical or High severity vulnerabilities under its stated threat model**, but reported medium- and lower-severity hardening issues. These include recovery after interrupted operations, incomplete validation of some local HTTP requests, and risks involving specially crafted manifests, symlinks or untrusted paths. Those limitations remain in the published v1.1.0 build and are documented in [KNOWN_ISSUES.md](KNOWN_ISSUES.md).

An original Windows build was flagged by Microsoft Defender. Following a submission for review, Microsoft stated that the submitted file did not meet its malware or potentially unwanted application criteria and **removed that detection**. This determination concerns the submitted file; it does not certify the entire application or guarantee identical results from every antivirus product.

---

## Investigations and workarounds not included in v1.1.0

The project tested a number of ideas that were not part of the final compatibility patch. Keeping them separate from shipped fixes helps explain the scope of the release:

- **Broad exception suppression and invalid-free workarounds:** rejected in favor of correcting the affected initialization and ownership paths.
- **Temporary controller mappings:** replaced after discovering numbering, steering-direction and camera conflicts.
- **Debug hooks, temporary logging and experimental GUI modifications:** used only during diagnosis.
- **Automatic cross-edition save conversion:** not part of the installer; save files are not rewritten.
- **Forcing Zink or a particular graphics driver:** not performed; renderer choice remains with the user.
- **OpenAL library replacement or a forced default audio device:** not included. The audio initialization issue remains a separate investigation.
- **Direct-IP/VPN matchmaking:** not implemented; the LAN fix preserves the game's existing broadcast discovery behavior.
- **Wide-ranging game HUD or widescreen redesigns:** outside the scope of the compatibility work.
- **Linux-only font and integrity changes on Windows:** deliberately excluded because the platform-specific behavior differs.

## Known limitations and ongoing reports

**Multiplayer freezes on some Linux configurations.** A reported CachyOS desktop freezes even when the game is tested without this patch, while other machines have successfully joined multiplayer sessions. Graphics drivers, Wine/Proton behavior, synchronization and the original game engine remain possible factors; none has been confirmed as the sole cause.

**Intermittent OpenAL device error.** The message `TSoundManager.Initialize: can't open preferred device` has been seen on Windows and Steam Deck, including unpatched game installations. It points to a failed OpenAL device-opening step. A reported association with 720p could not be reproduced consistently and is not an established cause.

**Configuration-specific input and graphics differences.** Some older game behaviors differ between Windows and Proton or between windowed and fullscreen operation. The release does not claim to solve every controller, compositor, display or graphics-driver issue.

**Interrupted installation and untrusted paths.** Automatic Restore is not guaranteed after abrupt interruption, and the original release has documented path-validation hardening gaps. Keep independent backups and do not patch game directories or backup manifests obtained from untrusted sources.

See [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for the concise user-facing list and reporting guidance.

## Development and maintenance principles

The public history is intentionally separate from the short [README](README.md) and Release notes. The README explains how to use the patcher; this document explains the technical decisions and investigations behind it.

Future changes should continue to use supported-file hashes, reversible operations, platform-specific behavior and real regression tests. A new build should not be considered equivalent to the released one merely because it has the same version number or a similar antivirus result.

If an issue can be reproduced, useful reports include the game edition, operating system, graphics hardware, Wine/Proton version, controller configuration, exact error and whether the behavior occurs without the patch. Reports of successful configurations are valuable as well.

The project remains unofficial and independent of the original game's developers or publisher. Game executables, assets, trademarks and other third-party content remain the property of their respective rights holders.
