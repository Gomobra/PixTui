package native

import (
	"errors"
	"sync"
	"unsafe"
)

// GraphemePool errors mirroring grapheme.zig.
var (
	ErrGraphemeOutOfMemory      = errors.New("grapheme pool out of memory")
	ErrGraphemeInvalidID        = errors.New("grapheme pool invalid id")
	ErrGraphemeWrongGeneration  = errors.New("grapheme pool wrong generation")
	ErrGraphemeRefcountOverflow = errors.New("grapheme pool refcount overflow")
	ErrGraphemeTooLong          = errors.New("grapheme too long")
)

const (
	gpGenBits        = 7
	gpSlotBits       = 16
	gpClassBits      = 3
	gpGenMask        = (1 << gpGenBits) - 1
	gpSlotMask       = (1 << gpSlotBits) - 1
	gpClassMask      = (1 << gpClassBits) - 1
	gpMaxGraphemeLen = 128
	gpIdPayloadBits  = gpClassBits + gpGenBits + gpSlotBits // 26

	// IMAGE_ID_MASK is the mask for image IDs packed into grapheme IDs.
	// Mirrors grapheme.zig IMAGE_ID_MASK.
	IMAGE_ID_MASK = 0x03FF_FFFF
)

// gpIdPayload is the packed grapheme id: [class (3 bits) | generation (7 bits) | slot_index (16 bits)]
type gpIdPayload uint32

// gpSlotHeader is the per-slot header stored in the slot store.
type gpSlotHeader struct {
	len         uint16
	refcount    uint32
	generation  uint32
	isAllocated uint32
}

const gpSlotHeaderSize = uintptr(unsafe.Sizeof(gpSlotHeader{}))

// Class sizes (max bytes per grapheme) and default slots per page per class.
var gpClassSizes = []uint32{8, 16, 32, 64, 128}
var gpDefaultSlotsPerPage = []uint32{256, 128, 64, 16, 8}

// GraphemePool manages grapheme clusters with slot-based generation-tagged IDs.
// Mirrors packages/native/src/grapheme.zig.
type GraphemePool struct {
	mu            sync.RWMutex
	slotCapacity  uint32
	slotsPerPage  uint32
	slotSizeBytes int
	slots         []byte
	freeList      []uint32
	numSlots      uint32
	interned      map[string]gpIdPayload // grapheme -> live id
}

// NewGraphemePool creates a new GraphemePool.
func NewGraphemePool() *GraphemePool {
	return &GraphemePool{
		slotCapacity:  gpMaxGraphemeLen,
		slotsPerPage:  gpDefaultSlotsPerPage[0],
		slotSizeBytes: int(gpSlotHeaderSize) + gpMaxGraphemeLen,
		slots:         make([]byte, 0),
		freeList:      make([]uint32, 0),
		interned:      make(map[string]gpIdPayload),
	}
}

