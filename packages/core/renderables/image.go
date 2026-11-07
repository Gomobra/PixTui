package renderables

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// ImageSource is a source for an image.
type ImageSource interface {
	NativeImage() *native.Image
}

// NativeImageSource wraps a native.Image as an ImageSource.
type NativeImageSource struct {
	Image *native.Image
}

func (s *NativeImageSource) NativeImage() *native.Image { return s.Image }

// ImageOptions contains options for ImageRenderable.
type ImageOptions struct {
	core.RenderableOptions[core.Renderable]
	Source    ImageSource
	Path      string
	Image     *native.Image
	Fit       core.ImageFit
	Protocol  core.ImageRenderProtocol
	Crop      *ExtractOptions
	OnLoad    func(*native.Image)
	OnError   func(error)
}

// AbortController allows cancelling an in-flight image load.
type AbortController struct {
	mu     sync.Mutex
	aborted bool
}

func (a *AbortController) Abort() {
	a.mu.Lock()
	a.aborted = true
	a.mu.Unlock()
}

func (a *AbortController) SignalAborted() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.aborted
}

// ExtractOptions defines a crop region in source pixels.
type ExtractOptions struct {
	Left   int
	Top    int
	Width  int
	Height int
}

// ImageRenderable renders an image in the terminal using Kitty/sixel/blocks protocols.
type ImageRenderable struct {
	*core.Renderable
	options ImageOptions

	image      *native.Image
	pendingImg *native.Image
	loadErr    error
	loadCtrl   *AbortController
	loadPromise chan struct{}
	loading    bool
	crop       *ExtractOptions
	source     ImageSource
	mu         sync.RWMutex
}

// NewImageRenderable creates a new ImageRenderable.
func NewImageRenderable(ctx *core.RenderContext, opts ImageOptions) *ImageRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:            opts.Width,
		Height:           opts.Height,
		Visible:          opts.Visible,
		ZIndex:           opts.ZIndex,
		ForegroundColor:  opts.ForegroundColor,
		BackgroundColor:  opts.BackgroundColor,
		RenderBefore:     opts.RenderBefore,
		RenderAfter:      opts.RenderAfter,
		OnMouse:          opts.OnMouse,
		OnMouseDown:      opts.OnMouseDown,
		OnMouseUp:        opts.OnMouseUp,
		OnMouseMove:      opts.OnMouseMove,
		OnMouseDrag:      opts.OnMouseDrag,
		OnMouseDragEnd:   opts.OnMouseDragEnd,
		OnMouseDrop:      opts.OnMouseDrop,
		OnMouseOver:      opts.OnMouseOver,
		OnMouseOut:       opts.OnMouseOut,
		OnMouseScroll:    opts.OnMouseScroll,
		OnPaste:          opts.OnPaste,
		OnKeyDown:        opts.OnKeyDown,
	})
	r := &ImageRenderable{
		Renderable: base,
		options:    opts,
		loadCtrl:   &AbortController{},
		loadPromise: make(chan struct{}, 1),
	}
	if opts.Fit != "" {
		// fit is already core.ImageFit type
	} else {
		r.options.Fit = core.ImageFitFit
	}
	if opts.Protocol != "" {
		// protocol is already core.ImageRenderProtocol type
	} else {
		r.options.Protocol = core.ImageRenderProtocolAuto
	}
	if opts.Crop != nil {
		r.crop = opts.Crop
	}
	if opts.Source != nil {
		r.source = opts.Source
		r.image = opts.Source.NativeImage()
	} else if opts.Image != nil {
		r.image = opts.Image
	} else if opts.Path != "" {
		img, err := loadImageFromPath(opts.Path)
		if err != nil {
			img = native.NewImage(80, 24)
		}
		r.image = img
	} else {
		r.image = native.NewImage(80, 24)
	}
	r.loading = r.image == nil || r.image.Width <= 0 || r.image.Height <= 0
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindImage)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
		r.applyNativeImage()
	}
	return r
}

