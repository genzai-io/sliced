// package single provides a mechanism to ensure, that only one instance of a program is running

package single

import (
	"errors"
	"os"
	"io/ioutil"
	"strings"
	"strconv"
	"time"
)

var (
	// ErrAlreadyRunning
	ErrAlreadyRunning = errors.New("the program is already running")
	ErrPIDInvalid     = errors.New("PID in lockfile is invalid")
	//
	Lockfile string
)

// Single represents the name and the open file descriptor
type Single struct {
	name string
	pid  int
	port int
	file *os.File
}

type LockResult struct {
	Success bool
	Err     error
	Pid     int
	Port    int
}

// New creates a Single instance
func New(name string) *Single {
	return &Single{name: name}
}

func readPIDFile(name string) LockResult {
	for i := 0; i < 10; i++ {
		contents, err := ioutil.ReadFile(name)
		if err != nil {
			return LockResult{
				Success: false,
				Err:     err,
			}
		}
		if len(contents) == 0 {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		split := strings.Split(string(contents), "\n")
		splitLen := len(split)

		if splitLen == 0 {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if splitLen == 1 {
			pid, err := strconv.Atoi(split[0])
			if err != nil {
				return LockResult{
					Success: false,
					Err:     ErrPIDInvalid,
				}
			}
			return LockResult{
				Success: false,
				Pid:     pid,
			}
		} else {
			pid, err := strconv.Atoi(split[0])
			if err != nil {
				return LockResult{
					Success: false,
					Err:     ErrPIDInvalid,
				}
			}
			port, err := strconv.Atoi(split[1])
			return LockResult{
				Success: false,
				Pid:     pid,
				Port:    port,
			}
		}
	}

	return LockResult{
		Success: false,
	}
}
