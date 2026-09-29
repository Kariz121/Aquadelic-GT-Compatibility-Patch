# Aquadelic GT Compatibility Patch — Development History

This document is a condensed reconstruction of the patch development history from preserved project notes, binary comparisons and diagnostic records.

It is not intended to list every temporary executable. Many intermediate builds were deliberately diagnostic and were later removed.

## Phase 1 — original XMLFix branch

The earliest branch grew from attempts to make the original 2008 executable behave reliably on modern Windows.

The work identified several independent compatibility problems:

- legacy XML/MSXML initialization ordering
- missing or incomplete Delphi unit initialization
- GUI/audio/CameraLifter startup initialization
- Ogg/Vorbis callback-table initialization
- profile Boolean serialization
- play-time display calculation
- several cleanup and invalid-free symptoms

This branch eventually reached an old `V35`, but it had accumulated many defensive guards and experimental workarounds.

A large part of the branch was intentionally discarded later.

Experiments that were useful diagnostically but were not intended to survive included:

- menu/audio disabling
- no-music variants
- OpenAL default-device bypasses
- forced OpenAL direct exports
- diagnostic MessageBoxes
- generic invalid-free suppression
- broad nil-list guards

The important result of this phase was knowledge: it identified which compatibility changes were genuinely useful and which changes merely hid deeper bugs.

## Phase 2 — Clean RootFix rebuild

The Clean RootFix branch was deliberately rebuilt close to the original executable instead of continuing to stack patches on the old V35.

`Clean RootFix V4` became the clean reference base.

It retained only selected, verified compatibility work from the earlier branch:

- XML/Delphi compatibility initialization
- the required `.itext` PE mapping adjustment
- Ogg/audio initialization
- profile Boolean save compatibility
- corrected play-time display behavior

The goal was to diagnose the remaining crash without old guards changing program behavior.

## Phase 3 — Runtime Error 2 investigation

The major remaining problem happened while stopping/cleaning up a race session.

The investigation progressed from a generic `Runtime Error 2` to a specific object-lifetime bug.

Major findings included:

- the failure occurred during cleanup, not map loading or normal gameplay
- the failing allocation was associated with Delphi list/object structures
- the invalid object was identified as a `TObjectList`
- repeated `TObject.Free` calls were observed for the same scene-node path
- the destruction chain was narrowed to `TCameraLifter.Destroy -> TSceneNode.Destroy`
- the same object was being destroyed twice

By V33, the investigation had already narrowed the problem to ownership/removal behavior in the scene-node child list.

Some intermediate V17–V27 diagnostic executables also contained instrumentation bugs of their own. These were corrected during the investigation and were never intended for release.

## Missing V34 record

The preserved chat history does not contain a complete standalone record for the transition from V33 to V34.

This is not treated as a separate public feature gap because the surviving later documentation records the final root cause and the root fix implemented immediately afterwards.

## Clean RootFix V35 — double-free root fix

The final cause was identified in the `TCameraLifter` destruction path.

The problematic sequence involved an owning `TObjectList` (`OwnsObjects=True`):

1. a child object was removed through a path that allowed the owning list to free it
2. destruction then continued on the already-freed object
3. the same object/list state was reached again during cleanup
4. the Delphi memory manager eventually reported `Runtime Error 2`

The root fix changed the removal semantics so the object is detached from the owning list without being automatically freed at that point.

This replaced the earlier strategy of suppressing invalid frees or broadly guarding cleanup calls.

## Clean RootFix V36–V38 — multiplayer initialization

After the memory crash was fixed, previously hidden multiplayer problems became visible.

Diagnostic logging found an uninitialized function pointer in the multiplayer path.

The original game contained the necessary initialization routines, but they were not being run at the required time.

V38 restored the original multiplayer/network unit initializers in their intended order.

After this change, multiplayer progressed normally instead of failing when entering the multiplayer menu/path.

## Clean RootFix V39–V40 — lobby/session timer

A separate multiplayer cleanup bug remained.

One reproducible case involved creating a lobby, waiting for more players, and then leaving it.

The session pointer could be cleared while a periodic timer/callback remained active. The callback then attempted to use the cleared session state.

The stable V40 base fixed this session/timer lifetime problem.

## V41–V44 — COM/XML Runtime Error 24

A clean Windows test system exposed another independent problem:

```text
Runtime Error 24 at address 004AB448
```

Installing MSXML did not solve it.

A diagnostic build captured the underlying HRESULT:

```text
0x800401F0
CO_E_NOTINITIALIZED
```

This proved that the XML failure was caused by COM not being initialized on the required thread/path.

V42 demonstrated the diagnosis by calling `CoInitialize`.

V43 attempted to move initialization earlier, but this was a regression and was discarded.

V44 became the stable implementation:

- checks the game's existing COM initialization state before the first affected XML operation
- invokes the game's original COM initialization routine when needed
- replays the original instructions
- does not repeatedly initialize COM for each XML operation
- leaves cleanup to the game's existing finalization path

## Stable public executable

The stable compatibility base used by the public patcher is V44.

Patched executable SHA-256:

```text
949768e738e43aa2240176b752ac2aad9d0b66dc442aa87fd21634b58d738aa5
```

The final build was regression-tested for startup, singleplayer, multiplayer lobby creation/joining, boat changes, ready state, race start/play, returning to the lobby, leaving/re-entering multiplayer and normal shutdown.

## Separate widescreen experiment

After V44 was stable, a separate experimental 16:9/FHD GUI project was developed.

That work changed XML GUI layouts and, in one test branch, the loading-progress calculation.

Those widescreen changes are intentionally **not part of the public compatibility patch**, because the original GUI behavior is safer across the game's supported resolution options.
