# GarageAI Bridge (`garageai`)

**Bridge** is the bridge between an operator's GPU and the GarageAI network, in both directions:
the garage sells to the network, and later the operator uses the network from the garage. It is one
binary, `garageai`, and since 2026-10-10 the default way to connect a garage: the portal's
command uses it. `scripts/garageai-connect.sh` stays as an internal backup and reference, kept in
step with Bridge by `scripts/tests/parity-doctor.sh`. The vision and the phases are in
[docs/bridge.md](../docs/bridge.md).

```sh
curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh
garageai doctor
```

## Commands

```sh
garageai connect [options]   # join the mesh, check the runtime, register, install the heartbeat
garageai doctor [--json]     # what runs on this machine, what is wrong, and how to fix it
garageai run [--once]        # the heartbeat (the installed service runs this every 5 minutes)
garageai uninstall           # remove the heartbeat and Bridge's configuration
garageai version
```

`connect` takes the same options and `GARAGEAI_*` environment variables as
`scripts/garageai-connect.sh`, and sends the same payloads to the portal (`register-node`,
`node-heartbeat`, `onboarding-report` with the garage profile). So the portal's command only swaps
`bash garageai-connect.sh` for `garageai connect`. It needs root for joining the mesh and installing
the heartbeat: run as a normal user it re-runs itself through `sudo` and passes the secrets on
stdin, never on the command line.

What `connect` does, like the script:
1. NetBird installed (installs the CLI client if not, after asking).
2. Join the mesh with the setup key, or keep the current connection. Without a setup key and not
   on the mesh, it says to get a new command from the portal.
3. The runtime answers an OpenAI-compatible API (with the API key when it needs one).
4. Reachable over the mesh: the runtime listens on the network, streamed replies carry usage, and
   for Ollama the loaded context window (reported as it is; the gateway enforces it).
5. Register; the portal runs the acceptance test.
6. Install the heartbeat as a service (systemd timer or launchd daemon) running `garageai run --once`
   from its own root-only copy of Bridge (`/usr/local/lib/garageai/garageai` on Linux,
   `/Library/Application Support/GarageAI/garageai` on macOS): the service runs as root, so it never
   runs the copy in the operator's home, which their account could change.
   The connect script's old heartbeat is removed, so there is one reporter with one token.

The heartbeat also carries Bridge's own report (`agent`): its version, the OS and what `doctor` finds
right now (severity, code and message; never keys). The portal keeps the latest one, and the
Operations Center shows it per garage, so a garage is seen from the inside, not only from the mesh.

Every step is reported to the portal (`onboarding-report`), with the garage profile at the start,
on a failure and when done, so a stuck connect is visible in the wizard, the admin and the
Operations Center.

**Updating:** run the install command again. When the machine has the heartbeat service, the
installer also runs `sudo garageai install-service`, which puts the new version in the service's
root-only place, so the heartbeat runs it from the next report on. `doctor` reports
`service_binary_unsafe` (and the Operations Center shows it) when a root service runs a binary a
normal user can change, as Bridge before v0.2.4 could set up.

## Implementation notes

- Standard library only. One codebase for Linux, macOS and Windows (amd64 and arm64). The
  heartbeat service is not implemented on Windows yet.
- **Linux:** reads `/proc` directly, no `ss` or `lsof` needed. **macOS:** `lsof`. **Windows:**
  `netstat` and `tasklist`.
- Mesh state comes from the NetBird service when it answers, else from the tunnel interface
  (a normal user may not be allowed to ask the service on macOS).
- Never reports keys or whole command lines.

## Build and test

```sh
cd cli
go test ./...
go build -o garageai .
../scripts/tests/parity-doctor.sh "$PWD/garageai"   # same profile as the script?
```

CI ("Bridge (Go)") tests on Linux, macOS and Windows and runs the parity test on Linux and macOS.
A tag `bridge-vX.Y.Z` publishes the six binaries plus SHA256SUMS as a pre-release; `install.sh`
picks the newest `bridge-v*` release and verifies the checksum.
