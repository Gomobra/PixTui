package native

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"sync"
)

// Image represents a terminal-renderable image.
type Image struct {
	mu         sync.RWMutex
	ID         uint64
	Width      int
	Height     int
	Pixels     []uint8 // RGBA
	Format     string
	Owned      bool
	EncodedPNG []uint8 // optional cached PNG encoding, sent directly on kitty transmit
	HasAlpha   bool    // whether the pixels carry meaningful alpha
}

// NewImage creates a new blank RGBA image with the given dimensions.
func NewImage(width, height int) *Image {
	return &Image{
		Width:    width,
		Height:   height,
		Pixels:   make([]uint8, width*height*4),
		Format:   "rgba",
		Owned:    true,
		HasAlpha: true,
	}
}

// NewImageFromRGBA creates an Image from an image.RGBA.
func NewImageFromRGBA(rgba *image.RGBA) *Image {
	return &Image{
		Width:    rgba.Rect.Dx(),
		Height:   rgba.Rect.Dy(),
		Pixels:   append([]byte(nil), rgba.Pix...),
		Format:   "rgba",
		Owned:    true,
		HasAlpha: true,
	}
}

// LoadPNG loads a PNG image from file.
func LoadPNG(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DecodeImage(f, "png")
}

// LoadJPEG loads a JPEG image from file.
func LoadJPEG(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DecodeImage(f, "jpg")
}

// LoadGIF loads a GIF image from file (first frame).
func LoadGIF(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DecodeImage(f, "gif")
}

// LoadWebP loads a WebP image from file.
func LoadWebP(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DecodeImage(f, "webp")
}

// LoadImageFromBytes loads an image from raw bytes, auto-detecting format.
func LoadImageFromBytes(data []byte) (*Image, error) {
	format := detectFormat(data)
	return DecodeImage(bytes.NewReader(data), format)
}

// DecodeImage decodes an image from a reader given a format hint.
func DecodeImage(r io.Reader, format string) (*Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}
	if format == "" || format == "auto" {
		format = detectFormat(data)
	}
	switch format {
	case "png":
		return decodePNG(data)
	case "jpg", "jpeg":
		return decodeJPEG(data)
	case "gif":
		return decodeGIF(data)
	case "webp":
		return decodeWebP(data)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
}

// detectFormat inspects the first bytes of data to determine the image format.
func detectFormat(data []byte) string {
	if len(data) < 12 {
		return "unknown"
	}
	switch {
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "png"
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "jpg"
	case bytes.HasPrefix(data, []byte{'G', 'I', 'F', '8'}):
		return "gif"
	case len(data) >= 12 && bytes.HasPrefix(data, []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(data[8:12], []byte{'W', 'E', 'B', 'P'}):
		return "webp"
	default:
		return "unknown"
	}
}

func decodePNG(data []byte) (*Image, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("png decode failed: %w", err)
	}
	return imageToRGBA(img, "png"), nil
}

func decodeJPEG(data []byte) (*Image, error) {
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("jpeg decode failed: %w", err)
	}
	return imageToRGBA(img, "jpg"), nil
}

func decodeGIF(data []byte) (*Image, error) {
	img, err := gif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gif decode failed: %w", err)
	}
	return imageToRGBA(img, "gif"), nil
}

// decodeWebP decodes a WebP image using the custom VP8L lossless decoder.
// The VP8L payload starts right after the RIFF header (12 bytes).
func decodeWebP(data []byte) (*Image, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("webp data too short")
	}
	if !bytes.HasPrefix(data, []byte{'R', 'I', 'F', 'F'}) ||
		!bytes.Equal(data[8:12], []byte{'W', 'E', 'B', 'P'}) {
		return nil, fmt.Errorf("not a valid WebP file")
	}
	// Decode using the custom VP8L lossless decoder.
	return decodeVP8L(data[12:])
}

