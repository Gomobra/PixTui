package qrcode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// segmentBitString returns the bit string of a segment's data buffer,
// limited to bitLen meaningful bits.
func segmentBitString(s *segment) string {
	var result strings.Builder
	byteCount := (s.bitLen + 7) / 8
	for i := 0; i < byteCount; i++ {
		b := s.data[i]
		startBit := 0
		if i == byteCount-1 {
			bitsInByte := s.bitLen % 8
			if bitsInByte == 0 {
				bitsInByte = 8
			}
			// Bits are packed MSB-first in each byte, so meaningful bits
			// occupy the top bitsInByte positions of the last byte.
			startBit = 8 - bitsInByte
		}
		for bit := 7; bit >= startBit; bit-- {
			if b&(1<<bit) != 0 {
				result.WriteByte('1')
			} else {
				result.WriteByte('0')
			}
		}
	}
	return result.String()
}

func TestNumericModeBitStream(t *testing.T) {
	seg := makeNumeric("01234567")
	if got := segmentBitString(seg); got != "000000110001010110011000011" {
		t.Errorf("numeric bit stream = %q, want %q", got, "000000110001010110011000011")
	}
}

func TestAlphanumericModeBitStream(t *testing.T) {
	seg := makeAlphanumeric("AC-42")
	if got := segmentBitString(seg); got != "0011100111011100111001000010" {
		t.Errorf("alphanumeric bit stream = %q, want %q", got, "0011100111011100111001000010")
	}
}

func TestKanjiModeBitStream(t *testing.T) {
	seg := makeKanjiFromShiftJis([]byte{0x93, 0x5f, 0xe4, 0xaa})
	if got := segmentBitString(seg); got != "01101100111111101010101010" {
		t.Errorf("kanji bit stream = %q, want %q", got, "01101100111111101010101010")
	}
}

func TestECIBitStreamBoundaries(t *testing.T) {
	seg9 := makeEci(9)
	if got := segmentBitString(seg9); got != "00001001" {
		t.Errorf("ECI 9 bit stream = %q, want %q", got, "00001001")
	}
	seg128 := makeEci(128)
	if got := segmentBitString(seg128); got != "1000000010000000" {
		t.Errorf("ECI 128 bit stream = %q, want %q", got, "1000000010000000")
	}
	seg999999 := makeEci(999999)
	if got := segmentBitString(seg999999); got != "110011110100001000111111" {
		t.Errorf("ECI 999999 bit stream = %q, want %q", got, "110011110100001000111111")
	}
}

func TestVersion1MNumericCodewords(t *testing.T) {
	qr, err := EncodeText("01234567", LevelM, EncodeOptions{
		BoostECL:   false,
		Optimize:   false,
		MinVersion: 1,
		MaxVersion: 1,
		Mask:       2,
	})
	if err != nil {
		t.Fatalf("EncodeText: %v", err)
	}

	codewords := readPlacedCodewords(qr, qr.Mask)
	expected := []byte{0x10, 0x20, 0x0c, 0x56, 0x61, 0x80, 0xec, 0x11, 0xec, 0x11, 0xec, 0x11, 0xec, 0x11, 0xec, 0x11, 0xa5, 0x24, 0xd4, 0xc1, 0xed, 0x36, 0xc7, 0x87, 0x2c, 0x55}
	if len(codewords) != len(expected) {
		t.Fatalf("codewords length = %d, want %d", len(codewords), len(expected))
	}
	for i := range expected {
		if codewords[i] != expected[i] {
			t.Errorf("codeword[%d] = 0x%02x, want 0x%02x", i, codewords[i], expected[i])
		}
	}
}

