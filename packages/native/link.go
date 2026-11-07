package native

import (
	"errors"
	"sync"
	"unsafe"
)

// LinkPool errors mirroring link.zig.
var (
	ErrLinkOutOfMemory      = errors.New("link pool out of memory")
	ErrLinkInvalidID        = errors.New("link pool invalid id")
	ErrLinkWrongGeneration  = errors.New("link pool wrong generation")
	ErrLinkRefcountOverflow = errors.New("link pool refcount overflow")
	ErrLinkURLTooLong       = errors.New("link pool url too long")
)

const (
	genBits    = 8
	slotBits   = 16
	genMask    = (1 << genBits) - 1
	slotMask   = (1 << slotBits) - 1
	maxURLLen  = 512
	retiredGen = genMask + 1
)

// idPayload is the packed link id: [generation (8 bits) | slot_index (16 bits)]
type idPayload uint32

// slotHeader is the per-slot header stored in the slot store.
type slotHeader struct {
	len        uint32
	refcount   uint32
	generation uint32
}

const slotHeaderSize = uintptr(unsafe.Sizeof(slotHeader{}))

// LinkPool stores URL strings with reusable, generation-tagged ids.
// Mirrors packages/native/src/link.zig.
type LinkPool struct {
	mu               sync.RWMutex
	slotCapacity     uint32
	slotsPerPage     uint32
	slotSizeBytes    int
	slots            []byte
	freeList         []uint32
	numSlots         uint32
	retiredSlotCount uint32
	interned         map[string]idPayload // url -> live id
}

// NewLinkPool creates a new LinkPool.
func NewLinkPool() *LinkPool {
	return &LinkPool{
		slotCapacity:  maxURLLen,
		slotsPerPage:  64,
		slotSizeBytes: int(slotHeaderSize) + maxURLLen,
		slots:         make([]byte, 0),
		freeList:      make([]uint32, 0),
		interned:      make(map[string]idPayload),
	}
}

// grow adds a new page of slots to the pool.
func (p *LinkPool) grow() error {
	slotCountMax := uint32(slotMask + 1)
	if p.numSlots > slotCountMax-p.slotsPerPage {
		return ErrLinkOutOfMemory
	}
	newNumSlots := p.numSlots + p.slotsPerPage
	addBytes := p.slotSizeBytes * int(p.slotsPerPage)
	if cap(p.slots) < len(p.slots)+addBytes {
		newSlots := make([]byte, len(p.slots), len(p.slots)+addBytes)
		copy(newSlots, p.slots)
		p.slots = newSlots
	}
	p.slots = p.slots[:len(p.slots)+addBytes]
	if cap(p.freeList) < int(newNumSlots) {
		newFree := make([]uint32, len(p.freeList), newNumSlots)
		copy(newFree, p.freeList)
		p.freeList = newFree
	}
	p.freeList = p.freeList[:len(p.freeList)]
	for i := uint32(0); i < p.slotsPerPage; i++ {
		p.freeList = append(p.freeList, p.numSlots+i)
	}
	p.numSlots = newNumSlots
	return nil
}

func (p *LinkPool) slotPtr(slotIndex uint32) []byte {
	offset := int(slotIndex) * p.slotSizeBytes
	return p.slots[offset : offset+p.slotSizeBytes : offset+p.slotSizeBytes]
}

func slotHeaderFromBytes(b []byte) *slotHeader {
	return (*slotHeader)(unsafe.Pointer(&b[0]))
}

func packID(slotIndex, generation uint32) (idPayload, error) {
	if slotIndex > slotMask {
		return 0, ErrLinkOutOfMemory
	}
	return idPayload((generation&genMask)<<slotBits | (slotIndex & slotMask)), nil
}

func unpackID(id idPayload) (slotIndex, generation uint32) {
	slotIndex = uint32(id) & slotMask
	generation = (uint32(id) >> slotBits) & genMask
	return
}

// internLiveID records url as the live identity for id (put-no-clobber).
func (p *LinkPool) internLiveID(id idPayload, url string) {
	if _, exists := p.interned[url]; !exists {
		p.interned[url] = id
	}
}

