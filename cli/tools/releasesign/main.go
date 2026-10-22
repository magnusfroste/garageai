// releasesign creates the release signing key and signs Bridge releases (docs/release-signing.md).
// It runs on the release manager's own computer, never in CI or on the gateway: the private key
// must not exist anywhere an attacker of GitHub or the gateway could reach.
//
//	go run ./tools/releasesign keygen -out ~/garageai-release.key
//	go run ./tools/releasesign sign -key ~/garageai-release.key -version v1.3.0 SHA256SUMS > SHA256SUMS.sig
//	go run ./tools/releasesign verify -pub <base64> -version v1.3.0 SHA256SUMS SHA256SUMS.sig
package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/magnusfroste/garageai/cli/internal/release"
)

const keyHeader = "garageai-release-private-key 1"

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "keygen":
		err = keygen(os.Args[2:])
	case "sign":
		err = sign(os.Args[2:])
	case "verify":
		err = verify(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "releasesign:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: releasesign keygen -out FILE | sign -key FILE -version vX.Y.Z SHA256SUMS | verify -pub BASE64 -version vX.Y.Z SHA256SUMS SHA256SUMS.sig")
	os.Exit(2)
}

func keygen(args []string) error {
	fs := flag.NewFlagSet("keygen", flag.ExitOnError)
	out := fs.String("out", "", "where to write the private key (0600); must not exist")
	fs.Parse(args)
	if *out == "" {
		return errors.New("-out is required")
	}
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "%s\n%s\n", keyHeader, base64.StdEncoding.EncodeToString(priv.Seed())); err != nil {
		return err
	}
	fmt.Printf("Private key written to %s (keep it off GitHub and off the gateway; keep a copy in your password manager).\n", *out)
	fmt.Printf("Public key (safe to share; it goes into cli/internal/release/release.go, trustedKeys):\n%s\n", base64.StdEncoding.EncodeToString(pub))
	fmt.Printf("Key id: %s\n", release.KeyID(pub))
	return nil
}

func readKey(path string) (ed25519.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 2 || strings.TrimSpace(lines[0]) != keyHeader {
		return nil, errors.New(path + " is not a GarageAI release key")
	}
	seed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errors.New(path + ": malformed key")
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func sign(args []string) error {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	key := fs.String("key", "", "the private key file")
	version := fs.String("version", "", "the release version, for example v1.3.0")
	fs.Parse(args)
	if *key == "" || *version == "" || fs.NArg() != 1 {
		return errors.New("sign needs -key, -version and the SHA256SUMS file")
	}
	priv, err := readKey(*key)
	if err != nil {
		return err
	}
	sums, err := os.ReadFile(fs.Arg(0))
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(release.Sign(priv, *version, sums))
	return err
}

func verify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	pub := fs.String("pub", "", "the public key (base64)")
	version := fs.String("version", "", "the release version")
	fs.Parse(args)
	if *pub == "" || *version == "" || fs.NArg() != 2 {
		return errors.New("verify needs -pub, -version, SHA256SUMS and SHA256SUMS.sig")
	}
	pk, err := base64.StdEncoding.DecodeString(*pub)
	if err != nil || len(pk) != ed25519.PublicKeySize {
		return errors.New("malformed public key")
	}
	sums, err := os.ReadFile(fs.Arg(0))
	if err != nil {
		return err
	}
	sig, err := os.ReadFile(fs.Arg(1))
	if err != nil {
		return err
	}
	defer release.SetKeysForTest(ed25519.PublicKey(pk))()
	if err := release.Verify(*version, sums, sig); err != nil {
		return err
	}
	fmt.Println("OK: signed by", release.KeyID(pk), "for", *version)
	return nil
}
