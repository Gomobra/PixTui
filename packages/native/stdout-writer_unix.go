//go:build !windows

package native

import (
	"syscall"
	"unsafe"
)

// Write writes bytes to the fd using raw write(2).
func (s FdSink) Write(p []byte) (int, SinkError) {
	ptr := unsafe.Pointer(unsafe.SliceData(p))
	n, _, errno := syscall.Syscall(syscall.SYS_WRITE, s.fd, uintptr(ptr), uintptr(len(p)))
	if errno != 0 {
		switch errno {
		case syscall.EINTR:
			return 0, SinkInterrupted
		case syscall.EAGAIN:
			return 0, SinkWouldBlock
		default:
			return 0, SinkFailed
		}
	}
	return int(n), 0
}
