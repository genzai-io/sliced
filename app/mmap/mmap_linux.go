package mmap

func isRemapAvailable() bool {
	return true
}

//func mremap(m MMap, newSize int) ([]byte, error) {
//	_, _, errno := syscall.Syscall(syscall.SYS_MREMAP, uintptr(&m[0]), uintptr(len(m)), uintptr(newSize))
//	if errno != 0 {
//		return syscall.Errno(errno)
//	}
//	return nil, nil
//}