// removeInternedLiveID removes the url's live id entry if it matches id.
func (p *LinkPool) removeInternedLiveID(url string, id idPayload) {
	if cur, ok := p.interned[url]; ok && cur == id {
		delete(p.interned, url)
	}
}

// Intern registers a URL and returns its unique link ID (compatibility API).
func (p *LinkPool) Intern(url string) uint32 {
	id, err := p.Acquire(url)
	if err != nil {
		return 0
	}
	return uint32(id)
}

// Lookup returns the URL for the given link ID (compatibility API).
func (p *LinkPool) Lookup(id uint32) string {
	url, err := p.Get(idPayload(id))
	if err != nil {
		return ""
	}
	return url
}

// Acquire returns one owned live reference for url, interning it if needed.
func (p *LinkPool) Acquire(url string) (idPayload, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(url) > maxURLLen {
		return 0, ErrLinkURLTooLong
	}
	if id, ok := p.lookupLive(url); ok {
		if err := p.increfLocked(id); err != nil {
			return 0, err
		}
		return id, nil
	}
	return p.acquireNewLocked(url)
}

func (p *LinkPool) lookupLive(url string) (idPayload, bool) {
	id, ok := p.interned[url]
	if !ok {
		return 0, false
	}
	if _, err := p.getLocked(id); err != nil {
		return 0, false
	}
	return id, true
}

func (p *LinkPool) acquireNewLocked(url string) (idPayload, error) {
	if len(p.freeList) == 0 {
		if err := p.grow(); err != nil {
			return 0, err
		}
	}
	slotIndex := p.freeList[len(p.freeList)-1]
	p.freeList = p.freeList[:len(p.freeList)-1]
	b := p.slotPtr(slotIndex)
	h := slotHeaderFromBytes(b)
	if h.generation >= genMask {
		return 0, ErrLinkOutOfMemory
	}
	newGen := h.generation + 1
	h.len = uint32(len(url))
	h.refcount = 0
	h.generation = newGen
	dataPtr := b[slotHeaderSize:]
	copy(dataPtr[:len(url)], url)

	id, err := packID(slotIndex, newGen)
	if err != nil {
		return 0, err
	}
	p.internLiveID(id, url)
	h.refcount = 1
	return id, nil
}

func (p *LinkPool) releaseSlot(slotIndex, expectedGen uint32) {
	b := p.slotPtr(slotIndex)
	h := slotHeaderFromBytes(b)
	if h.generation != expectedGen {
		return
	}
	if h.refcount != 0 {
		return
	}
	if h.generation == genMask {
		h.generation = retiredGen
		p.retiredSlotCount++
	} else {
		p.freeList = append(p.freeList, slotIndex)
	}
}

// Incref retains an already-live id.
func (p *LinkPool) Incref(id idPayload) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.increfLocked(id)
}

func (p *LinkPool) increfLocked(id idPayload) error {
	slotIndex, generation := unpackID(id)
	if slotIndex >= p.numSlots {
		return ErrLinkInvalidID
	}
	b := p.slotPtr(slotIndex)
	h := slotHeaderFromBytes(b)
	if h.generation != generation {
		return ErrLinkWrongGeneration
	}
	if h.refcount == 0 {
		return ErrLinkInvalidID
	}
	if h.refcount == ^uint32(0) {
		return ErrLinkRefcountOverflow
	}
	h.refcount++
	return nil
}

// Decref releases one reference to id.
func (p *LinkPool) Decref(id idPayload) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	slotIndex, generation := unpackID(id)
	if slotIndex >= p.numSlots {
		return ErrLinkInvalidID
	}
	b := p.slotPtr(slotIndex)
	h := slotHeaderFromBytes(b)
	if h.refcount == 0 {
		return ErrLinkInvalidID
	}
	if h.generation != generation {
		return ErrLinkWrongGeneration
	}
	if h.refcount == 1 {
		url, err := p.getLocked(id)
		if err == nil {
			p.removeInternedLiveID(url, id)
		}
	}
	h.refcount--
	if h.refcount == 0 {
		p.releaseSlot(slotIndex, generation)
	}
	return nil
}