func TestFormatInformationSequences(t *testing.T) {
	expectedByDataBits := []int{
		0x5412, 0x5125, 0x5e7c, 0x5b4b, 0x45f9, 0x40ce, 0x4f97, 0x4aa0,
		0x77c4, 0x72f3, 0x7daa, 0x789d, 0x662f, 0x6318, 0x6c41, 0x6976,
		0x1689, 0x13be, 0x1ce7, 0x19d0, 0x0762, 0x0255, 0x0d0c, 0x083b,
		0x355f, 0x3068, 0x3f31, 0x3a06, 0x24b4, 0x2183, 0x2eda, 0x2bed,
	}
	eclByDataBits := []ErrorCorrectionLevel{
		LevelM, LevelL, LevelH, LevelQ,
	}

	for dataBits := 0; dataBits < len(expectedByDataBits); dataBits++ {
		ecl := eclByDataBits[dataBits>>3]
		mask := dataBits & 0b111
		qr, err := EncodeText("A", ecl, EncodeOptions{
			BoostECL:   false,
			Optimize:   false,
			MinVersion: 1,
			MaxVersion: 1,
			Mask:       mask,
		})
		if err != nil {
			t.Fatalf("EncodeText (ecl=%v, mask=%d): %v", ecl, mask, err)
		}
		if got := readPrimaryFormatBits(qr); got != expectedByDataBits[dataBits] {
			t.Errorf("primary format bits (ecl=%v, mask=%d) = 0x%04x, want 0x%04x", ecl, mask, got, expectedByDataBits[dataBits])
		}
		if got := readSecondaryFormatBits(qr); got != expectedByDataBits[dataBits] {
			t.Errorf("secondary format bits (ecl=%v, mask=%d) = 0x%04x, want 0x%04x", ecl, mask, got, expectedByDataBits[dataBits])
		}
	}
}

func TestVersionInformationSequences(t *testing.T) {
	tests := []struct {
		version   int
		expected  int
	}{
		{7, 0x07c94},
		{8, 0x085bc},
		{14, 0x0e60d},
		{21, 0x15683},
		{32, 0x209d5},
		{40, 0x28c69},
	}

	for _, tc := range tests {
		qr, err := EncodeText("A", LevelH, EncodeOptions{
			BoostECL:   false,
			Optimize:   false,
			MinVersion: tc.version,
			MaxVersion: tc.version,
			Mask:       0,
		})
		if err != nil {
			t.Fatalf("EncodeText (version=%d): %v", tc.version, err)
		}
		if got := readTopRightVersionBits(qr); got != tc.expected {
			t.Errorf("top-right version bits (version=%d) = 0x%05x, want 0x%05x", tc.version, got, tc.expected)
		}
		if got := readBottomLeftVersionBits(qr); got != tc.expected {
			t.Errorf("bottom-left version bits (version=%d) = 0x%05x, want 0x%05x", tc.version, got, tc.expected)
		}
	}
}

func TestAlignmentPatternCoordinates(t *testing.T) {
	tests := []struct {
		version   int
		positions []int
	}{
		{2, []int{6, 18}},
		{7, []int{6, 22, 38}},
		{14, []int{6, 26, 46, 66}},
		{21, []int{6, 28, 50, 72, 94}},
		{32, []int{6, 34, 60, 86, 112, 138}},
		{40, []int{6, 30, 58, 86, 114, 142, 170}},
	}

	for _, tc := range tests {
		qr, err := EncodeText("A", LevelH, EncodeOptions{
			BoostECL:   false,
			Optimize:   false,
			MinVersion: tc.version,
			MaxVersion: tc.version,
			Mask:       0,
		})
		if err != nil {
			t.Fatalf("EncodeText (version=%d): %v", tc.version, err)
		}
		modules := qr.ToMatrix()
		for _, y := range tc.positions {
			for _, x := range tc.positions {
				if overlapsFinderPattern(x, y, qr.Size) {
					continue
				}
				if !expectAlignmentPattern(modules, x, y) {
					t.Errorf("alignment pattern missing at (%d,%d) for version %d", x, y, tc.version)
				}
			}
		}
	}
}

