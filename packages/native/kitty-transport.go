package native

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Mode is the kitty graphics transport mode.
type Mode int

const (
	ModeRaw Mode = iota
	ModeZlib
	ModeFile
)

// FileState is the state of the file-based transport.
type FileState int

const (
	FileStateDisabled FileState = iota
	FileStateProbing
	FileStateReady
	FileStateUnsupported
	FileStateTimeout
	FileStateIOError
	FileStateCancelled
)

// Fallback is the fallback reason when file transmission cannot proceed.
type Fallback int

const (
	FallbackNone Fallback = iota
	FallbackNotReady
	FallbackUnavailable
	FallbackBudget
	FallbackBusy
	FallbackPreparation
	FallbackCompression
)

const (
	leasesMax = 8
	bytesMax  = 64 * 1024 * 1024 // 64 MiB per renderer
	timeoutNS = 5 * time.Second
)

// Lease tracks a file leased to the terminal for kitty file-based transfer.
type Lease struct {
	ID          uint32
	Path        string
	PathLen     int
	Size        int64
	DeadlineNS  int64
	Unpublished bool
}

// Release deletes the leased file and clears the lease.
func (l *Lease) Release() error {
	if l.PathLen == 0 {
		return nil
	}
	err := os.Remove(l.Path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	l.ID = 0
	l.Path = ""
	l.PathLen = 0
	l.Size = 0
	l.DeadlineNS = 0
	l.Unpublished = false
	return nil
}

// Transport manages kitty graphics image transmission with optional file-based
// leasing and capability probing. Mirrors packages/native/src/kitty-transport.zig.
type Transport struct {
	mu            sync.Mutex
	mode          Mode
	effective     Mode
	fileState     FileState
	fallback      Fallback
	leases        [leasesMax]Lease
	queryID       uint32
	uploadProbeID uint32
	queryOK       bool
	uploadOK      bool
	retryImages   bool
}

// NewTransport creates a new Transport in the given mode.
func NewTransport(mode Mode) *Transport {
	return &Transport{
		mode:      mode,
		effective: ModeRaw,
		fileState: FileStateDisabled,
		fallback:  FallbackNone,
	}
}

// PendingBytes returns the total bytes of all active leases.
func (t *Transport) PendingBytes() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	var total int64
	for _, l := range t.leases {
		total += l.Size
	}
	return total
}

// PendingCount returns the number of active leases.
func (t *Transport) PendingCount() uint32 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.pendingCountLocked()
}

func (t *Transport) pendingCountLocked() uint32 {
	var count uint32
	for _, l := range t.leases {
		if l.PathLen != 0 {
			count++
		}
	}
	return count
}

// Cancel cancels all leases and transitions the file state.
func (t *Transport) Cancel(reason FileState) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cancelLocked(reason)
}

func (t *Transport) cancelLocked(reason FileState) {
	t.retryImages = t.retryImages || t.fileState == FileStateReady
	for i := range t.leases {
		_ = t.leases[i].Release()
	}
	if t.mode == ModeFile || t.fileState != FileStateDisabled {
		if t.pendingCountLocked() == 0 {
			t.fileState = reason
		} else {
			t.fileState = FileStateIOError
		}
	}
}

// FinishFrame releases leases of skipped frames when not published.
func (t *Transport) FinishFrame(published bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i := range t.leases {
		if t.leases[i].Unpublished && !published {
			_ = t.leases[i].Release()
		}
		t.leases[i].Unpublished = false
	}
}

// Discard releases the unpublished lease with the given id.
func (t *Transport) Discard(id uint32) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i := range t.leases {
		if t.leases[i].Unpublished && t.leases[i].ID == id {
			_ = t.leases[i].Release()
		}
	}
}

// Arm arms new leases that have no deadline yet.
func (t *Transport) Arm(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.armLocked(now)
}

func (t *Transport) armLocked(now time.Time) {
	deadline := now.Add(timeoutNS).UnixNano()
	for i := range t.leases {
		l := &t.leases[i]
		if l.PathLen != 0 && l.DeadlineNS == 0 {
			l.DeadlineNS = deadline
		}
	}
}

