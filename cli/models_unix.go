//go:build !windows

package main

import "syscall"

// freeGB is the free space where Ollama keeps models (the root filesystem, or the home volume: on
// the machines we have seen they are the same). 0 when unknown.
func freeGB() float64 {
	var st syscall.Statfs_t
	if err := syscall.Statfs("/", &st); err != nil {
		return 0
	}
	return float64(st.Bavail) * float64(st.Bsize) / 1e9
}
