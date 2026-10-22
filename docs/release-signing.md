# Signed Bridge releases

Decided 2026-10-22. Status: code ready, **turned on when the first public key is in the build**.

## Why

Bridge runs as root (Administrator on Windows) on every garage, including garages of people we
do not know, and it updates itself. Until now an update was trusted because it came from this
repository's GitHub releases and matched the release's own `SHA256SUMS`. Whoever controls the
GitHub account (a stolen session, a leaked token, a malicious workflow) could therefore publish a
release that every garage installs as root within a day. Signing moves the trust to a key that is
**not on GitHub and not on the gateway**: without it, a release does not install.

## How it works

- `cli/internal/release` holds the trusted **public keys** (`trustedKeys` in `release.go`). With
  at least one key in the build, self-update (`selfUpdate` and `garageai update`) fetches the
  release's `SHA256SUMS.sig` and refuses the release unless it verifies.
- The signature covers `garageai-release-signature 1`, the **version**, and the release's
  `SHA256SUMS`. An old signed release cannot be passed off as a newer one, and a new release
  cannot be passed off as an old one.
- Ed25519, Go standard library only. Key id = the first 8 bytes of the public key's SHA-256.
- The Xyntetik Runner that `garageai runtime install` downloads is Joakim's release, checked
  against its own `SHA256SUMS`. It is not part of this.
- **Not covered yet:** the first install. `install.sh` and `install.ps1` still trust the
  GitHub release and its checksum, so a new garage trusts GitHub once. Everything after that is
  signed. Verifying in the installers is a later step (OpenSSL 3 can verify Ed25519, older
  macOS LibreSSL cannot).

## Keys

Two keys, both in `trustedKeys`: a **primary** for every release and a **backup** kept offline.
If the primary is lost or leaks, a release signed with the backup removes it and adds a new one.
If both were lost, garages could no longer update themselves, and every owner would have to run
the install line again. Keep both safe.

Create them **on your own computer**, not on the gateway and not in CI:

```sh
cd garageai/cli
go run ./tools/releasesign keygen -out ~/garageai-release-primary.key
go run ./tools/releasesign keygen -out ~/garageai-release-backup.key
```

- Store each private key file in your password manager (as a secure note or an attachment). Then
  remove the backup key from the disk.
- Paste the two **public** keys printed by `keygen` into `trustedKeys` in
  `cli/internal/release/release.go`. Public keys are safe to share, so they can go in a PR. From
  the release that contains them on, signatures are required.

## Releasing (each time)

1. Tag and push as before: `git tag -a bridge-vX.Y.Z -m … && git push origin bridge-vX.Y.Z`. CI
   builds the binaries and publishes the release with `SHA256SUMS`.
2. Sign on your own computer:

   ```sh
   gh release download bridge-vX.Y.Z -p SHA256SUMS -R magnusfroste/garageai -D /tmp/rel
   go run ./tools/releasesign sign -key ~/garageai-release-primary.key -version vX.Y.Z /tmp/rel/SHA256SUMS > /tmp/rel/SHA256SUMS.sig
   go run ./tools/releasesign verify -pub <primary public key> -version vX.Y.Z /tmp/rel/SHA256SUMS /tmp/rel/SHA256SUMS.sig
   gh release upload bridge-vX.Y.Z /tmp/rel/SHA256SUMS.sig -R magnusfroste/garageai
   ```

3. Canary garages then update at their next heartbeat. An unsigned release is refused: the
   heartbeat reports `update.last_error` ("release vX.Y.Z has no signature …"), and the Operations
   Center shows it. Nothing is installed until the signature is there.

## Turning it on

1. Generate the two keys and send the two public keys. A PR puts them in `trustedKeys`.
2. Release that version as usual. That release itself is installed by garages running older
   Bridge versions, which do not check signatures, so sign it anyway: from then on every release
   must be signed.
3. Update AUDIT.md claim 6 and the security page ("self-update installs only releases signed with
   a key kept off GitHub").