func loadImageFromPath(path string) (*native.Image, error) {
	img, err := native.LoadPNG(path)
	if err != nil {
		img, err = native.LoadJPEG(path)
	}
	if err != nil {
		img, err = native.LoadGIF(path)
	}
	if err != nil {
		img, err = native.LoadWebP(path)
	}
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Image returns the underlying image.
func (r *ImageRenderable) Image() *native.Image {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.image
}

// SetImage sets the image and requests a re-render.
func (r *ImageRenderable) SetImage(img *native.Image) {
	r.mu.Lock()
	r.image = img
	r.loading = img == nil || img.Width <= 0 || img.Height <= 0
	r.mu.Unlock()
	r.applyNativeImage()
	r.RequestRender()
}

// Options returns the image options.
func (r *ImageRenderable) Options() ImageOptions {
	return r.options
}

// SetOptions updates the image options and requests a re-render.
func (r *ImageRenderable) SetOptions(opts ImageOptions) {
	r.mu.Lock()
	r.options = opts
	if opts.Image != nil {
		r.image = opts.Image
	}
	if opts.Crop != nil {
		r.crop = opts.Crop
	}
	r.loading = r.image == nil || r.image.Width <= 0 || r.image.Height <= 0
	r.mu.Unlock()
	r.applyNativeImage()
	r.RequestRender()
}

// Source returns the image source.
func (r *ImageRenderable) Source() ImageSource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.source
}

// SetSource sets the image source and begins loading.
func (r *ImageRenderable) SetSource(source ImageSource) {
	r.mu.Lock()
	if source == r.source {
		r.mu.Unlock()
		return
	}
	if r.loadCtrl != nil {
		r.loadCtrl.Abort()
	}
	r.source = source
	r.image = nil
	if source != nil {
		r.image = source.NativeImage()
	}
	r.loading = r.image == nil || r.image.Width <= 0 || r.image.Height <= 0
	r.loadErr = nil
	r.mu.Unlock()
	r.applyNativeImage()
	r.RequestRender()
}

// Fit returns the fit mode.
func (r *ImageRenderable) Fit() core.ImageFit {
	r.mu.RLock(); defer r.mu.RUnlock()
	return r.options.Fit
}

// SetFit sets the fit mode.
func (r *ImageRenderable) SetFit(f core.ImageFit) {
	r.mu.Lock()
	r.options.Fit = f
	r.mu.Unlock()
	r.applyNativeImage()
	r.RequestRender()
}

// Crop returns the crop region.
func (r *ImageRenderable) Crop() *ExtractOptions {
	r.mu.RLock(); defer r.mu.RUnlock()
	if r.crop == nil { return nil }
	c := *r.crop
	return &c
}

// SetCrop sets the crop region.
func (r *ImageRenderable) SetCrop(c *ExtractOptions) {
	r.mu.Lock()
	r.crop = c
	r.mu.Unlock()
	r.applyNativeImage()
	r.RequestRender()
}

// Protocol returns the render protocol.
func (r *ImageRenderable) Protocol() core.ImageRenderProtocol {
	r.mu.RLock(); defer r.mu.RUnlock()
	return r.options.Protocol
}

// SetProtocol sets the render protocol.
func (r *ImageRenderable) SetProtocol(p core.ImageRenderProtocol) {
	r.mu.Lock()
	r.options.Protocol = p
	r.mu.Unlock()
	r.applyNativeImage()
	r.RequestRender()
}

// Loading returns whether the image is currently loading.
func (r *ImageRenderable) Loading() bool {
	r.mu.RLock(); defer r.mu.RUnlock()
	return r.loading
}

// LoadError returns any load error.
func (r *ImageRenderable) LoadError() error {
	r.mu.RLock(); defer r.mu.RUnlock()
	return r.loadErr
}

// LoadPromise returns a channel that closes when loading completes.
func (r *ImageRenderable) LoadPromise() <-chan struct{} {
	return r.loadPromise
}

// CellAspectRatio returns the terminal cell aspect ratio (height/width).
func (r *ImageRenderable) CellAspectRatio() float64 {
	return 2.0
}

// GetFittedSize computes the fitted display size for the image.
func (r *ImageRenderable) GetFittedSize(targetW, targetH int, cellAspect float64, sourceW, sourceH int) (int, int) {
	if sourceW <= 0 || sourceH <= 0 || targetW <= 0 || targetH <= 0 {
		return 0, 0
	}
	if cellAspect <= 0 { cellAspect = 2.0 }
	if r.options.Fit == core.ImageFitFill {
		return targetW, targetH
	}
	displayAspect := (float64(sourceW) / float64(sourceH)) * cellAspect
	var scale float64
	if r.options.Fit == core.ImageFitCover {
		scale = mathMax(float64(targetW)/displayAspect, float64(targetH))
	} else {
		scale = mathMin(float64(targetW)/displayAspect, float64(targetH))
	}
	w := int(mathRound(displayAspect * scale))
	h := int(mathRound(scale))
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return w, h
}

func mathMin(a, b float64) float64 { if a < b { return a }; return b }
func mathMax(a, b float64) float64 { if a > b { return a }; return b }
func mathRound(f float64) float64 {
	if f < 0 { return -mathRound(-f) }
	return float64(int(f + 0.5))
}