// decodeVP8L decodes a WebP lossless (VP8L) bitstream.
// This implements a subset sufficient for common lossless WebP images:
// 8-bit RGB/RGBA, no color cache (or simple color cache), no palette.
func decodeVP8L(data []byte) (*Image, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("vp8l: data too short")
	}

	// VP8L header: version(3) + colorspace(1) + alpha_is_used(1) +
	//                width(14) + height(14) = 32 bits total
	br := newBitReader(data)
	header := br.readBits(32)
	if br.err != nil {
		return nil, fmt.Errorf("vp8l: failed to read header: %w", br.err)
	}
	version := header & 0x7
	if version != 0 {
		return nil, fmt.Errorf("vp8l: unsupported version %d", version)
	}
	colorspace := (header >> 3) & 0x1
	alphaUsed := (header >> 4) & 0x1
	width := int((header>>5)&0x3FFF) + 1
	height := int((header>>19)&0x3FFF) + 1

	if colorspace != 0 {
		return nil, fmt.Errorf("vp8l: only sRGB colorspace supported, got %d", colorspace)
	}

	// ICC profile present flag (1 bit) after the 32-bit header
	if br.readBits(1) != 0 {
		return nil, fmt.Errorf("vp8l: ICC profile not supported")
	}

	// Color transform: 2 bits transform type
	transformType := br.readBits(2)
	if br.err != nil {
		return nil, fmt.Errorf("vp8l: failed to read transform: %w", br.err)
	}
	if transformType != 0 {
		return nil, fmt.Errorf("vp8l: color transform %d not supported", transformType)
	}

	br.alignToByte()

	return decodeVP8LImage(br, width, height, alphaUsed == 1)
}

// bitReader reads bits from a byte slice, LSB first within each byte.
type bitReader struct {
	data  []byte
	pos   int // bit position
	total int // total bits
	err   error
}

func newBitReader(data []byte) *bitReader {
	return &bitReader{data: data, total: len(data) * 8}
}

func (r *bitReader) readBits(n int) uint32 {
	if r.err != nil {
		return 0
	}
	var result uint32
	for i := 0; i < n; i++ {
		bitPos := r.pos
		if bitPos >= r.total {
			r.err = fmt.Errorf("bit reader overrun at bit %d", bitPos)
			return 0
		}
		byteIdx := bitPos / 8
		bitIdx := bitPos % 8
		bit := uint32((r.data[byteIdx] >> bitIdx) & 1)
		result |= bit << uint32(i)
		r.pos++
	}
	return result
}

func (r *bitReader) alignToByte() {
	r.pos = (r.pos + 7) & ^7
}

