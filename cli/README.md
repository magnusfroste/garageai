# garageai (Go) — experimental

A single binary for garage operators, built next to `scripts/garageai-connect.sh` to see whether it
works better before anything is decided. Today it does one thing: the garage profile.

```sh
garageai doctor           # what runs on this machine, what is wrong, and how to fix it
garageai doctor --json    # the same as the garage profile (schema 1), identical to
                          # garageai-connect.sh --doctor --json
garageai version
```

- Standard library only. One codebase for Linux, macOS and Windows (amd64 and arm64).
- **Linux:** reads `/proc` directly, no `ss` or `lsof` needed.
- **macOS:** uses `lsof`.
- **Windows:** uses `netstat` and `tasklist`.
- Probes candidate ports in parallel, so it is faster than the script.
- Never reports keys or whole command lines. Only selected vLLM/SGLang flags, and whether an API key is set.

## Build and test

```sh
cd cli
go test ./...
go build -o garageai .
../scripts/tests/parity-doctor.sh "$PWD/garageai"   # same profile as the script?
```

CI builds six binaries (`garageai-{linux,darwin,windows}-{amd64,arm64}`) with SHA256SUMS as the
artifact `garageai-binaries`. Installing via `curl` like norrivaagent needs no signing: curl sets no
quarantine flag on macOS.

Next, if the comparison holds:
- `connect`, with the same onboarding reports and profile,
- `run` as the heartbeat service,
- a release flow with an install script.
