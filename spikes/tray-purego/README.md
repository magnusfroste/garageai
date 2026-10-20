# Spike: cgo-free macOS menu bar (2026-10-20)

`main.go` puts "GarageAI" with a Quit menu in the macOS menu bar using
[purego](https://github.com/ebitengine/purego) (`purego/objc`), with **no C and no cgo**.

Result: cross-compiled from Linux with `CGO_ENABLED=0` for darwin/arm64 and darwin/amd64 (3.1 MB);
on GitHub's macos-latest runner it started, stayed running, and the item showed in the menu bar
(screenshot from the run: "GarageAI" next to Spotlight). Go's linker ad-hoc signs it.

    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o tray .

See docs/bridge-menubar.md, "Building it".