func (r *ImageRenderable) applyNativeImage() {
	ns := r.GetNativeScene()
	if ns == nil { return }
	r.mu.RLock()
	img := r.image
	fit := r.options.Fit
	proto := r.options.Protocol
	r.mu.RUnlock()
	_ = ns.SetImage(r.Renderable, img, fit, proto, nil)
}

// RenderSelf renders the image into the buffer.
func (r *ImageRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}

	r.mu.RLock()
	img := r.image
	proto := r.options.Protocol
	crop := r.crop
	r.mu.RUnlock()

	if img == nil || img.Width <= 0 || img.Height <= 0 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	x := int(r.GetX())
	y := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())

	if w < 1 || h < 1 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > buffer.Width() {
		w = buffer.Width() - x
	}
	if y+h > buffer.Height() {
		h = buffer.Height() - y
	}
	if w <= 0 || h <= 0 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	// Fill background with the renderable's background color.
	bg := r.options.BackgroundColor
	if bg != nil {
		fillRect(buffer, x, y, w, h, ' ', nil, bg, 0)
	}

	if proto == core.ImageRenderProtocolAuto {
		proto = core.ImageRenderProtocolBlocks
	}

	switch proto {
	case core.ImageRenderProtocolKitty:
		r.renderKitty(buffer, x, y, w, h)
	case core.ImageRenderProtocolSixel:
		r.renderSixel(buffer, x, y, w, h)
	default:
		r.renderBlocks(buffer, x, y, w, h, r.options.Fit, crop)
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

func (r *ImageRenderable) renderKitty(buffer *core.NativeBuffer, x, y, w, h int) {
	r.mu.RLock()
	img := r.image
	r.mu.RUnlock()
	if img == nil {
		r.renderBlocks(buffer, x, y, w, h, r.options.Fit, nil)
		return
	}
	escape, err := img.KittyEscape()
	if err == nil && escape != "" {
		r.writeTextAt(buffer, x, y, escape, nil, nil, 0)
		y++
		h--
	}
	if h > 0 && w > 0 {
		r.renderBlocks(buffer, x, y, w, h, r.options.Fit, nil)
	}
}

func (r *ImageRenderable) renderSixel(buffer *core.NativeBuffer, x, y, w, h int) {
	r.mu.RLock()
	img := r.image
	r.mu.RUnlock()
	if img == nil || len(img.Pixels) == 0 {
		r.renderBlocks(buffer, x, y, w, h, r.options.Fit, nil)
		return
	}
	var sb strings.Builder
	sb.WriteString("\x1bP")
	sb.WriteString("0;0;0;0#")
	imgW := img.Width
	imgH := img.Height
	for row := 0; row < imgH; row++ {
		sb.WriteString("-")
		for col := 0; col < imgW; col++ {
			idx := (row*imgW + col) * 4
			if idx+3 >= len(img.Pixels) { break }
			rr := img.Pixels[idx]
			gg := img.Pixels[idx+1]
			bb := img.Pixels[idx+2]
			aa := img.Pixels[idx+3]
			if aa == 0 {
				sb.WriteString("@")
				continue
			}
			sb.WriteString(fmt.Sprintf("#0;2;%d;%d;%d#", rr, gg, bb))
			sb.WriteString("\"")
		}
	}
	sb.WriteString("\x1b\\")
	r.writeTextAt(buffer, x, y, sb.String(), nil, nil, 0)
	y++
	h--
	if h > 0 && w > 0 {
		r.renderBlocks(buffer, x, y, w, h, r.options.Fit, nil)
	}
}

func (r *ImageRenderable) renderBlocks(buffer *core.NativeBuffer, x, y, w, h int, fit core.ImageFit, crop *ExtractOptions) {
	r.mu.RLock()
	img := r.image
	r.mu.RUnlock()
	if img == nil || len(img.Pixels) == 0 {
		r.writeTextAt(buffer, x, y, "[Image]", nil, nil, 0)
		return
	}

	imgW := img.Width
	imgH := img.Height
	pixels := img.Pixels

	sx0, sy0, sx1, sy1 := 0, 0, imgW, imgH
	if crop != nil {
		sx0 = crop.Left
		sy0 = crop.Top
		sx1 = crop.Left + crop.Width
		sy1 = crop.Top + crop.Height
		if sx0 < 0 { sx0 = 0 }
		if sy0 < 0 { sy0 = 0 }
		if sx1 > imgW { sx1 = imgW }
		if sy1 > imgH { sy1 = imgH }
	}
	cropW := sx1 - sx0
	cropH := sy1 - sy0
	if cropW <= 0 || cropH <= 0 {
		r.writeTextAt(buffer, x, y, "[Image]", nil, nil, 0)
		return
	}

	fw, fh := r.GetFittedSize(w, h, r.CellAspectRatio(), cropW, cropH)
	if fw <= 0 || fh <= 0 {
		r.writeTextAt(buffer, x, y, "[Image]", nil, nil, 0)
		return
	}
	ox := x + (w-fw)/2
	oy := y + (h-fh)/2

	sx := float64(cropW) / float64(fw)
	sy := float64(cropH) / float64(fh*2)

	for row := 0; row < fh; row++ {
		screenY := oy + row
		if screenY >= buffer.Height() { break }
		for col := 0; col < fw; col++ {
			screenX := ox + col
			if screenX >= buffer.Width() { break }
			px0 := int(float64(col) * sx)
			px1 := int(float64(col+1) * sx)
			py0 := int(float64(row) * sy * 2)
			py1 := int(float64(row+1) * sy * 2)
			if px1 > cropW { px1 = cropW }
			if py1 > cropH { py1 = cropH }
			if px0 >= px1 || py0 >= py1 {
				writeCell(buffer, screenX, screenY, ' ', nil, nil, 0)
				continue
			}
			upper := averagePixel(pixels, imgW, sx0+px0, sx0+px1, sy0+py0, sy0+(py0+py1)/2)
			lower := averagePixel(pixels, imgW, sx0+px0, sx0+px1, sy0+(py0+py1)/2, sy0+py1)
			ch, upperEmpty, lowerEmpty := blockChar(upper, lower)
			var fg, bg *core.RGBA
			if !upperEmpty && !lowerEmpty {
				ch = '█'
				fg = rgbaPtr(upper)
			} else if !upperEmpty && lowerEmpty {
				ch = '▄'
				fg = rgbaPtr(upper)
				bg = rgbaPtr(core.RGBA{R: 0, G: 0, B: 0, A: 0})
			} else if upperEmpty && !lowerEmpty {
				ch = '▀'
				bg = rgbaPtr(lower)
			} else {
				writeCell(buffer, screenX, screenY, ' ', nil, nil, 0)
				continue
			}
			writeCell(buffer, screenX, screenY, ch, fg, bg, 0)
		}
	}
}

func (r *ImageRenderable) writeTextAt(buffer *core.NativeBuffer, x, y int, text string, fg, bg *core.RGBA, attrs uint32) {
	cx := x
	for _, ch := range text {
		if ch == '\n' {
			cx = x
			y++
			continue
		}
		if cx >= buffer.Width() || y >= buffer.Height() { return }
		writeCell(buffer, cx, y, uint32(ch), fg, bg, attrs)
		cx++
	}
}

func averagePixel(pixels []uint8, imgW, px0, px1, py0, py1 int) core.RGBA {
	var r, g, b, a uint32
	count := 0
	for py := py0; py < py1; py++ {
		if py < 0 || py >= len(pixels)/(imgW*4) { continue }
		rowOffset := py * imgW * 4
		for px := px0; px < px1; px++ {
			if px < 0 || px >= imgW { continue }
			idx := rowOffset + px*4
			if idx+3 >= len(pixels) { continue }
			r += uint32(pixels[idx])
			g += uint32(pixels[idx+1])
			b += uint32(pixels[idx+2])
			a += uint32(pixels[idx+3])
			count++
		}
	}
	if count == 0 {
		return core.RGBA{R: 0, G: 0, B: 0, A: 0}
	}
	return core.RGBA{
		R: uint8(r / uint32(count)),
		G: uint8(g / uint32(count)),
		B: uint8(b / uint32(count)),
		A: uint8(a / uint32(count)),
	}
}

func blockChar(upper, lower core.RGBA) (uint32, bool, bool) {
	upperEmpty := upper.A == 0 || (upper.R == 0 && upper.G == 0 && upper.B == 0 && upper.A == 0)
	lowerEmpty := lower.A == 0 || (lower.R == 0 && lower.G == 0 && lower.B == 0 && lower.A == 0)
	if !upperEmpty && !lowerEmpty {
		return '█', false, false
	}
	if !upperEmpty {
		return '▄', false, true
	}
	if !lowerEmpty {
		return '▀', true, false
	}
	return ' ', true, true
}

func rgbaPtr(c core.RGBA) *core.RGBA { return &c }

var _ = core.ImageFitFit
var _ = native.ImageProtocolAuto