func TestMaskConditions(t *testing.T) {
	coordinates := [][2]int{
		{0, 0}, {1, 0}, {3, 2}, {5, 7}, {8, 9},
	}
	for mask := 0; mask < 8; mask++ {
		for _, c := range coordinates {
			x, y := c[0], c[1]
			if got := maskCondition(mask, x, y); got != expectedMaskCondition(mask, x, y) {
				t.Errorf("maskCondition(%d,%d,%d) = %v, want %v", mask, x, y, got, expectedMaskCondition(mask, x, y))
			}
		}
	}
}

func TestDataCapacityBoundaries(t *testing.T) {
	// Version 1-H holds exactly 17 numeric digits.
	if _, err := EncodeNumericExact(1, LevelH, 17); err != nil {
		t.Errorf("expected 17 digits to fit in v1-H, got %v", err)
	}
	if _, err := EncodeNumericExact(1, LevelH, 18); err == nil {
		t.Error("expected 18 digits to NOT fit in v1-H")
	}
	// Version 10-Q holds exactly 364 numeric digits.
	if _, err := EncodeNumericExact(10, LevelQ, 364); err != nil {
		t.Errorf("expected 364 digits to fit in v10-Q, got %v", err)
	}
	if _, err := EncodeNumericExact(10, LevelQ, 365); err == nil {
		t.Error("expected 365 digits to NOT fit in v10-Q")
	}
	// Version 40-L holds exactly 7089 numeric digits.
	if _, err := EncodeNumericExact(40, LevelL, 7089); err != nil {
		t.Errorf("expected 7089 digits to fit in v40-L, got %v", err)
	}
	if _, err := EncodeNumericExact(40, LevelL, 7090); err == nil {
		t.Error("expected 7090 digits to NOT fit in v40-L")
	}
}

func TestStructuredAppendParity(t *testing.T) {
	segs := []*segment{makeNumeric("0123456789"), makeKanjiFromShiftJis([]byte{0x93, 0x5f, 0xe4, 0xaa})}
	// Compute parity: XOR of all segment raw bytes (mirrors TypeScript parityBytes).
	var parity byte
	for _, seg := range segs {
		for _, b := range seg.rawBytes {
			parity ^= b
		}
	}
	if parity != 0x83 {
		t.Errorf("structured append parity = 0x%02x, want 0x%02x", parity, 0x83)
	}
}

func TestECIUTF8Option(t *testing.T) {
	// Use text that forces byte mode (non-alphanumeric) so an ECI prefix is emitted.
	text := "ABC©"

	// Default: ECIForUTF8 zero value means "use default" = true (ECI prefix present).
	qrDefault, err := EncodeText(text, LevelM, EncodeOptions{
		BoostECL: false,
		Optimize: false,
	})
	if err != nil {
		t.Fatalf("EncodeText: %v", err)
	}

	// Explicitly disable ECI.
	falseBool := false
	qrNoECI, err := EncodeText(text, LevelM, EncodeOptions{
		BoostECL:     false,
		Optimize:     false,
		ECIForUTF8:   &falseBool,
		ECIAssignment: nil,
	})
	if err != nil {
		t.Fatalf("EncodeText: %v", err)
	}

	// With ECI disabled, the matrix should differ from the default (which has an ECI prefix).
	if matricesEqual(qrDefault, qrNoECI) {
		t.Error("expected matrix to differ when ECI is disabled")
	}
}

// ---------------------------------------------------------------------------
// Additional segment and API conformance tests
// ---------------------------------------------------------------------------

func TestMakeNumericSegment(t *testing.T) {
	seg := MakeNumeric("01234567")
	if got := segmentBitString(seg); got != "000000110001010110011000011" {
		t.Errorf("numeric bit stream = %q, want %q", got, "000000110001010110011000011")
	}
	if seg.GetTotalBits(1) != 41 {
		t.Errorf("numeric total bits = %d, want 41", seg.GetTotalBits(1))
	}
}

