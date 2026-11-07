//go:build !windows

package native

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// getTermSize gets terminal dimensions using ioctl TIOCGWINSZ
func (t *Terminal) getTermSize() (int, int) {
	var dims [4]uint16
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		t.stdoutFd,
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&dims[0])),
		0, 0, 0,
	)
	if errno != 0 {
		return 0, 0
	}
	return int(dims[1]), int(dims[0]) // cols, rows
}

// WatchResize watches for terminal resize events using the SIGWINCH signal.
func (t *Terminal) WatchResize(callback func(ResizeEvent)) {
	t.StopResizeWatch()

	t.stopResize = make(chan struct{})
	t.resizeDone = make(chan struct{})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	t.resizeSignal = sigCh

	go func() {
		defer close(t.resizeDone)
		for {
			select {
			case <-t.stopResize:
				signal.Stop(sigCh)
				return
			case <-sigCh:
				w, h := t.getTermSize()
				if w > 0 && h > 0 {
					t.SetDimensions(w, h)
					callback(ResizeEvent{Width: w, Height: h})
				}
			}
		}
	}()
}
