package native

import (
	"encoding/base64"
	"sync/atomic"
)

// NonNegative returns max(value, 0) for a signed integer value.
// Mirrors utils.zig nonNegative.
func NonNegative[T ~int | ~int8 | ~int16 | ~int32 | ~int64](value T) T {
	if value < 0 {
		return 0
	}
	return value
}

// Eql reports whether two slices of integers are equal.
// It is a straightforward byte comparison; Go's runtime handles
// efficient memory comparison natively.
// Mirrors utils.zig eql.
func Eql[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Base64Encode encodes source into dest using standard base64 encoding
// and returns the encoded slice. dest must be large enough to hold
// the encoded output (base64.StdEncoding.EncodedLen(len(source))).
// Mirrors utils.zig base64Encode.
func Base64Encode(dest []byte, source []byte) []byte {
	encoded := base64.StdEncoding.EncodeToString(source)
	n := copy(dest, encoded)
	return dest[:n]
}

// PaintEpoch is a process-wide monotonic paint epoch counter.
// Mirrors utils.zig nextPaintEpoch.
var paintEpoch atomic.Uint64

// NextPaintEpoch returns the next paint epoch, monotonically increasing.
// Mirrors utils.zig nextPaintEpoch.
func NextPaintEpoch() uint64 {
	return paintEpoch.Add(1) + 1
}

// FillU32 fills destination with value.
// Mirrors utils.zig fillU32.
func FillU32(destination []uint32, value uint32) {
	for i := range destination {
		destination[i] = value
	}
}
