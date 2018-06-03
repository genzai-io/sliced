// +build !windows,!plan9,!linux,!openbsd

package mmap

import "os"

// Fdatasync flushes written data to a file descriptor.
func Fdatasync(file *os.File) error {
	return file.Sync()
}