// grow adds a new page of slots to the pool.
func (p *GraphemePool) grow() error {
	slotCountMax := uint32(gpSlotMask + 1)
	if p.numSlots > slotCountMax-p.slotsPerPage {
		return ErrGraphemeOutOfMemory
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

func (p *GraphemePool) slotPtr(slotIndex uint32) []byte {
	offset := int(slotIndex) * p.slotSizeBytes
	return p.slots[offset : offset+p.slotSizeBytes : offset+p.slotSizeBytes]
}

func gpSlotHeaderFromBytes(b []byte) *gpSlotHeader {
	return (*gpSlotHeader)(unsafe.Pointer(&b[0]))
}

func gpPackID(classID, slotIndex, generation uint32) (gpIdPayload, error) {
	if classID > gpClassMask {
		return 0, ErrGraphemeInvalidID
	}
	if slotIndex > gpSlotMask {
		return 0, ErrGraphemeOutOfMemory
	}
	if generation > gpGenMask {
		return 0, ErrGraphemeInvalidID
	}
	return gpIdPayload((classID << (gpGenBits + gpSlotBits)) | ((generation & gpGenMask) << gpSlotBits) | (slotIndex & gpSlotMask)), nil
}

func gpUnpackID(id gpIdPayload) (classID, slotIndex, generation uint32) {
	classID = (uint32(id) >> (gpGenBits + gpSlotBits)) & gpClassMask
	slotIndex = uint32(id) & gpSlotMask
	generation = (uint32(id) >> gpSlotBits) & gpGenMask
	return
}

// internLiveID records grapheme as the live identity for id (put-no-clobber).
func (p *GraphemePool) internLiveID(id gpIdPayload, grapheme string) {
	if _, exists := p.interned[grapheme]; !exists {
		p.interned[grapheme] = id
	}
}

// removeInternedLiveID removes the grapheme's live id entry if it matches id.
func (p *GraphemePool) removeInternedLiveID(grapheme string, id gpIdPayload) {
	if cur, ok := p.interned[grapheme]; ok && cur == id {
		delete(p.interned, grapheme)
	}
}

// lookupLive returns the live id for grapheme if it exists and is valid.
func (p *GraphemePool) lookupLive(grapheme string) (gpIdPayload, bool) {
	id, ok := p.interned[grapheme]
	if !ok {
		return 0, false
	}
	if _, err := p.getLocked(id); err != nil {
		return 0, false
	}
	return id, true
}

// gpClassForSize returns the size class index for a grapheme byte length.
func gpClassForSize(size int) uint32 {
	if size <= 8 {
		return 0
	}
	if size <= 16 {
		return 1
	}
	if size <= 32 {
		return 2
	}
	if size <= 64 {
		return 3
	}
	return 4 // up to 128
}

// Intern registers a grapheme and returns its unique ID (compatibility API).
func (p *GraphemePool) Intern(grapheme string) uint32 {
	id, err := p.Acquire(grapheme)
	if err != nil {
		return 0
	}
	return uint32(id)
}

// Lookup returns the grapheme for the given ID, or empty string if not found.
func (p *GraphemePool) Lookup(id uint32) string {
	g, err := p.Get(gpIdPayload(id))
	if err != nil {
		return ""
	}
	return g
}

// Size returns the number of graphemes in the pool.
func (p *GraphemePool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.interned)
}

// Acquire returns one owned live reference for grapheme, interning it if needed.
func (p *GraphemePool) Acquire(grapheme string) (gpIdPayload, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(grapheme) > gpMaxGraphemeLen {
		return 0, ErrGraphemeTooLong
	}
	if id, ok := p.lookupLive(grapheme); ok {
		if err := p.increfLocked(id); err != nil {
			return 0, err
		}
		return id, nil
	}
	return p.acquireNewLocked(grapheme)
}

func (p *GraphemePool) acquireNewLocked(grapheme string) (gpIdPayload, error) {
	if len(p.freeList) == 0 {
		if err := p.grow(); err != nil {
			return 0, err
		}
	}
	slotIndex := p.freeList[len(p.freeList)-1]
	p.freeList = p.freeList[:len(p.freeList)-1]
	b := p.slotPtr(slotIndex)
	h := gpSlotHeaderFromBytes(b)
	if h.generation >= gpGenMask {
		return 0, ErrGraphemeOutOfMemory
	}
	newGen := h.generation + 1
	h.len = uint16(len(grapheme))
	h.refcount = 0
	h.generation = newGen
	h.isAllocated = 1
	dataPtr := b[gpSlotHeaderSize:]
	copy(dataPtr[:len(grapheme)], grapheme)

	id, err := gpPackID(gpClassForSize(len(grapheme)), slotIndex, newGen)
	if err != nil {
		return 0, err
	}
	p.internLiveID(id, grapheme)
	h.refcount = 1
	return id, nil
}

// Incref retains an already-live id.
func (p *GraphemePool) Incref(id gpIdPayload) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.increfLocked(id)
}

func (p *GraphemePool) increfLocked(id gpIdPayload) error {
	classID, slotIndex, generation := gpUnpackID(id)
	if classID > gpClassMask || slotIndex >= p.numSlots {
		return ErrGraphemeInvalidID
	}
	b := p.slotPtr(slotIndex)
	h := gpSlotHeaderFromBytes(b)
	if h.generation != generation {
		return ErrGraphemeWrongGeneration
	}
	if h.isAllocated != 1 || h.refcount == 0 {
		return ErrGraphemeInvalidID
	}
	if h.refcount == ^uint32(0) {
		return ErrGraphemeRefcountOverflow
	}
	h.refcount++
	return nil
}

// Decref releases one reference to id.
func (p *GraphemePool) Decref(id gpIdPayload) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	classID, slotIndex, generation := gpUnpackID(id)
	if classID > gpClassMask || slotIndex >= p.numSlots {
		return ErrGraphemeInvalidID
	}
	b := p.slotPtr(slotIndex)
	h := gpSlotHeaderFromBytes(b)
	if h.refcount == 0 {
		return ErrGraphemeInvalidID
	}
	if h.generation != generation {
		return ErrGraphemeWrongGeneration
	}
	if h.refcount == 1 {
		grapheme, err := p.getLocked(id)
		if err == nil {
			p.removeInternedLiveID(grapheme, id)
		}
	}
	h.refcount--
	if h.refcount == 0 {
		h.isAllocated = 0
		if h.generation == gpGenMask {
			// Retired slot: not returned to free list.
		} else {
			p.freeList = append(p.freeList, slotIndex)
		}
	}
	return nil
}

// Get returns the grapheme for id.
func (p *GraphemePool) Get(id gpIdPayload) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getLocked(id)
}