// decodeVP8LImage decodes the full VP8L image stream using Huffman + LZ77.
func decodeVP8LImage(br *bitReader, width, height int, alpha bool) (*Image, error) {
	pixels := make([]uint8, width*height*4)

	compressed := br.readBits(1)
	if br.err != nil {
		return nil, fmt.Errorf("vp8l: failed to read huffman flag: %w", br.err)
	}

	if compressed == 1 {
		_ = br.readBits(8)
	}

	clCodeLengths := make([]int, 19)
	clOrder := []int{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15}
	for i := 0; i < 19; i++ {
		clCodeLengths[clOrder[i]] = int(br.readBits(3))
	}
	if br.err != nil {
		return nil, fmt.Errorf("vp8l: failed to read code length codes: %w", br.err)
	}
	clHuff := buildHuffman(clCodeLengths)

	totalSymbols := 280 + 296
	symbolLengths := make([]int, totalSymbols)
	idx := 0
	for idx < totalSymbols {
		sym := clHuff.decode(br)
		if br.err != nil {
			return nil, fmt.Errorf("vp8l: failed to decode symbol length: %w", br.err)
		}
		if sym < 16 {
			symbolLengths[idx] = sym
			idx++
		} else if sym == 16 {
			repeat := int(br.readBits(2)) + 3
			if idx == 0 {
				return nil, fmt.Errorf("vp8l: repeat with no previous")
			}
			prev := symbolLengths[idx-1]
			for i := 0; i < repeat && idx < totalSymbols; i++ {
				symbolLengths[idx] = prev
				idx++
			}
		} else if sym == 17 {
			repeat := int(br.readBits(3)) + 3
			for i := 0; i < repeat && idx < totalSymbols; i++ {
				symbolLengths[idx] = 0
				idx++
			}
		} else if sym == 18 {
			repeat := int(br.readBits(7)) + 11
			for i := 0; i < repeat && idx < totalSymbols; i++ {
				symbolLengths[idx] = 0
				idx++
			}
		}
	}
	if br.err != nil {
		return nil, fmt.Errorf("vp8l: failed to read symbol lengths: %w", br.err)
	}

	litLenLengths := symbolLengths[:280]
	distLengths := symbolLengths[280:]

	litLenHuff := buildHuffman(litLenLengths)
	distHuff := buildHuffman(distLengths)

	output := make([]uint32, width*height)
	outIdx := 0

	for outIdx < len(output) {
		sym := litLenHuff.decode(br)
		if br.err != nil {
			return nil, fmt.Errorf("vp8l: LZ77 literal/length decode failed: %w", br.err)
		}
		if sym < 256 {
			if outIdx >= len(output) {
				break
			}
			output[outIdx] = uint32(sym)
			outIdx++
		} else if sym < 280 {
			lengthBase := lengthBase[sym-256]
			lengthExtra := lengthExtraBits[sym-256]
			length := lengthBase + int(br.readBits(lengthExtra))
			distSym := distHuff.decode(br)
			if br.err != nil {
				return nil, fmt.Errorf("vp8l: distance decode failed: %w", br.err)
			}
			distBase := distBase[distSym]
			distExtra := distExtraBits[distSym]
			dist := distBase + int(br.readBits(distExtra))
			if dist == 0 || dist > outIdx {
				return nil, fmt.Errorf("vp8l: invalid back-reference distance %d", dist)
			}
			for i := 0; i < length; i++ {
				if outIdx >= len(output) {
					break
				}
				output[outIdx] = output[outIdx-dist]
				outIdx++
			}
		} else {
			return nil, fmt.Errorf("vp8l: invalid symbol %d", sym)
		}
	}

	hasAlpha := alpha
	for i := 0; i < len(output); i++ {
		v := output[i]
		r := uint8((v >> 16) & 0xFF)
		g := uint8((v >> 8) & 0xFF)
		b := uint8(v & 0xFF)
		a := uint8((v >> 24) & 0xFF)
		if !hasAlpha {
			a = 0xFF
		}
		idx := i * 4
		pixels[idx] = r
		pixels[idx+1] = g
		pixels[idx+2] = b
		pixels[idx+3] = a
	}

	return &Image{
		Width:  width,
		Height: height,
		Pixels: pixels,
		Format: "webp",
		Owned:  true,
	}, nil
}

var lengthBase = []int{
	3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31,
	35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258,
}
var lengthExtraBits = []int{
	0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2,
	3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0,
}

var distBase = []int{
	1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193,
	257, 385, 513, 769, 1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577,
}
var distExtraBits = []int{
	0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6,
	7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13,
}

type huffmanNode struct {
	left  *huffmanNode
	right *huffmanNode
	value int
}

func buildHuffman(lengths []int) *huffmanDecoder {
	maxBits := 0
	for _, l := range lengths {
		if l > maxBits {
			maxBits = l
		}
	}
	blCount := make([]int, maxBits+1)
	for _, l := range lengths {
		if l > 0 {
			blCount[l]++
		}
	}

	nextCode := make([]int, maxBits+1)
	code := 0
	for bits := 1; bits <= maxBits; bits++ {
		code = (code + blCount[bits-1]) << 1
		nextCode[bits] = code
	}

	type symCode struct {
		sym  int
		code uint32
		len  int
	}
	var assigned []struct {
		sym  int
		code uint32
		len  int
	}
	for sym, l := range lengths {
		if l != 0 {
			assigned = append(assigned, struct {
				sym  int
				code uint32
				len  int
			}{sym: sym, code: uint32(nextCode[l]), len: l})
			nextCode[l]++
		}
	}

	return &huffmanDecoder{
		maxBits: maxBits,
		codes:   assigned,
	}
}