// Get returns the URL for id.
func (p *LinkPool) Get(id idPayload) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getLocked(id)
}

func (p *LinkPool) getLocked(id idPayload) (string, error) {
	slotIndex, generation := unpackID(id)
	if slotIndex >= p.numSlots {
		return "", ErrLinkInvalidID
	}
	b := p.slotPtr(slotIndex)
	h := slotHeaderFromBytes(b)
	if h.generation != generation {
		return "", ErrLinkWrongGeneration
	}
	dataPtr := b[slotHeaderSize:]
	return string(dataPtr[:h.len]), nil
}

// GetRefcount returns the current refcount for id.
func (p *LinkPool) GetRefcount(id idPayload) (uint32, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	slotIndex, generation := unpackID(id)
	if slotIndex >= p.numSlots {
		return 0, ErrLinkInvalidID
	}
	b := p.slotPtr(slotIndex)
	h := slotHeaderFromBytes(b)
	if h.generation != generation {
		return 0, ErrLinkWrongGeneration
	}
	return h.refcount, nil
}

// GetTotalSlots returns the total number of slots allocated.
func (p *LinkPool) GetTotalSlots() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.numSlots
}

// GetFreeSlotCount returns the number of free slots.
func (p *LinkPool) GetFreeSlotCount() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return uint32(len(p.freeList))
}

// GetLiveSlotCount returns the number of live (non-free, non-retired) slots.
func (p *LinkPool) GetLiveSlotCount() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.numSlots - uint32(len(p.freeList)) - p.retiredSlotCount
}

// Close releases all interned live ids and frees the slot store.
func (p *LinkPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for url := range p.interned {
		delete(p.interned, url)
	}
	p.slots = nil
	p.freeList = nil
	p.numSlots = 0
	p.retiredSlotCount = 0
	return nil
}

// LinkTracker tracks link usage per buffer with per-cell refcounting.
type LinkTracker struct {
	mu   sync.Mutex
	pool *LinkPool
	used map[idPayload]uint32 // id -> cell count
}

// NewLinkTracker creates a new LinkTracker backed by the given pool.
func NewLinkTracker(pool *LinkPool) *LinkTracker {
	return &LinkTracker{
		pool: pool,
		used: make(map[idPayload]uint32),
	}
}

func (t *LinkTracker) decRefAll() {
	for id := range t.used {
		_ = t.pool.Decref(id)
	}
	t.used = make(map[idPayload]uint32)
}

// Close releases all tracked ids.
func (t *LinkTracker) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.decRefAll()
	return nil
}

// Clear releases all tracked ids.
func (t *LinkTracker) Clear() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.decRefAll()
	return nil
}

// TrackUrl tracks a URL once per document, independent of cell count.
func (t *LinkTracker) TrackUrl(url string) (idPayload, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	id, err := t.pool.Acquire(url)
	if err != nil {
		return 0, err
	}
	if _, exists := t.used[id]; exists {
		return id, nil
	}
	t.used[id] = 1
	return id, nil
}

// AddCellRef adds a cell reference for id.
func (t *LinkTracker) AddCellRef(id idPayload) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	count, exists := t.used[id]
	if !exists {
		if err := t.pool.Incref(id); err != nil {
			return err
		}
		t.used[id] = 1
		return nil
	}
	t.used[id] = count + 1
	return nil
}

// RemoveCellRef removes a cell reference for id.
func (t *LinkTracker) RemoveCellRef(id idPayload) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	count, ok := t.used[id]
	if !ok || count == 0 {
		return nil
	}
	count--
	if count == 0 {
		delete(t.used, id)
		_ = t.pool.Decref(id)
		return nil
	}
	t.used[id] = count
	return nil
}

// HasAny returns true if any links are tracked.
func (t *LinkTracker) HasAny() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.used) > 0
}

// GetLinkCount returns the number of distinct tracked links.
func (t *LinkTracker) GetLinkCount() uint32 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return uint32(len(t.used))
}