func (p *GraphemePool) getLocked(id gpIdPayload) (string, error) {
	classID, slotIndex, generation := gpUnpackID(id)
	if classID > gpClassMask || slotIndex >= p.numSlots {
		return "", ErrGraphemeInvalidID
	}
	b := p.slotPtr(slotIndex)
	h := gpSlotHeaderFromBytes(b)
	if h.generation != generation {
		return "", ErrGraphemeWrongGeneration
	}
	if h.isAllocated != 1 {
		return "", ErrGraphemeInvalidID
	}
	dataPtr := b[gpSlotHeaderSize:]
	return string(dataPtr[:h.len]), nil
}

// GetRefcount returns the current refcount for id.
func (p *GraphemePool) GetRefcount(id gpIdPayload) (uint32, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	classID, slotIndex, generation := gpUnpackID(id)
	if classID > gpClassMask || slotIndex >= p.numSlots {
		return 0, ErrGraphemeInvalidID
	}
	b := p.slotPtr(slotIndex)
	h := gpSlotHeaderFromBytes(b)
	if h.generation != generation {
		return 0, ErrGraphemeWrongGeneration
	}
	return h.refcount, nil
}

// GetTotalSlots returns the total number of slots allocated.
func (p *GraphemePool) GetTotalSlots() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.numSlots
}

// GetFreeSlotCount returns the number of free slots.
func (p *GraphemePool) GetFreeSlotCount() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return uint32(len(p.freeList))
}

// GetLiveSlotCount returns the number of live (non-free) slots.
func (p *GraphemePool) GetLiveSlotCount() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.numSlots - uint32(len(p.freeList))
}

// Close releases all interned live ids and frees the slot store.
func (p *GraphemePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for g := range p.interned {
		delete(p.interned, g)
	}
	p.slots = nil
	p.freeList = nil
	p.numSlots = 0
	return nil
}

// GraphemeTracker tracks grapheme usage per buffer with per-cell refcounting.
// Mirrors packages/native/src/grapheme.zig GraphemeTracker.
type GraphemeTracker struct {
	mu      sync.Mutex
	pool    *GraphemePool
	usedIDs map[gpIdPayload]uint32 // id -> number of cells in this buffer
}

// NewGraphemeTracker creates a new GraphemeTracker backed by the given pool.
func NewGraphemeTracker(pool *GraphemePool) *GraphemeTracker {
	return &GraphemeTracker{
		pool:    pool,
		usedIDs: make(map[gpIdPayload]uint32),
	}
}

func (t *GraphemeTracker) decRefAll() {
	for id := range t.usedIDs {
		_ = t.pool.Decref(id)
	}
	t.usedIDs = make(map[gpIdPayload]uint32)
}

// Close releases all tracked ids.
func (t *GraphemeTracker) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.decRefAll()
	return nil
}

// Clear releases all tracked ids.
func (t *GraphemeTracker) Clear() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.decRefAll()
	return nil
}

// Add tracks a grapheme id once (first cell), incrementing the pool refcount.
func (t *GraphemeTracker) Add(id gpIdPayload) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	count, exists := t.usedIDs[id]
	if !exists {
		if err := t.pool.Incref(id); err != nil {
			return err
		}
		t.usedIDs[id] = 1
		return nil
	}
	count++
	t.usedIDs[id] = count
	return nil
}

// Remove removes a cell reference for id.
func (t *GraphemeTracker) Remove(id gpIdPayload) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	count, ok := t.usedIDs[id]
	if !ok || count == 0 {
		return nil
	}
	count--
	if count == 0 {
		delete(t.usedIDs, id)
		_ = t.pool.Decref(id)
		return nil
	}
	t.usedIDs[id] = count
	return nil
}

// Replace replaces old_id with new_id.
func (t *GraphemeTracker) Replace(oldID, newID gpIdPayload) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if oldID == newID {
		return nil
	}
	if newID != 0 {
		if err := t.Add(newID); err != nil {
			return err
		}
	}
	if oldID != 0 {
		if err := t.Remove(oldID); err != nil {
			return err
		}
	}
	return nil
}

// Contains returns true if id is tracked.
func (t *GraphemeTracker) Contains(id gpIdPayload) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.usedIDs[id]
	return ok
}

// HasAny returns true if any graphemes are tracked.
func (t *GraphemeTracker) HasAny() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.usedIDs) > 0
}

// GetGraphemeCount returns the number of distinct tracked graphemes.
func (t *GraphemeTracker) GetGraphemeCount() uint32 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return uint32(len(t.usedIDs))
}

// GetGraphemeCellCount returns the total cell count across all tracked graphemes.
func (t *GraphemeTracker) GetGraphemeCellCount() uint32 {
	t.mu.Lock()
	defer t.mu.Unlock()
	var total uint32
	for _, count := range t.usedIDs {
		total += count
	}
	return total
}

var _ = errors.New