func TestMakeAlphanumericSegment(t *testing.T) {
	seg := MakeAlphanumeric("AC-42")
	if got := segmentBitString(seg); got != "0011100111011100111001000010" {
		t.Errorf("alphanumeric bit stream = %q, want %q", got, "0011100111011100111001000010")
	}
	if seg.GetTotalBits(1) != 41 {
		t.Errorf("alphanumeric total bits = %d, want 41", seg.GetTotalBits(1))
	}
}

func TestMakeEciSegment(t *testing.T) {
	if got := segmentBitString(MakeEci(9)); got != "00001001" {
		t.Errorf("ECI 9 bit stream = %q, want %q", got, "00001001")
	}
	if got := segmentBitString(MakeEci(128)); got != "1000000010000000" {
		t.Errorf("ECI 128 bit stream = %q, want %q", got, "1000000010000000")
	}
	if got := segmentBitString(MakeEci(999999)); got != "110011110100001000111111" {
		t.Errorf("ECI 999999 bit stream = %q, want %q", got, "110011110100001000111111")
	}
}

func TestEncodeBytes(t *testing.T) {
	qr, err := EncodeBytes([]byte{0x41, 0x42, 0x43}, LevelM, EncodeOptions{
		BoostECL:   false,
		Optimize:   false,
		MinVersion: 1,
		MaxVersion: 1,
		Mask:       2,
	})
	if err != nil {
		t.Fatalf("EncodeBytes: %v", err)
	}
	if qr.Version != 1 {
		t.Errorf("expected version 1, got %d", qr.Version)
	}
	if qr.Level != LevelM {
		t.Errorf("expected level M, got %v", qr.Level)
	}
}

func TestEncodeSegments(t *testing.T) {
	segs := []*segment{
		MakeNumeric("12345"),
		MakeAlphanumeric("ABC"),
	}
	qr, err := EncodeSegments(segs, LevelM, EncodeOptions{
		BoostECL:   false,
		Optimize:   false,
		MinVersion: 1,
		MaxVersion: 1,
		Mask:       0,
	})
	if err != nil {
		t.Fatalf("EncodeSegments: %v", err)
	}
	if qr.Version != 1 {
		t.Errorf("expected version 1, got %d", qr.Version)
	}
}

func TestQRCodeFields(t *testing.T) {
	qr, err := EncodeText("ABC", LevelM, EncodeOptions{
		BoostECL:   false,
		Optimize:   false,
		MinVersion: 1,
		MaxVersion: 1,
		Mask:       0,
	})
	if err != nil {
		t.Fatalf("EncodeText: %v", err)
	}
	if qr.SymbologyIdentifier != "]Q1" {
		t.Errorf("expected symbology ]Q1, got %s", qr.SymbologyIdentifier)
	}
	if qr.ContainsEci {
		t.Error("expected ContainsEci to be false for plain text")
	}
	if qr.Fnc1 != "none" {
		t.Errorf("expected Fnc1 none, got %s", qr.Fnc1)
	}
	if qr.Version != 1 {
		t.Errorf("expected version 1, got %d", qr.Version)
	}
	if qr.Size != 21 {
		t.Errorf("expected size 21, got %d", qr.Size)
	}
}

func TestQRCodeToMatrix(t *testing.T) {
	qr, err := EncodeText("A", LevelL, EncodeOptions{
		BoostECL:   false,
		Optimize:   false,
		MinVersion: 1,
		MaxVersion: 1,
		Mask:       0,
	})
	if err != nil {
		t.Fatalf("EncodeText: %v", err)
	}
	matrix := qr.ToMatrix()
	if len(matrix) != qr.Size {
		t.Errorf("expected matrix size %d, got %d", qr.Size, len(matrix))
	}
	if len(matrix[0]) != qr.Size {
		t.Errorf("expected matrix row size %d, got %d", qr.Size, len(matrix[0]))
	}
	// ToMatrix returns a deep copy.
	matrix[0][0] = !matrix[0][0]
	if matrix[0][0] == qr.modules[0][0] {
		t.Error("expected ToMatrix to return a deep copy")
	}
}