// Expire arms new leases and expires any that have passed their deadline.
func (t *Transport) Expire(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.armLocked(now)
	for i := range t.leases {
		l := &t.leases[i]
		if l.PathLen == 0 {
			continue
		}
		if now.UnixNano() >= l.DeadlineNS {
			t.cancelLocked(FileStateTimeout)
			return
		}
	}
}

// CreateLease creates a new lease file for the given image in the directory.
func (t *Transport) CreateLease(id uint32, image *Image, directory string) (*Lease, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.createLeaseLocked(id, image, directory)
}

func (t *Transport) createLeaseLocked(id uint32, image *Image, directory string) (*Lease, error) {
	if !filepath.IsAbs(directory) {
		return nil, errors.New("invalid directory")
	}
	size := int64(image.Width * image.Height * 4)
	if size > bytesMax-int64(t.PendingBytes()) {
		return nil, errors.New("budget exceeded")
	}
	var slot *Lease
	for i := range t.leases {
		if t.leases[i].PathLen != 0 && t.leases[i].ID == id {
			return nil, errors.New("busy")
		}
		if t.leases[i].PathLen == 0 {
			slot = &t.leases[i]
		}
	}
	if slot == nil {
		return nil, errors.New("budget exceeded")
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, fmt.Sprintf("pixtui-kitty-%s", hex.EncodeToString(random)))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, err
	}
	pngData, err := image.PNGBytes()
	if err != nil {
		file.Close()
		os.Remove(path)
		return nil, err
	}
	if _, err := file.Write(pngData); err != nil {
		file.Close()
		os.Remove(path)
		return nil, err
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return nil, err
	}
	slot.ID = id
	slot.Path = path
	slot.PathLen = len(path)
	slot.Size = int64(len(pngData))
	slot.Unpublished = false
	return slot, nil
}

// WriteReference writes a kitty file-reference escape sequence for the lease.
func (t *Transport) WriteReference(w *bytes.Buffer, lease *Lease, image *Image, id uint32, action byte) error {
	format := uint32(32)
	if image.Width > 0 && image.Width%2 == 0 {
		format = 24
	}
	payload := base64.StdEncoding.EncodeToString([]byte(lease.Path))
	fmt.Fprintf(w, "\x1b_Ga=%c,t=f,f=%d,s=%d,v=%d,i=%d,q=0;%s\x1b\\",
		action, format, image.Width, image.Height, id, payload)
	return nil
}

// StartProbe begins file-mode capability probing.
func (t *Transport) StartProbe(w *bytes.Buffer, firstID uint32, directory string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.mode != ModeFile || t.fileState != FileStateDisabled {
		return nil
	}
	image := NewImage(1, 1)
	image.Fill(0, 0, 0, 255)
	t.fileState = FileStateProbing
	t.queryOK = false
	t.uploadOK = false
	t.queryID = firstID
	t.uploadProbeID = firstID + 1
	lease, err := t.createLeaseLocked(t.uploadProbeID, image, directory)
	if err != nil {
		t.cancelLocked(FileStateIOError)
		return err
	}
	if err := t.WriteReference(w, lease, image, t.queryID, 'q'); err != nil {
		return err
	}
	if err := t.WriteReference(w, lease, image, t.uploadProbeID, 't'); err != nil {
		return err
	}
	_ = t.writeKittyDelete(w, t.uploadProbeID, true, false)
	return nil
}

