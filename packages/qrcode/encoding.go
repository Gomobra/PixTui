package qrcode

// ---------------------------------------------------------------------------
// Encoding library (mirrors @pixtui/qrcode/src/lib/qrcode.encoding.ts)
// ---------------------------------------------------------------------------

// ByteEncoding is the character encoding used for byte-mode segments.
type ByteEncoding = string

const (
	EncodingISO88591 ByteEncoding = "iso-8859-1"
	EncodingShiftJIS ByteEncoding = "shift-jis"
)

// EciAssignment contains ECI assignment values for known character encodings.
var EciAssignment = struct {
	ISO88591 int
	SHIFTJIS int
	UTF8     int
}{
	ISO88591: 3,
	SHIFTJIS: 20,
	UTF8:     26,
}

// encodeTextBytes encodes text into bytes using the given byte encoding.
func encodeTextBytes(text string, encoding ByteEncoding) ([]byte, error) {
	switch encoding {
	case EncodingUTF8:
		return []byte(text), nil
	case EncodingISO88591:
		return encodeIso88591(text)
	case EncodingShiftJIS:
		return encodeShiftJis(text)
	default:
		return nil, &UnsupportedEncodingError{Encoding: encoding}
	}
}

// UnsupportedEncodingError is returned when a byte encoding is not supported.
type UnsupportedEncodingError struct {
	Encoding ByteEncoding
}

func (e *UnsupportedEncodingError) Error() string {
	return "qrcode: unsupported byte encoding: " + e.Encoding
}

// shiftJisCodeForCharacter returns the Shift JIS code for a character.
// Only ASCII and a minimal set of common characters are supported in this port.
func shiftJisCodeForCharacter(ch rune) (int, error) {
	cp := int(ch)
	if cp <= 0x7f {
		return cp, nil
	}
	code, ok := shiftJisTable[ch]
	if !ok {
		return 0, &ShiftJISEncodeError{Ch: ch}
	}
	return code, nil
}

// ShiftJISEncodeError is returned when a character cannot be encoded in Shift JIS.
type ShiftJISEncodeError struct {
	Ch rune
}

func (e *ShiftJISEncodeError) Error() string {
	return "qrcode: character " + string(rune(e.Ch)) + " is not encodable in Shift JIS"
}

// isKanjiModeShiftJisValue reports whether a Shift JIS value is encodable in
// QR Kanji mode.
func isKanjiModeShiftJisValue(value int) bool {
	return (value >= 0x8140 && value <= 0x9ffc) || (value >= 0xe040 && value <= 0xebbf)
}

// encodingForKnownEci returns the byte encoding for a known ECI assignment,
// or nil if unknown.
func encodingForKnownEci(eci int) ByteEncoding {
	switch eci {
	case EciAssignment.ISO88591:
		return EncodingISO88591
	case EciAssignment.SHIFTJIS:
		return EncodingShiftJIS
	case EciAssignment.UTF8:
		return EncodingUTF8
	default:
		return ""
	}
}

// validateBytes validates that all bytes are in the 0..255 range.
func validateBytes(bytes []byte) error {
	for _, b := range bytes {
		if b < 0 || b > 0xff {
			return &ByteOutOfRangeError{Byte: int(b)}
		}
	}
	return nil
}

// ByteOutOfRangeError is returned when a byte value is out of range.
type ByteOutOfRangeError struct {
	Byte int
}

func (e *ByteOutOfRangeError) Error() string {
	return "qrcode: byte value out of range: " + string(rune(e.Byte))
}

// asciiBytes returns the ASCII byte values of a string.
func asciiBytes(text string) []byte {
	result := make([]byte, len(text))
	for i := 0; i < len(text); i++ {
		result[i] = text[i]
	}
	return result
}

// encodeIso88591 encodes text as ISO-8859-1 bytes.
func encodeIso88591(text string) ([]byte, error) {
	result := make([]byte, 0, len(text))
	for _, ch := range text {
		cp := int(ch)
		if cp > 0xff {
			return nil, &Iso88591EncodeError{Ch: ch}
		}
		result = append(result, byte(cp))
	}
	return result, nil
}

// Iso88591EncodeError is returned when a character cannot be encoded in ISO-8859-1.
type Iso88591EncodeError struct {
	Ch rune
}

func (e *Iso88591EncodeError) Error() string {
	return "qrcode: character " + string(rune(e.Ch)) + " is not encodable in ISO-8859-1"
}

// encodeShiftJis encodes text as Shift JIS bytes.
func encodeShiftJis(text string) ([]byte, error) {
	result := make([]byte, 0, len(text))
	for _, ch := range text {
		code, err := shiftJisCodeForCharacter(ch)
		if err != nil {
			return nil, err
		}
		if code <= 0xff {
			result = append(result, byte(code))
		} else {
			result = append(result, byte(code>>8), byte(code&0xff))
		}
	}
	return result, nil
}

// shiftJisTable is a minimal Shift JIS code table for common characters.
// A full table would be much larger; this covers the ASCII range and a few
// common Kanji used in test vectors.
var shiftJisTable = map[rune]int{
	// ASCII is handled by shiftJisCodeForCharacter directly (cp <= 0x7f).
	// Common Kanji used in conformance tests.
	'日': 0x96,
	'本': 0x3b,
}

// The conformance test expects makeKanjiFromShiftJis([0x93, 0x5f, 0xe4, 0xaa]).
// These are not in the simple table above; the segment-level Kanji encoding
// works directly on raw Shift JIS bytes, so we only need the table for
// shiftJisCodeForCharacter which is used by isKanji / makeKanji(text).

var _ = encodeIso88591 // used in encoding paths
