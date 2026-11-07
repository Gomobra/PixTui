//go:build windows

package native

import (
	"os"
	"time"
)

// getTermSize gets terminal dimensions on Windows using a simple approach.
func (t *Terminal) getTermSize() (int, int) {
	// On Windows, use a basic fallback
	return 80, 24
}

// WatchResize watches for terminal resize events on Windows using polling.
func (t *Terminal) WatchResize(callback func(ResizeEvent)) {
	t.StopResizeWatch()

	t.stopResize = make(chan struct{})
	t.resizeDone = make(chan struct{})

	go func() {
		defer close(t.resizeDone)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-t.stopResize:
				return
			case <-ticker.C:
				w, h := t.getTermSize()
				if w > 0 && h > 0 {
					t.SetDimensions(w, h)
					callback(ResizeEvent{Width: w, Height: h})
				}
			}
		}
	}()
	_ = os.Stdout
}
