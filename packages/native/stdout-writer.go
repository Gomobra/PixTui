package native

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

const (
	// handoverMax is the most bytes one hand-over copies: the driver's drain budget.
	handoverMax = 64 * 1024
)

// WriterState is the single state word shared between owner and writer thread.
type WriterState uint32

const (
	stateIdle WriterState = iota
	stateInFlight
	stateWriting
	stateDoneOk
	stateDoneFailed
	stateStop
)

// SinkError is the error classification returned by a sink write.
type SinkError int

const (
	SinkWouldBlock SinkError = iota
	SinkInterrupted
	SinkFailed
)

// Outcome is the result of a collect.
type Outcome int

const (
	OutcomeBusy Outcome = iota
	OutcomeWritten
	OutcomeFailed
	OutcomeReclaimed
)

// Sink is the abstraction the writer writes to. The writer thread touches no
// context memory: it uses raw write(2) and poll(2) and its own memory.
type Sink interface {
	// Write writes bytes and returns (n, SinkError). n is the number of bytes
	// written; a non-nil SinkError classifies the failure.
	Write(p []byte) (int, SinkError)
	// WaitWritable waits until the sink can accept more writes.
	WaitWritable()
}

// FdSink writes to a file descriptor using raw write(2).
type FdSink struct {
	fd uintptr
}

// WaitWritable waits for the fd to become writable.
// On platforms without syscall.Poll this is a short sleep.
func (s FdSink) WaitWritable() {
	_ = s.fd
	time.Sleep(time.Millisecond)
}

// Writer is a thread-per-writer output handler. The owner hands over one copy
// at a time and collects it on a later call; collecting completes the copy's
// ticket.
type Writer struct {
	sink    Sink
	state   atomic.Uint32 // WriterState
	len     uint32
	done    uint32
	buffer  [handoverMax]byte
	wake    chan struct{} // signals the writer thread of a state change
	stop    chan struct{} // closed to request shutdown
	doneCh  chan struct{} // closed when the writer thread has exited
	started bool
}

