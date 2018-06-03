package fs

/*
h := syscall.MustLoadDLL("kernel32.dll")
c := h.MustFindProc("GetDiskFreeSpaceExW")

var freeBytes int64

_, _, err := c.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(wd))),
    uintptr(unsafe.Pointer(&freeBytes)), nil, nil)
 */