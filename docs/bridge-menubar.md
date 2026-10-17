# Bridge in the menu bar

Sketch, 2026-10-17. Decided: build it after v0.7.0. Inspired by Xyntetik Runner's tray
(`src/tray_macos.m`, `src/tray_win.c`: an `NSStatusItem` on macOS and a notification-area icon on
Windows, rendering a menu that portable C code describes).

## Name: GarageAI, not Bridge (decided 2026-10-17)

What the operator sees on their own machine is called **GarageAI**: the menu-bar item and its
tooltip ("GarageAI · garage-m2 · Live"), notifications, the login item ("Open GarageAI at login"),
the control panel. "Bridge" stays the name of the program where how it works is the point
(GitHub, the docs, the landing page's "GarageAI Bridge connects your GPU"), the way Tailscale's menu
says Tailscale while tailscaled does the work. The command is `garageai` already. Service
identifiers stay as they are (renaming them means migrating every garage); the Windows services'
display names can become "GarageAI tunnel" and "GarageAI heartbeat" without risk, and the
terminal headers "GarageAI v1.x — connect" when those texts are touched anyway.

## Why

Today a garage is invisible on the machine it runs on. The operator learns that it works from the
portal, and that it stopped from an e-mail or from us. A menu-bar icon is the smallest form of the
control panel in the vision ("see what your GPU can earn, go live with one click"):

- **Reassurance.** Live, serving, paused: one glance, no terminal, no portal.
- **Earnings where the hardware is.** "Today: 142 requests, €0.41" is the reason to keep the
  machine on.
- **The common actions without a terminal.** Pause for an hour, pause while gaming, open My
  garage, check this machine.
- **A place for the few notifications that matter.** First request served, garage offline,
  Smart fetched a model, the weekly figure.

## What it shows

The menu, top to bottom:

1. **Status.** "GarageAI", the garage name and one word: Live, Serving, Paused, Away (on battery, gaming) or
   Offline, with the reason when it is not live ("Ollama stopped", "tunnel down"). The runtime
   and the model being served, with tokens per second from the last acceptance test.
2. **Today.** Requests, tokens, earned (in the portal's currency), and when the last request came
   and through which route (pool or specific garage).
3. **Models.** Offered models; models Smart or the portal is fetching, with progress ("2.1 of
   13.8 GB"); a model that failed to download, with its reason.
4. **Pause.** For 1 hour, until I resume, and "Pause while I game" (the automatic pause from the
   vision, when it exists). Resume when paused.
5. **Open My garage…** (the portal) and **Check this machine…** (doctor's summary in a small
   window, with "Copy for support").
6. **Bridge version** ("up to date", or "updating to 1.2.0"), **Show at login**, and **Hide icon**:
   hiding the icon never stops the garage.

Icon states: live (outline), serving (accent, with the number of requests in flight), paused or
away (pause glyph), needs you (warning: runtime stopped, tunnel down, heartbeat rejected). A
template image on macOS, so it follows light and dark menu bars.

## How it works

**A third process, in the user's session, read-only by default.** The heartbeat (root/SYSTEM,
every 5 minutes) and the tunnel (root/SYSTEM, always on) stay as they are. `garageai tray` runs as
the logged-in user, started at login (a LaunchAgent on macOS; the Run key under HKCU on Windows).
It never holds the garage's secrets and never talks to the gateway.

**Its data are files the services already write or will write, secret-free and readable by
users:**

| Data | Source | Exists |
|---|---|---|
| Garage name, runtime, port, transport | `bridge-public.json` | yes |
| Requests in flight and in the last 5 minutes, TTFB | tunnel stats summary (`tunnel/stats.go`) | yes, root-only today: make a public copy |
| Lifecycle state (paused, away, why) | `agent.state` in the heartbeat; pause state file | yes |
| Models, downloads and failures | models state (`service/models.go`) | yes |
| Problems (doctor codes) | the last heartbeat's profile | in memory: write `status.json` |
| Live in the portal, offered models, today's requests and earnings | the portal's heartbeat answer | **new**: the portal adds `garage: {live, offered, requests_today, tokens_today, earned_today}` to the node-heartbeat response, the heartbeat writes it to `status.json` |

So the one new thing on the portal side is a few fields in the heartbeat answer (Lovable), and
on Bridge's side a `status.json` next to `bridge-public.json`, rewritten by every heartbeat. The
tray polls the files every 5 seconds (like Runner's tray polls `/health`); the figures are at
most 5 minutes old, which is fine for "today".

**Actions.** Opening the portal and doctor need nothing. Pause and resume change a root-owned
state file: the tray runs `garageai pause` through the platform's own elevation prompt
(`osascript … with administrator privileges`, which offers Touch ID; UAC on Windows). Later, if
that is too much friction for a one-hour pause, a per-garage "pause requests" file the user may
write, which the tunnel service (it polls state every 10 s) honours: pausing is harmless, so a
local user being able to pause is acceptable.

**Notifications** come from the same files: the tray compares two polls (first request ever,
live → offline, a download finished) and posts a native notification. At most a few a day; the
weekly figure on Monday morning.

## Building it

Bridge is pure Go and cross-compiled without cgo today. The tray changes that only where needed:

- **Windows:** no cgo. `Shell_NotifyIconW`, a hidden message window and `TrackPopupMenu` through
  `golang.org/x/sys/windows` (already a dependency). Notifications through the same icon.
- **macOS:** a small Objective-C file (as Runner's `tray_macos.m`): `NSStatusItem`, `NSMenu`,
  `NSApplicationActivationPolicyAccessory` (no Dock icon), `UNUserNotificationCenter`. That needs
  cgo, so the darwin binaries are built on `macos-latest` in the release workflow (arm64 and
  amd64 with `-arch`), not cross-compiled from Linux. One binary still: `garageai tray` is a
  subcommand, and every other command keeps working without a window server.
- **Linux:** no tray (servers). Maybe AppIndicator later for Linux desktops.

The menu itself is described by portable Go (a list of rows: label, action, check, separator,
submenu), like Runner's `tray.c`; the two platform files only render it and pass clicks back.
That keeps the logic testable in CI on every OS without a screen.

**Getting it onto the machine:** `garageai connect` ends with "Show GarageAI in the menu bar?
[Y/n]" on macOS and Windows (installs the login item and starts it); `garageai tray --install` /
`--uninstall` for later; `garageai uninstall` removes it. The portal's Connect page mentions it.

## Phases

1. **Read-only icon:** `status.json` from the heartbeat (with today's figures from the portal),
   the tunnel stats public copy, `garageai tray` with status, today, models, Open My garage,
   Check this machine, Show at login, Hide icon. macOS and Windows.
2. **Pause and resume** from the menu, with the elevation prompt; then the user-writable pause
   request if the prompt is too much.
3. **Notifications:** first request, offline, Smart download done, weekly figure.
4. **The control panel:** a small window behind "Open GarageAI…" with the week's earnings curve,
   price per model, what Smart would fetch, and "go live with one click" for a machine that has
   Bridge but no garage yet (`garageai runtime` + connect behind one button).

Gaming auto-pause (from the vision) slots into phase 2: the tray, in the user's session, is the
one process that can see that a game is in the foreground.
