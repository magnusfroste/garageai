//go:build !windows

package main

// refreshUserCopies is Windows only: elsewhere the self-update refreshes /usr/local/bin/garageai,
// and a copy in a user's directory updates itself the next time it runs (useServiceCopy).
func refreshUserCopies() {}
