package release

import (
	"crypto/ed25519"
	"strings"
	"testing"
)

func TestSignVerify(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	other, otherPriv, _ := ed25519.GenerateKey(nil)
	defer SetKeysForTest(pub)()
	sums := []byte("abc  garageai-linux-amd64\n")
	sig := Sign(priv, "v1.3.0", sums)

	if err := Verify("v1.3.0", sums, sig); err != nil {
		t.Fatal("a good signature failed:", err)
	}
	if err := Verify("bridge-v1.3.0", sums, sig); err != nil {
		t.Fatal("the tag form of the version should be the same version:", err)
	}
	if err := Verify("v1.2.0", sums, sig); err == nil || !strings.Contains(err.Error(), "not v1.2.0") {
		t.Fatal("a signature for another version was accepted (replay):", err)
	}
	if err := Verify("v1.3.0", []byte("abd  garageai-linux-amd64\n"), sig); err == nil {
		t.Fatal("changed SHA256SUMS were accepted")
	}
	if err := Verify("v1.3.0", sums, Sign(otherPriv, "v1.3.0", sums)); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatal("a signature by an untrusted key was accepted:", err)
	}
	if err := Verify("v1.3.0", sums, []byte("hello")); err == nil {
		t.Fatal("garbage was accepted")
	}
	// A second trusted key (the backup) works too.
	restore := SetKeysForTest(pub, other)
	if err := Verify("v1.3.0", sums, Sign(otherPriv, "v1.3.0", sums)); err != nil {
		t.Fatal("the backup key failed:", err)
	}
	restore()
}

func TestRequired(t *testing.T) {
	restore := SetKeysForTest()
	if Required() {
		t.Fatal("no keys: verification must be off")
	}
	restore()
	pub, _, _ := ed25519.GenerateKey(nil)
	defer SetKeysForTest(pub)()
	if !Required() {
		t.Fatal("a key: verification must be on")
	}
}
