//go:build windows

package native

import "os"

// Write writes bytes to the fd using os.File on Windows.
func (s FdSink) Write(p []byte) (int, SinkError) {
	f := os.NewFile(uintptr(s.fd), "stdout")
	if f == nil {
		return 0, SinkFailed
	}
	n, err := f.Write(p)
	if err != nil {
		return n, SinkFailed
	}
	return n, 0
}