// HandleReply processes a kitty response escape sequence.
func (t *Transport) HandleReply(response []byte) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.fileState == FileStateDisabled || len(response) > 4096 {
		return false
	}
	if !bytes.HasPrefix(response, []byte("\x1b_G")) || !bytes.HasSuffix(response, []byte("\x1b\\")) {
		return false
	}
	sep := bytes.IndexByte(response, ';')
	if sep < 0 {
		return false
	}
	fields := bytes.Split(response[3:sep], []byte{','})
	var id *uint32
	for _, f := range fields {
		if !bytes.HasPrefix(f, []byte("i=")) {
			continue
		}
		if id != nil {
			return false
		}
		var val uint32
		_, err := fmt.Sscanf(string(f[2:]), "%d", &val)
		if err != nil {
			return false
		}
		id = &val
	}
	if id == nil {
		return false
	}
	imageID := *id
	ok := bytes.Equal(response[sep+1:len(response)-2], []byte("OK"))
	if imageID == t.queryID || imageID == t.uploadProbeID {
		if t.fileState != FileStateProbing {
			return true
		}
		if !ok {
			t.cancelLocked(FileStateUnsupported)
		} else {
			if imageID == t.queryID {
				t.queryOK = true
			} else {
				t.uploadOK = true
			}
			if t.queryOK && t.uploadOK {
				for i := range t.leases {
					_ = t.leases[i].Release()
				}
				if t.pendingCountLocked() == 0 {
					t.fileState = FileStateReady
				} else {
					t.fileState = FileStateIOError
				}
				if t.fileState == FileStateReady && t.mode == ModeFile {
					t.retryImages = true
				}
			}
		}
		return true
	}
	for i := range t.leases {
		l := &t.leases[i]
		if l.PathLen == 0 || l.ID != imageID {
			continue
		}
		if ok {
			_ = l.Release()
			if l.PathLen != 0 {
				t.cancelLocked(FileStateIOError)
			}
		} else {
			t.cancelLocked(FileStateUnsupported)
		}
		return true
	}
	return false
}

// writeKittyDelete writes a kitty delete escape sequence.
func (t *Transport) writeKittyDelete(w *bytes.Buffer, id uint32, final, forceDelete bool) error {
	_ = final
	_ = forceDelete
	fmt.Fprintf(w, "\x1b_Ga=d,i=%d\x1b\\", id)
	return nil
}

// Transmit sends an image over the transport.
func (t *Transport) Transmit(w *bytes.Buffer, image *Image, id uint32, tmux bool, directory string) error {
	t.mu.Lock()
	mode := t.mode
	fileState := t.fileState
	t.fallback = FallbackNone
	t.mu.Unlock()

	if mode == ModeFile {
		if fileState != FileStateReady || tmux {
			t.mu.Lock()
			if fileState == FileStateProbing {
				t.fallback = FallbackNotReady
			} else {
				t.fallback = FallbackUnavailable
			}
			t.mu.Unlock()
			return nil
		}
		lease, err := t.CreateLease(id, image, directory)
		if err != nil {
			t.mu.Lock()
			switch {
			case err.Error() == "budget exceeded":
				t.fallback = FallbackBudget
			case err.Error() == "busy":
				t.fallback = FallbackBusy
			default:
				t.fallback = FallbackPreparation
				if t.fallback == FallbackPreparation {
					t.cancelLocked(FileStateIOError)
				}
			}
			t.mu.Unlock()
			return nil
		}
		lease.Unpublished = true
		t.mu.Lock()
		t.effective = ModeFile
		t.mu.Unlock()
		return t.WriteReference(w, lease, image, id, 't')
	}
	escape, err := image.KittyEscapeWithOptions(KittyOptions{
		Action:     "T",
		Format:     32,
		Compressed: mode == ModeZlib,
	})
	if err != nil {
		return err
	}
	w.WriteString(escape)
	t.mu.Lock()
	t.effective = ModeRaw
	if mode == ModeZlib && t.effective == ModeRaw {
		t.fallback = FallbackCompression
	}
	t.mu.Unlock()
	return nil
}

// Mode returns the configured transport mode.
func (t *Transport) Mode() Mode {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.mode
}

// Effective returns the effective encoding mode after the last transmit.
func (t *Transport) Effective() Mode {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.effective
}

// FileState returns the current file state.
func (t *Transport) FileState() FileState {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.fileState
}

// Fallback returns the current fallback reason.
func (t *Transport) Fallback() Fallback {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.fallback
}