type huffmanDecoder struct {
	maxBits int
	codes   []struct {
		sym  int
		code uint32
		len  int
	}
	minCode [16]int
	maxCode [16]int
	valPtr  [16]int
}

func (d *huffmanDecoder) decode(br *bitReader) int {
	if len(d.codes) == 0 {
		return 0
	}
	code := uint32(0)
	len := 0
	for len < d.maxBits {
		code = (code << 1) | br.readBits(1)
		len++
		if br.err != nil {
			return 0
		}
		for _, c := range d.codes {
			if c.len == len && c.code == code {
				return c.sym
			}
		}
	}
	return 0
}
func imageToRGBA(img image.Image, format string) *Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixels := make([]uint8, width*height*4)
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[idx] = uint8(r >> 8)
			pixels[idx+1] = uint8(g >> 8)
			pixels[idx+2] = uint8(b >> 8)
			pixels[idx+3] = uint8(a >> 8)
			idx += 4
		}
	}
	return &Image{
		Width:    width,
		Height:   height,
		Pixels:   pixels,
		Format:   format,
		Owned:    true,
		HasAlpha: true,
	}
}

// ResizeMode defines the resampling algorithm.
type ResizeMode int

const (
	ResizeNearest ResizeMode = iota
	ResizeBilinear
	ResizeLanczos3
)

func (img *Image) Resize(width, height int, mode ResizeMode) *Image {
	img.mu.RLock()
	srcW := img.Width
	srcH := img.Height
	srcPixels := make([]uint8, len(img.Pixels))
	copy(srcPixels, img.Pixels)
	img.mu.RUnlock()

	if width <= 0 || height <= 0 {
		return nil
	}
	if srcW == width && srcH == height {
		return img.clone()
	}

	dstPixels := make([]uint8, width*height*4)
	switch mode {
	case ResizeBilinear:
		resizeBilinear(srcPixels, srcW, srcH, dstPixels, width, height)
	case ResizeLanczos3:
		resizeLanczos(srcPixels, srcW, srcH, dstPixels, width, height)
	default:
		resizeNearest(srcPixels, srcW, srcH, dstPixels, width, height)
	}

	return &Image{
		Width:  width,
		Height: height,
		Pixels: dstPixels,
		Format: img.Format,
		Owned:  true,
	}
}

// ResizeWithFilter resizes the image using the specified resampling filter/mode.
// It is a more explicitly named alternative to Resize that takes a ResizeMode.
func (img *Image) ResizeWithFilter(width, height int, mode ResizeMode) *Image {
	return img.Resize(width, height, mode)
}

func resizeNearest(src []uint8, srcW, srcH int, dst []uint8, dstW, dstH int) {
	xScale := float32(srcW) / float32(dstW)
	yScale := float32(srcH) / float32(dstH)
	for y := 0; y < dstH; y++ {
		sy := int(float32(y) * yScale)
		if sy >= srcH {
			sy = srcH - 1
		}
		srcRow := sy * srcW * 4
		dstRow := y * dstW * 4
		for x := 0; x < dstW; x++ {
			sx := int(float32(x) * xScale)
			if sx >= srcW {
				sx = srcW - 1
			}
			srcIdx := srcRow + sx*4
			dstIdx := dstRow + x*4
			dst[dstIdx] = src[srcIdx]
			dst[dstIdx+1] = src[srcIdx+1]
			dst[dstIdx+2] = src[srcIdx+2]
			dst[dstIdx+3] = src[srcIdx+3]
		}
	}
}

