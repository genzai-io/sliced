package mmap

import (
	"os"
	"syscall"
)

// Fdatasync flushes written data to a file descriptor.
func Fdatasync(file *os.File) error {
	return syscall.Fdatasync(int(file.Fd()))
}
