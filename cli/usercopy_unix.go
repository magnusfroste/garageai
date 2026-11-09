//go:build !windows

package main

// refreshUserCopies is Windows only: elsewhere the self-update refreshes /usr/local/bin/garageai,
// and a copy in a user's directory updates itself the next time it runs (useServiceCopy).
func refreshUserCopies() {}

// writeOwnCopyRefusal: never as root, which would follow a symlink the user planted.
func writeOwnCopyRefusal(string) string {
	if isAdmin() {
		return "not as root; run it once without sudo to update it"
	}
	return ""
}