func resizeBilinear(src []uint8, srcW, srcH int, dst []uint8, dstW, dstH int) {
	xScale := float32(srcW) / float32(dstW)
	yScale := float32(srcH) / float32(dstH)
	for y := 0; y < dstH; y++ {
		fy := float32(y) * yScale
		sy0 := int(fy)
		if sy0 >= srcH-1 {
			sy0 = srcH - 2
		}
		sy1 := sy0 + 1
		fyFrac := fy - float32(sy0)
		row0 := sy0 * srcW * 4
		row1 := sy1 * srcW * 4
		dstRow := y * dstW * 4
		for x := 0; x < dstW; x++ {
			fx := float32(x) * xScale
			sx0 := int(fx)
			if sx0 >= srcW-1 {
				sx0 = srcW - 2
			}
			sx1 := sx0 + 1
			fxFrac := fx - float32(sx0)
			for c := 0; c < 4; c++ {
				v00 := float32(src[row0+sx0*4+c])
				v01 := float32(src[row0+sx1*4+c])
				v10 := float32(src[row1+sx0*4+c])
				v11 := float32(src[row1+sx1*4+c])
				v0 := v00 + (v01-v00)*fxFrac
				v1 := v10 + (v11-v10)*fxFrac
				v := v0 + (v1-v0)*fyFrac
				dst[dstRow+x*4+c] = uint8(v)
			}
		}
	}
}

func resizeLanczos(src []uint8, srcW, srcH int, dst []uint8, dstW, dstH int) {
	const lobes = 3
	xScale := float32(srcW) / float32(dstW)
	yScale := float32(srcH) / float32(dstH)
	lanczos := func(x float32) float32 {
		if x == 0 {
			return 1
		}
		if x <= -float32(lobes) || x >= float32(lobes) {
			return 0
		}
		pi := float32(math.Pi)
		xPi := x * pi
		s1 := float32(math.Sin(float64(xPi)))
		s2 := float32(math.Sin(float64(xPi / float32(lobes))))
		return float32(lobes) * s1 / xPi * s2 / (xPi / float32(lobes))
	}
	for y := 0; y < dstH; y++ {
		fy := float32(y) * yScale
		sy0 := int(fy) - lobes + 1
		if sy0 < 0 {
			sy0 = 0
		}
		sy1 := int(fy) + lobes
		if sy1 >= srcH {
			sy1 = srcH - 1
		}
		dstRow := y * dstW * 4
		for x := 0; x < dstW; x++ {
			fx := float32(x) * xScale
			sx0 := int(fx) - lobes + 1
			if sx0 < 0 {
				sx0 = 0
			}
			sx1 := int(fx) + lobes
			if sx1 >= srcW {
				sx1 = srcW - 1
			}
			dstIdx := dstRow + x*4
			for c := 0; c < 4; c++ {
				var acc float32
				var wsum float32
				for sy := sy0; sy <= sy1; sy++ {
					wy := lanczos(float32(sy) - fy)
					if wy == 0 {
						continue
					}
					row := sy * srcW * 4
					for sx := sx0; sx <= sx1; sx++ {
						wx := lanczos(float32(sx) - fx)
						if wx == 0 {
							continue
						}
						w := wx * wy
						acc += float32(src[row+sx*4+c]) * w
						wsum += w
					}
				}
				if wsum > 0 {
					acc /= wsum
				}
				if acc > 255 {
					acc = 255
				}
				if acc < 0 {
					acc = 0
				}
				dst[dstIdx+c] = uint8(acc)
			}
		}
	}
}

func (img *Image) clone() *Image {
	img.mu.RLock()
	defer img.mu.RUnlock()
	pixels := make([]uint8, len(img.Pixels))
	copy(pixels, img.Pixels)
	return &Image{
		ID:     img.ID,
		Width:  img.Width,
		Height: img.Height,
		Pixels: pixels,
		Format: img.Format,
		Owned:  true,
	}
}

