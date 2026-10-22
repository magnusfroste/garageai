// Package release signs and verifies Bridge releases (docs/release-signing.md).
//
// A release's SHA256SUMS is signed with an ed25519 key that is kept off GitHub and off the
// gateway. Bridge verifies the signature before self-update installs anything, so whoever controls
// the GitHub repository (or the portal) still cannot make every garage run new code as root
// without the key. The signed message names the version, so an old signed release cannot be
// replayed to push garages back to it.
package release

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const magic = "garageai-release-signature 1"

// SigAsset is the signature file published next to SHA256SUMS.
const SigAsset = "SHA256SUMS.sig"

// trustedKeys are the release public keys (base64, 32 bytes). More than one, so that a lost or
// retired key can be replaced by a release signed with another. Empty: verification is off (a dev
// build, or before the first key existed).
var trustedKeys = []string{}

// keys is trustedKeys decoded, by key id; a variable so tests can use their own.
var keys = decode(trustedKeys)

func decode(list []string) map[string]ed25519.PublicKey {
	out := map[string]ed25519.PublicKey{}
	for _, k := range list {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(k))
		if err != nil || len(b) != ed25519.PublicKeySize {
			panic("release: bad trusted key " + k)
		}
		out[KeyID(b)] = ed25519.PublicKey(b)
	}
	return out
}

// Required reports whether this build verifies release signatures.
func Required() bool { return len(keys) > 0 }

// KeyID names a public key: the first 8 bytes of its SHA-256, in hex.
func KeyID(pub []byte) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:8])
}

// message is what is signed: the format, the version, and the release's SHA256SUMS.
func message(version string, sums []byte) []byte {
	return append([]byte(magic+"\nversion "+normal(version)+"\n"), sums...)
}

func normal(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "bridge-")
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}

// Sign returns the content of SHA256SUMS.sig for a release.
func Sign(priv ed25519.PrivateKey, version string, sums []byte) []byte {
	sig := ed25519.Sign(priv, message(version, sums))
	pub := priv.Public().(ed25519.PublicKey)
	return []byte(fmt.Sprintf("%s\nversion %s\nkey %s\nsig %s\n", magic, normal(version), KeyID(pub), base64.StdEncoding.EncodeToString(sig)))
}

// Verify checks a SHA256SUMS.sig against the release's SHA256SUMS and the version Bridge is
// about to install, with the trusted keys.
func Verify(version string, sums, sigFile []byte) error {
	return verifyWith(keys, version, sums, sigFile)
}

func verifyWith(keys map[string]ed25519.PublicKey, version string, sums, sigFile []byte) error {
	fields := map[string]string{}
	lines := strings.Split(strings.TrimSpace(string(sigFile)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != magic {
		return errors.New("not a GarageAI release signature")
	}
	for _, l := range lines[1:] {
		k, v, ok := strings.Cut(strings.TrimSpace(l), " ")
		if ok {
			fields[k] = v
		}
	}
	if fields["version"] != normal(version) {
		return fmt.Errorf("the signature is for %s, not %s", fields["version"], normal(version))
	}
	pub, ok := keys[fields["key"]]
	if !ok {
		return fmt.Errorf("signed with an unknown key (%s)", fields["key"])
	}
	sig, err := base64.StdEncoding.DecodeString(fields["sig"])
	if err != nil || len(sig) != ed25519.SignatureSize {
		return errors.New("malformed signature")
	}
	if !ed25519.Verify(pub, message(version, sums), sig) {
		return errors.New("the signature does not match this release's SHA256SUMS")
	}
	return nil
}

// SetKeysForTest replaces the trusted keys; it returns a function that restores them.
func SetKeysForTest(pubs ...ed25519.PublicKey) func() {
	old := keys
	keys = map[string]ed25519.PublicKey{}
	for _, p := range pubs {
		keys[KeyID(p)] = p
	}
	return func() { keys = old }
}