func TestSavePNG(t *testing.T) {
	qr, err := EncodeText("A", LevelL, EncodeOptions{
		BoostECL:   false,
		Optimize:   false,
		MinVersion: 1,
		MaxVersion: 1,
		Mask:       0,
	})
	if err != nil {
		t.Fatalf("EncodeText: %v", err)
	}
	path := filepath.Join(t.TempDir(), "qr.png")
	if err := qr.SavePNG(path); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved PNG: %v", err)
	}
	if len(data) < 8 {
		t.Error("expected non-empty PNG file")
	}
	// PNG signature.
	if data[0] != 0x89 || data[1] != 0x50 || data[2] != 0x4e || data[3] != 0x47 {
		t.Error("expected valid PNG signature")
	}
}

// matricesEqual reports whether two QR codes have identical module matrices.
func matricesEqual(a, b *QRCode) bool {
	if a.Size != b.Size {
		return false
	}
	ma := a.ToMatrix()
	mb := b.ToMatrix()
	for y := 0; y < a.Size; y++ {
		for x := 0; x < a.Size; x++ {
			if ma[y][x] != mb[y][x] {
				return false
			}
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Helpers for conformance tests
// ---------------------------------------------------------------------------

// EncodeNumericExact encodes a string of `length` numeric digits at an exact
// version with boost disabled.
func EncodeNumericExact(version int, ecl ErrorCorrectionLevel, length int) (*QRCode, error) {
	return EncodeText(strings.Repeat("1", length), ecl, EncodeOptions{
		BoostECL:   false,
		Optimize:   false,
		MinVersion: version,
		MaxVersion: version,
		Mask:       0,
	})
}

// readPlacedCodewords reads the data codewords from a QR code matrix by
// reversing the mask and reading in the standard QR codeword order.
func readPlacedCodewords(qr *QRCode, mask int) []byte {
	modules := qr.modules
	size := qr.Size
	functionModules := makeFunctionModuleMap(size)

	var dataBits []bool
	for right := size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5
		}
		for vert := 0; vert < size; vert++ {
			for j := 0; j < 2; j++ {
				x := right - j
				upward := ((right+1)&2) == 0
				y := vert
				if !upward {
					y = size - 1 - vert
				}
				if functionModules[y][x] {
					continue
				}
				dark := modules[y][x] != maskCondition(mask, x, y)
				dataBits = append(dataBits, dark)
			}
		}
	}

	var codewords []byte
	for i := 0; i+8 <= len(dataBits); i += 8 {
		var v byte
		for j := 0; j < 8; j++ {
			if dataBits[i+j] {
				v |= 1 << (7 - j)
			}
		}
		codewords = append(codewords, v)
	}
	return codewords
}

// makeFunctionModuleMap builds a map of function modules for a given size.
func makeFunctionModuleMap(size int) [][]bool {
	m := make([][]bool, size)
	for i := 0; i < size; i++ {
		m[i] = make([]bool, size)
	}
	mark := func(x, y int) {
		if x >= 0 && y >= 0 && x < size && y < size {
			m[y][x] = true
		}
	}
	markFinder := func(cx, cy int) {
		for dy := -4; dy <= 4; dy++ {
			for dx := -4; dx <= 4; dx++ {
				mark(cx+dx, cy+dy)
			}
		}
	}
	markAlignment := func(cx, cy int) {
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				mark(cx+dx, cy+dy)
			}
		}
	}

	for i := 0; i < size; i++ {
		mark(i, 6)
		mark(6, i)
	}
	markFinder(3, 3)
	markFinder(size-4, 3)
	markFinder(3, size-4)

	// Version 1 has no alignment patterns.
	if size > 21 {
		numAlign := size/7 + 2
		var step int
		if size == 145 { // version 32
			step = 26
		} else {
			step = (((size+4)/(numAlign*2-2)) + 1) * 2
		}
		pos := size - 7
		positions := []int{6}
		for len(positions) < numAlign {
			positions = append(append([]int{6}, pos), positions[1:]...)
			pos -= step
		}
		for _, y := range positions {
			for _, x := range positions {
				if (x == 6 && y == 6) || (x == 6 && y == size-7) || (x == size-7 && y == 6) {
					continue
				}
				markAlignment(x, y)
			}
		}
	}

	for i := 0; i <= 5; i++ {
		mark(8, i)
	}
	mark(8, 7)
	mark(8, 8)
	mark(7, 8)
	for i := 9; i < 15; i++ {
		mark(14-i, 8)
	}
	for i := 0; i < 8; i++ {
		mark(size-1-i, 8)
	}
	for i := 8; i < 15; i++ {
		mark(8, size-15+i)
	}
	mark(8, size-8)

	if size >= 45 { // version >= 7
		for i := 0; i < 18; i++ {
			a := size - 11 + (i % 3)
			b := i / 3
			mark(a, b)
			mark(b, a)
		}
	}
	return m
}