func (img *Image) Crop(x, y, w, h int) *Image {
	img.mu.RLock()
	defer img.mu.RUnlock()
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+w > img.Width {
		w = img.Width - x
	}
	if y+h > img.Height {
		h = img.Height - y
	}
	if w <= 0 || h <= 0 {
		return NewImage(0, 0)
	}
	dst := NewImage(w, h)
	for row := 0; row < h; row++ {
		srcIdx := ((y+row)*img.Width + x) * 4
		dstIdx := row * w * 4
		copy(dst.Pixels[dstIdx:dstIdx+w*4], img.Pixels[srcIdx:srcIdx+w*4])
	}
	return dst
}

func (img *Image) Blend(dst *Image, dx, dy int) {
	img.mu.RLock()
	dst.mu.Lock()
	defer dst.mu.Unlock()
	defer img.mu.RUnlock()

	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			px := x + dx
			py := y + dy
			if px < 0 || px >= dst.Width || py < 0 || py >= dst.Height {
				continue
			}
			srcIdx := (y*img.Width + x) * 4
			dstIdx := (py*dst.Width + px) * 4
			sa := float32(img.Pixels[srcIdx+3]) / 255.0
			if sa <= 0 {
				continue
			}
			if sa >= 1.0 {
				dst.Pixels[dstIdx] = img.Pixels[srcIdx]
				dst.Pixels[dstIdx+1] = img.Pixels[srcIdx+1]
				dst.Pixels[dstIdx+2] = img.Pixels[srcIdx+2]
				dst.Pixels[dstIdx+3] = img.Pixels[srcIdx+3]
				continue
			}
			da := float32(dst.Pixels[dstIdx+3]) / 255.0
			oa := sa + da*(1-sa)
			if oa <= 0 {
				continue
			}
			dst.Pixels[dstIdx+3] = uint8(oa * 255)
			for c := 0; c < 3; c++ {
				s := float32(img.Pixels[srcIdx+c])
				d := float32(dst.Pixels[dstIdx+c])
				dst.Pixels[dstIdx+c] = uint8((s*sa + d*da*(1-sa)) / oa)
			}
		}
	}
}

func (img *Image) SetPixel(x, y int, r, g, b, a uint8) {
	img.mu.Lock()
	defer img.mu.Unlock()
	if x < 0 || x >= img.Width || y < 0 || y >= img.Height {
		return
	}
	idx := (y*img.Width + x) * 4
	img.Pixels[idx] = r
	img.Pixels[idx+1] = g
	img.Pixels[idx+2] = b
	img.Pixels[idx+3] = a
}

func (img *Image) GetPixel(x, y int) (uint8, uint8, uint8, uint8, bool) {
	img.mu.RLock()
	defer img.mu.RUnlock()
	if x < 0 || x >= img.Width || y < 0 || y >= img.Height {
		return 0, 0, 0, 0, false
	}
	idx := (y*img.Width + x) * 4
	return img.Pixels[idx], img.Pixels[idx+1], img.Pixels[idx+2], img.Pixels[idx+3], true
}

func (img *Image) Fill(r, g, b, a uint8) {
	img.mu.Lock()
	defer img.mu.Unlock()
	for i := 0; i < len(img.Pixels); i += 4 {
		img.Pixels[i] = r
		img.Pixels[i+1] = g
		img.Pixels[i+2] = b
		img.Pixels[i+3] = a
	}
}

func (img *Image) SavePNG(path string) error {
	img.mu.RLock()
	rgba := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	copy(rgba.Pix, img.Pixels)
	img.mu.RUnlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, rgba)
}

func (img *Image) SaveJPEG(path string, quality int) error {
	img.mu.RLock()
	rgba := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	copy(rgba.Pix, img.Pixels)
	img.mu.RUnlock()

	if quality < 1 {
		quality = 80
	}
	if quality > 100 {
		quality = 100
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, rgba, &jpeg.Options{Quality: quality})
}

func (img *Image) ToRGBA() *image.RGBA {
	img.mu.RLock()
	defer img.mu.RUnlock()
	rgba := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	copy(rgba.Pix, img.Pixels)
	return rgba
}

