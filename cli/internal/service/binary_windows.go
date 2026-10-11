package service

import "os"

func rootOnly(string, os.FileMode) error { return nil }

func UnsafeProgram(string) string { return "" }