func overlapsFinderPattern(x, y, size int) bool {
	return (x == 6 && y == 6) || (x == 6 && y == size-7) || (x == size-7 && y == 6)
}

func expectedMaskCondition(mask, x, y int) bool {
	switch mask {
	case 0:
		return (y+x)%2 == 0
	case 1:
		return y%2 == 0
	case 2:
		return x%3 == 0
	case 3:
		return (y+x)%3 == 0
	case 4:
		return ((y/2)+(x/3))%2 == 0
	case 5:
		return ((y*x)%2+(y*x)%3) == 0
	case 6:
		return (((y*x)%2+(y*x)%3))%2 == 0
	case 7:
		return (((y+x)%2+(y*x)%3))%2 == 0
	default:
		panic("invalid mask")
	}
}

// readPrimaryFormatBits reads the 15 primary format bits from the matrix.
func readPrimaryFormatBits(qr *QRCode) int {
	coords := [15][2]int{
		{8, 0}, {8, 1}, {8, 2}, {8, 3}, {8, 4}, {8, 5},
		{8, 7}, {8, 8}, {7, 8}, {5, 8}, {4, 8}, {3, 8}, {2, 8}, {1, 8}, {0, 8},
	}
	return readBitsAtCoordinates(qr.modules, coords[:])
}

// readSecondaryFormatBits reads the 15 secondary format bits from the matrix.
func readSecondaryFormatBits(qr *QRCode) int {
	size := qr.Size
	var coords [15][2]int
	for i := 0; i < 8; i++ {
		coords[i] = [2]int{size - 1 - i, 8}
	}
	for i := 8; i < 15; i++ {
		coords[i] = [2]int{8, size - 15 + i}
	}
	return readBitsAtCoordinates(qr.modules, coords[:])
}

// readTopRightVersionBits reads the 18 version bits from the top-right area.
func readTopRightVersionBits(qr *QRCode) int {
	size := qr.Size
	var coords [18][2]int
	for i := 0; i < 18; i++ {
		coords[i] = [2]int{size - 11 + (i % 3), i / 3}
	}
	return readBitsAtCoordinates(qr.modules, coords[:])
}

// readBottomLeftVersionBits reads the 18 version bits from the bottom-left area.
func readBottomLeftVersionBits(qr *QRCode) int {
	size := qr.Size
	var coords [18][2]int
	for i := 0; i < 18; i++ {
		coords[i] = [2]int{i / 3, size - 11 + (i % 3)}
	}
	return readBitsAtCoordinates(qr.modules, coords[:])
}

func readBitsAtCoordinates(modules [][]bool, coords [][2]int) int {
	bits := 0
	for i, c := range coords {
		x, y := c[0], c[1]
		if modules[y][x] {
			bits |= 1 << i
		}
	}
	return bits
}

func expectAlignmentPattern(modules [][]bool, cx, cy int) bool {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			dist := max(abs(dx), abs(dy))
			expected := dist != 1
			if modules[cy+dy][cx+dx] != expected {
				return false
			}
		}
	}
	return true
}
