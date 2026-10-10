# GarageAI Bridge (`garageai`) — experimental

**Bridge** is the bridge between an operator's GPU and the GarageAI network, in both directions:
the garage sells to the network, and later the operator uses the network from the garage. It is one
binary, `garageai`, built next to `scripts/garageai-connect.sh`. The two are compared before
anything is decided. The vision and the phases are in [docs/bridge.md](../docs/bridge.md).

Today it does one thing: the garage profile.

```sh
garageai doctor           # what runs on this machine, what is wrong, and how to fix it
garageai doctor --json    # the garage profile (schema 1), identical to garageai-connect.sh --doctor --json
garageai version
```

- Standard library only. One codebase for Linux, macOS and Windows (amd64 and arm64).
- **Linux:** reads `/proc` directly, no `ss` or `lsof` needed.
- **macOS:** uses `lsof`.
- **Windows:** uses `netstat` and `tasklist`.
- Probes candidate ports in parallel: 0.24 s against 0.7 s for the script.
- Never reports keys or whole command lines. Only selected vLLM/SGLang flags, and whether an API key is set.

## Build and test

```sh
cd cli
go test ./...
go build -o garageai .
../scripts/tests/parity-doctor.sh "$PWD/garageai"   # same profile as the script?
```

CI ("Bridge (Go)") tests on Linux, macOS and Windows, and runs the comparison with the script on
Linux and macOS. It builds six binaries (`garageai-{linux,darwin,windows}-{amd64,arm64}`) with
SHA256SUMS as the artifact `garageai-bridge-binaries`.

Installing via `curl` like norrivaagent needs no signing: curl sets no quarantine flag on macOS.
