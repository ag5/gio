# AG5 additions to Gio

This document describes the five commits on branch `ag5/v0.10.3` on top of
upstream Gio **v0.10.3** (`6cc1cf3e`).
The module path remains `gioui.org`, including imports such as
`gioui.org/app` and `gioui.org/widget`.

The patches were rebased unchanged from `ag5/v0.10.2`. The tag
`v0.10.2-patched.1` still identifies the previous version based on Gio v0.10.2.

## Patch inventory

| Commit | Addition | Platform |
| --- | --- | --- |
| `0e1e7661` | Expose editor dimensions and scrolling helpers | All |
| `f47bc84a` | Add `app.Hidden` to hide the application | macOS |
| `8f8d577e` | Add `app.WindowID` and `app.Config.ID` | API available everywhere; used by macOS |
| `010f1250` | Remember window position and size across restarts | macOS |
| `215ccce1` | Add `app.PresentationTransaction` | macOS, Metal renderer |

The changes are in [widget/editor.go](widget/editor.go),
[app/os.go](app/os.go), [app/window.go](app/window.go), and
[app/os_macos.go](app/os_macos.go), with an option test in
[app/window_test.go](app/window_test.go).

## Editor scroll helpers

Four methods expose information from the editor's internal text view so callers
can display a scrollbar or scroll to the bottom of multiline content:

| Method on `*widget.Editor` | Result or effect |
| --- | --- |
| `Dimensions() layout.Dimensions` | Visible text viewport dimensions. |
| `FullDimensions() layout.Dimensions` | Dimensions of all shaped text, including content outside the viewport. |
| `ScrollTop() int` | Current vertical scroll offset in pixels; this is a getter. |
| `ScrollToEnd()` | Scrolls vertically to the bottom using the current content and viewport heights. |

Read the dimensions after laying out the editor with the current content, font,
and constraints. These getters return cached layout information; they do not
perform a new layout. For example, with `editor` referring to a laid-out editor:

```go
viewportHeight := editor.Dimensions().Size.Y
contentHeight := editor.FullDimensions().Size.Y
top := editor.ScrollTop()

// Fractions of the full content height for a vertical scrollbar thumb.
start, end := float32(0), float32(1)
if contentHeight > viewportHeight {
    start = float32(top) / float32(contentHeight)
    end = float32(top+viewportHeight) / float32(contentHeight)
}
// Use start and end when laying out the scrollbar.
```

`ScrollToEnd` adds `contentHeight - viewportHeight` to the current vertical offset
and lets the text view clamp it to its scroll bounds. It does not move the caret
or selection, scroll a single-line editor horizontally, or enable automatic
following of newly appended text. Apply it after layout has established the
relevant dimensions, then render another frame to display the updated offset.
The editor's normal caret-following behavior can subsequently change the offset.

This patch supplies metrics and a scroll-to-bottom operation. It does not add a
scrollbar widget or an API for setting an arbitrary editor scroll offset.

## Hide and show the application on macOS

`app.Hidden` is an additional `WindowMode`; its string representation is
`"hidden"`. On an existing window:

```go
w.Option(app.Hidden.Option()) // Hide the entire application.
// Later, show and activate it again:
w.Option(app.Windowed.Option())
```

The backend calls AppKit's application hide operation, so this affects all of the
application's windows. Configuring a window as `Windowed`, `Maximized`, or
`Fullscreen` activates the application again if it is hidden. `Minimized` retains
its window-minimization behavior and does not perform this application activation.
Only the macOS backend implements `Hidden`.

## Window identity and saved geometry on macOS

`app.WindowID(id string)` sets `app.Config.ID`, an identity for state belonging to
one window. It is separate from `app.ID`, which identifies the application.
The macOS backend uses this identity with AppKit frame autosaving to persist
window position and size in user defaults as the user moves or resizes a window.

Configure the identity before the first call to `w.Event()`, which creates the
native window:

```go
w := new(app.Window)
w.Option(
    app.Title("Editor"),
    app.WindowID("main-editor"),
    app.Size(1000, 700),
)
// Continue with the application's normal window event loop.
```

On subsequent runs, a saved frame is restored for a window starting in
`Windowed` mode, overriding its initial size and cascading placement. AppKit
refits the restored frame to the current screen configuration.

The persistence rules are:

- The key is the initial `WindowID`, falling back to the initial title when the
  ID is empty. Gio's default title is `"Gio"`, so even windows without explicit
  options participate in autosaving.
- Use a stable, distinct ID for each window, especially when titles change or
  multiple windows share a title. AppKit rejects a second window registering an
  autosave name already in use in the same process; that window falls back to
  cascading placement. Reusing a key across runs reuses its saved geometry.
- Changing `WindowID` after creation updates the reported configuration on macOS
  but does not change the autosave key. Changing the title also leaves that key
  unchanged.
- With both the ID and title empty, no autosaving is registered.
- If no saved frame exists, registration fails, or the initial mode is anything
  other than `Windowed`, creation falls back to cascading placement. For a
  non-windowed initial mode, autosaving is still registered when a key is available,
  but this restoration path does not apply the saved frame.

The other platform backends do not implement persistence through `WindowID`.

## Metal presentation synchronization on macOS

`app.PresentationTransaction(enabled bool)` controls whether Metal drawable
presentation is synchronized with Core Animation transactions. Synchronization
remains **enabled by default**, preserving upstream behavior that avoids visual
glitches during interactive resizing.

Applications that resize windows programmatically can opt out so completed
frames can be presented without waiting for the current Core Animation
transaction:

```go
w := new(app.Window)
w.Option(app.PresentationTransaction(false))
// Set this before the first w.Event() call.
```

The option is consumed when the native view is created; changing it later has no
effect. `app.CustomRenderer(true)` keeps presentation synchronization disabled
regardless of this option. Other platforms ignore it, and the macOS OpenGL layer
does not use it. Disabling it gives up the default synchronization during
interactive resizing.

## Maintaining the patch stack

To inspect the exact implementation covered here:

```sh
git log --reverse --oneline v0.10.3..215ccce1
git diff v0.10.3..215ccce1 -- app widget
```

When rebasing onto a new Gio version, update the baseline and inventory above,
check whether upstream now provides equivalent APIs, and recheck the platform
behavior described here. The existing `TestPresentationTransaction` checks the
option's default and boolean toggling; it does not exercise native presentation,
hiding, or window restoration.
