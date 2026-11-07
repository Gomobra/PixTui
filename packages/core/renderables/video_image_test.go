package renderables

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

func hasNonEmptyContent(access native.BufferAccess) bool {
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] != 0 && access.Char[idx] != ' ' {
				return true
			}
		}
	}
	return false
}

// TestVideoPlaceholder verifies VideoRenderable renders placeholder text when no frame.
func TestVideoPlaceholder(t *testing.T) {
	ctx := newTestCtx()
	vid := native.NewVideo(20, 5)
	vid.Frame = nil
	r := NewVideoRenderable(ctx, VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-placeholder",
			Visible: boolPtr(true),
		},
		AutoPlay: false,
		Video:    vid,
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(5)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	if !hasNonEmptyContent(buf.Access()) {
		t.Error("expected video placeholder text to be rendered")
	}
}

// TestVideoPlayPause verifies VideoRenderable play/pause state affects placeholder.
func TestVideoPlayPause(t *testing.T) {
	ctx := newTestCtx()
	vid := native.NewVideo(20, 5)
	vid.Frame = nil
	r := NewVideoRenderable(ctx, VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-play",
			Visible: boolPtr(true),
		},
		AutoPlay: false,
		Video:    vid,
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(5)

	// Paused → placeholder should render
	r.Pause()
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
	if !hasNonEmptyContent(buf.Access()) {
		t.Error("expected placeholder text when video is paused")
	}

	// Playing → placeholder should render
	r.Play()
	buf = core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
	if !hasNonEmptyContent(buf.Access()) {
		t.Error("expected placeholder text when video is playing")
	}
}

// TestVideoWithFrame verifies VideoRenderable renders frame pixels when available.
func TestVideoWithFrame(t *testing.T) {
	ctx := newTestCtx()
	vid := native.NewVideo(4, 4)
	vid.Frame = &native.Image{
		Width:  4,
		Height: 4,
		Pixels: []uint8{
			255, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255,
			0, 0, 0, 255, 255, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255,
			0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255, 255, 0, 0, 255,
			0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255,
		},
	}
	vid.Playing = false

	r := NewVideoRenderable(ctx, VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-frame",
			Visible: boolPtr(true),
		},
		Video: vid,
	})
	r.SetPosition(0, 0)
	r.SetWidth(4)
	r.SetHeight(2)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	if !hasNonEmptyContent(buf.Access()) {
		t.Error("expected video frame to render block characters")
	}
}

// TestImageRenderBlocks verifies ImageRenderable renders using blocks protocol.
func TestImageRenderBlocks(t *testing.T) {
	ctx := newTestCtx()
	img := native.NewImage(4, 4)
	img.Pixels = []uint8{
		255, 0, 0, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	r := NewImageRenderable(ctx, ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-blocks",
			Visible: boolPtr(true),
		},
		Image:    img,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	r.SetPosition(0, 0)
	r.SetWidth(4)
	r.SetHeight(2)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	if !hasNonEmptyContent(buf.Access()) {
		t.Error("expected image to render block characters")
	}
}

// TestImageEmptyPlaceholder verifies ImageRenderable shows placeholder when image has no pixels.
func TestImageEmptyPlaceholder(t *testing.T) {
	ctx := newTestCtx()
	img := &native.Image{Width: 10, Height: 10, Pixels: []uint8{}}
	r := NewImageRenderable(ctx, ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-empty",
			Visible: boolPtr(true),
		},
		Image: img,
	})
	r.SetPosition(0, 0)
	r.SetWidth(10)
	r.SetHeight(5)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	if !hasNonEmptyContent(buf.Access()) {
		t.Error("expected '[Image]' placeholder when image has no pixels")
	}
}

// TestImageSetImage verifies ImageRenderable SetImage updates the rendered image.
func TestImageSetImage(t *testing.T) {
	ctx := newTestCtx()
	img1 := native.NewImage(2, 2)
	img1.Pixels = []uint8{255, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255}

	r := NewImageRenderable(ctx, ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-set",
			Visible: boolPtr(true),
		},
		Image: img1,
	})
	r.SetPosition(0, 0)
	r.SetWidth(2)
	r.SetHeight(1)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	if !hasNonEmptyContent(buf.Access()) {
		t.Error("expected initial image to render block characters")
	}

	img2 := &native.Image{Width: 2, Height: 2, Pixels: []uint8{}}
	r.SetImage(img2)
	buf = core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	if !hasNonEmptyContent(buf.Access()) {
		t.Error("expected '[Image]' placeholder after SetImage with empty image")
	}
}

// TestNativeImageCreate verifies native.Image creation.
func TestNativeImageCreate(t *testing.T) {
	img := native.NewImage(100, 50)
	if img.Width != 100 || img.Height != 50 {
		t.Errorf("expected 100x50 image, got %dx%d", img.Width, img.Height)
	}
	if len(img.Pixels) != 100*50*4 {
		t.Errorf("expected %d pixels, got %d", 100*50*4, len(img.Pixels))
	}
}

// TestNativeVideoCreate verifies native.Video creation.
func TestNativeVideoCreate(t *testing.T) {
	vid := native.NewVideo(100, 50)
	if vid.Width != 100 || vid.Height != 50 {
		t.Errorf("expected 100x50 video, got %dx%d", vid.Width, vid.Height)
	}
	if vid.Playing {
		t.Error("expected video not playing initially")
	}
}