type ImageProtocol int

const (
	ImageProtocolAuto ImageProtocol = iota
	ImageProtocolKitty
	ImageProtocolSixel
	ImageProtocolBlocks
)

func (p ImageProtocol) String() string {
	switch p {
	case ImageProtocolAuto:
		return "auto"
	case ImageProtocolKitty:
		return "kitty"
	case ImageProtocolSixel:
		return "sixel"
	case ImageProtocolBlocks:
		return "blocks"
	default:
		return "unknown"
	}
}

func (img *Image) KittyEscape() (string, error) {
	return img.KittyEscapeWithOptions(KittyOptions{})
}

type KittyOptions struct {
	Action       string
	AttachmentID uint32
	Format       int
	Quality      int
	WidthCells   int
	HeightCells  int
	Compressed   bool
	Placement    int
	DeleteID     uint32
}

func (img *Image) KittyEscapeWithOptions(opts KittyOptions) (string, error) {
	img.mu.RLock()
	width := img.Width
	height := img.Height
	pixels := make([]uint8, len(img.Pixels))
	copy(pixels, img.Pixels)
	img.mu.RUnlock()

	var action byte
	if opts.Action == "" {
		action = 'T'
	} else {
		action = opts.Action[0]
	}

	var payload []byte
	if action == 'd' {
		return fmt.Sprintf("\x1b_Ga=d,A=%d\x1b\\", opts.DeleteID), nil
	}

	pngBuf := &bytes.Buffer{}
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(rgba.Pix, pixels)
	if err := png.Encode(pngBuf, rgba); err != nil {
		return "", fmt.Errorf("png encode failed: %w", err)
	}
	raw := pngBuf.Bytes()

	if opts.Compressed {
		var compBuf bytes.Buffer
		zw, err := zlib.NewWriterLevel(&compBuf, zlib.BestCompression)
		if err != nil {
			return "", err
		}
		if _, err := zw.Write(raw); err != nil {
			return "", err
		}
		if err := zw.Close(); err != nil {
			return "", err
		}
		payload = compBuf.Bytes()
	} else {
		payload = raw
	}

	encoded := base64.StdEncoding.EncodeToString(payload)

	var params string
	params += fmt.Sprintf("a=%c", action)
	if opts.Format != 0 {
		params += fmt.Sprintf(",f=%d", opts.Format)
	} else {
		params += ",f=32"
	}
	if opts.Quality > 0 {
		params += fmt.Sprintf(",q=%d", opts.Quality)
	}
	if opts.WidthCells > 0 {
		params += fmt.Sprintf(",W=%d", opts.WidthCells)
	}
	if opts.HeightCells > 0 {
		params += fmt.Sprintf(",H=%d", opts.HeightCells)
	}
	if opts.Compressed {
		params += ",z=1"
	}
	if opts.AttachmentID > 0 {
		params += fmt.Sprintf(",A=%d", opts.AttachmentID)
	}
	if opts.Placement > 0 {
		params += fmt.Sprintf(",m=%d", opts.Placement)
	}

	return fmt.Sprintf("\x1b_G%s;%s\x1b\\", params, encoded), nil
}

func (img *Image) KittyEscapeFrame(attachmentID uint32) (string, error) {
	opts := KittyOptions{
		Action:       "T",
		Format:       32,
		AttachmentID: attachmentID,
		Compressed:   true,
	}
	return img.KittyEscapeWithOptions(opts)
}

func (img *Image) BytesPerPixel() int {
	return 4
}

func (img *Image) PNGBytes() ([]byte, error) {
	img.mu.RLock()
	rgba := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	copy(rgba.Pix, img.Pixels)
	img.mu.RUnlock()
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (img *Image) MemorySize() int64 {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return int64(len(img.Pixels))
}

var _ = color.RGBA{}
var _ = crc32.ChecksumIEEE