// NewWriter creates a new Writer backed by the given sink.
func NewWriter(sink Sink) *Writer {
	return &Writer{
		sink:   sink,
		wake:   make(chan struct{}, 1),
		stop:   make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

// Start begins a hand-over: the first len bytes of buffer are the copy to
// write. The writer must be idle.
func (w *Writer) Start(data []byte) error {
	if len(data) == 0 || len(data) > handoverMax {
		return fmt.Errorf("invalid length %d", len(data))
	}
	if !w.state.CompareAndSwap(uint32(stateIdle), uint32(stateInFlight)) {
		return errors.New("writer not idle")
	}
	copy(w.buffer[:], data)
	atomic.StoreUint32(&w.len, uint32(len(data)))
	atomic.StoreUint32(&w.done, 0)
	// Wake the writer thread (non-blocking, buffered).
	select {
	case w.wake <- struct{}{}:
	default:
	}
	if !w.started {
		w.started = true
		go w.threadMain()
	}
	return nil
}

// Collect returns the result of the hand-over once the thread published it.
// Returns OutcomeBusy if the write is still in flight.
func (w *Writer) Collect() Outcome {
	switch WriterState(w.state.Load()) {
	case stateInFlight, stateWriting:
		return OutcomeBusy
	case stateDoneOk:
		w.state.Store(uint32(stateIdle))
		return OutcomeWritten
	case stateDoneFailed:
		w.state.Store(uint32(stateIdle))
		return OutcomeFailed
	default:
		return OutcomeBusy
	}
}

// WaitCollect waits for a running write without a hard limit. A copy that the
// thread has not picked up within grace is taken back unwritten (reclaimed).
func (w *Writer) WaitCollect(grace time.Duration) Outcome {
	deadline := time.Now().Add(grace)
	waitedInFlight := false
	for {
		state := WriterState(w.state.Load())
		switch state {
		case stateDoneOk, stateDoneFailed:
			return w.Collect()
		case stateInFlight:
			if waitedInFlight {
				if !w.state.CompareAndSwap(uint32(stateInFlight), uint32(stateStop)) {
					continue
				}
				return OutcomeReclaimed
			}
			waitedInFlight = true
			w.waitChange(state, grace)
		case stateWriting:
			w.waitChange(state, 0)
		default:
			return OutcomeBusy
		}
		if time.Now().After(deadline) && waitedInFlight {
			if w.state.CompareAndSwap(uint32(stateInFlight), uint32(stateStop)) {
				return OutcomeReclaimed
			}
		}
	}
}

func (w *Writer) waitChange(state WriterState, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-w.wake:
	case <-w.stop:
	case <-timer.C:
	}
}

// Abandon gives the writer up. Afterwards the owner must not use it.
func (w *Writer) Abandon() error {
	prev := w.state.Swap(uint32(stateStop))
	if prev == uint32(stateWriting) {
		// The thread is inside write(2); it will free itself on publish failure.
		close(w.stop)
		return nil
	}
	return w.release()
}

// release completes an abandon when the state is stop and the thread is not
// inside a write.
func (w *Writer) release() error {
	close(w.stop)
	<-w.doneCh
	return nil
}

// pick takes the posted copy. Returns false if the owner stopped first.
func (w *Writer) pick() bool {
	return w.state.CompareAndSwap(uint32(stateInFlight), uint32(stateWriting))
}

func (w *Writer) writeStepInternal() (more bool, written bool) {
	done := atomic.LoadUint32(&w.done)
	length := atomic.LoadUint32(&w.len)
	if done >= length {
		return false, false
	}
	rest := w.buffer[done:length]
	n, err := w.sink.Write(rest)
	switch err {
	case SinkInterrupted:
		return true, false
	case SinkWouldBlock:
		w.sink.WaitWritable()
		return true, false
	case SinkFailed:
		return false, false
	}
	if n == 0 || n > len(rest) {
		return false, false
	}
	atomic.AddUint32(&w.done, uint32(n))
	if atomic.LoadUint32(&w.done) == length {
		return false, true
	}
	return true, false
}

// publish publishes the result. Returns false if the owner abandoned.
func (w *Writer) publish(written bool) bool {
	result := WriterState(stateDoneOk)
	if !written {
		result = stateDoneFailed
	}
	if !w.state.CompareAndSwap(uint32(stateWriting), uint32(result)) {
		return false
	}
	select {
	case w.wake <- struct{}{}:
	default:
	}
	return true
}

func (w *Writer) threadMain() {
	defer close(w.doneCh)
	defer close(w.stop)
	for {
		state := WriterState(w.state.Load())
		for state == stateIdle || state == stateDoneOk || state == stateDoneFailed {
			w.waitChange(state, 0)
			state = WriterState(w.state.Load())
		}
		if state == stateStop || !w.pick() {
			return
		}
		more := true
		written := false
		for more && WriterState(w.state.Load()) != stateStop {
			more, written = w.writeStepInternal()
		}
		if !w.publish(written) {
			return
		}
	}
}

// StdoutWriter is a Writer backed by a real fd sink writing to fd 1.
type StdoutWriter struct {
	*Writer
}

// NewStdoutWriter creates a new StdoutWriter writing to fd 1 (stdout).
func NewStdoutWriter() *StdoutWriter {
	return &StdoutWriter{
		Writer: NewWriter(FdSink{fd: os.Stdout.Fd()}),
	}
}

// StdoutSupported reports whether a real-thread stdout writer is supported
// (the output is a terminal).
func StdoutSupported() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// PinLibrary pins the binary so a detached writer never returns into unmapped
// code. In a Go binary the executable is always resident, so this is a no-op
// returning true.
func PinLibrary() bool {
	return true
}

var _ = errors.New
var _ = time.Second
