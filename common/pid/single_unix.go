// +build linux solaris darwin freebsd

package single

import (
	"os"
	"syscall"
	"fmt"
)

// Lock tries to obtain an exclude lock on a lockfile and exits the program if an error occurs
func (s *Single) Lock() LockResult {
	// open/create lock file
	f, err := os.OpenFile(s.Filename(), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return LockResult{
			Success: false,
			Err:     err,
		}
	}
	s.file = f
	// set the lock type to F_WRLCK, therefor the file has to be opened writable
	flock := syscall.Flock_t{
		Type: syscall.F_WRLCK,
		Pid:  int32(os.Getpid()),
	}
	// try to obtain an exclusive lock - FcntlFlock seems to be the portable *ix way
	if err := syscall.FcntlFlock(s.file.Fd(), syscall.F_SETLK, &flock); err != nil {
		return readPIDFile(s.Filename())
	}

	_, err = f.Write([]byte(fmt.Sprintf("%d\n%d", os.Getpid(), 0)))
	if err != nil {
		s.Unlock()
		return LockResult{
			Success: false,
			Err:     err,
		}
	}

	return LockResult{
		Success: true,
		Pid:     os.Getpid(),
		Port:    0,
	}
}

// Unlock releases the lock, closes and removes the lockfile. All errors will be reported directly.
func (s *Single) Unlock() error {
	// set the lock type to F_UNLCK
	flock := syscall.Flock_t{
		Type: syscall.F_UNLCK,
		Pid:  int32(os.Getpid()),
	}
	if err := syscall.FcntlFlock(s.file.Fd(), syscall.F_SETLK, &flock); err != nil {
		return err
	}
	if err := s.file.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.Filename()); err != nil {
		return err
	}
	return nil
}
