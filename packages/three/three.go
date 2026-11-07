package three

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/Gomobra/PixTui/packages/core"
)

// Renderer represents a WebGPU renderer for terminal output.
// Since Go doesn't have native WebGPU in terminal, this provides
// a CPU-based 3D rendering pipeline that outputs to terminal cells.
type Renderer struct {
	width          int
	height         int
	device         *Device
	scene          *Scene
	camera         *Camera
	shader         *Shader
	material       *Material
	geometry       *Geometry
	spriteRenderer *SpriteRenderer
	physics        PhysicsEngine

	clearColor    color.Color
	doRenderStats bool
	renderTimeMs  float64
	destroyed     bool
}

// NewRenderer creates a new WebGPU renderer.
func NewRenderer(width, height int) *Renderer {
	r := &Renderer{
		width:      width,
		height:     height,
		device:     NewDevice(),
		scene:      NewScene(),
		camera:     NewCamera(),
		clearColor: color.RGBA{R: 0, G: 0, B: 0, A: 255},
	}
	r.spriteRenderer = NewSpriteRenderer(r)
	return r
}

// SetScene sets the scene to render.
func (r *Renderer) SetScene(scene *Scene) {
	r.scene = scene
}

// SetCamera sets the camera for rendering.
func (r *Renderer) SetCamera(camera *Camera) {
	r.camera = camera
}

// SetPhysics sets the physics engine for the renderer.
func (r *Renderer) SetPhysics(physics PhysicsEngine) {
	r.physics = physics
}

// GetPhysics returns the physics engine.
func (r *Renderer) GetPhysics() PhysicsEngine {
	return r.physics
}

// SetShader sets the active shader.
func (r *Renderer) SetShader(shader *Shader) {
	r.shader = shader
}

// GetShader returns the active shader.
func (r *Renderer) GetShader() *Shader {
	return r.shader
}

// SetMaterial sets the active material.
func (r *Renderer) SetMaterial(material *Material) {
	r.material = material
}

// GetMaterial returns the active material.
func (r *Renderer) GetMaterial() *Material {
	return r.material
}

// SetGeometry sets the active geometry.
func (r *Renderer) SetGeometry(geometry *Geometry) {
	r.geometry = geometry
}

// GetGeometry returns the active geometry.
func (r *Renderer) GetGeometry() *Geometry {
	return r.geometry
}

// SetClearColor sets the buffer clear color.
func (r *Renderer) SetClearColor(c color.Color) {
	r.clearColor = c
}

// GetClearColor returns the buffer clear color.
func (r *Renderer) GetClearColor() color.Color {
	if r.clearColor == nil {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	return r.clearColor
}

// SetActiveCamera sets the active camera for rendering.
func (r *Renderer) SetActiveCamera(camera *Camera) {
	r.camera = camera
}

// GetActiveCamera returns the active camera.
func (r *Renderer) GetActiveCamera() *Camera {
	return r.camera
}

// ToggleDebugStats toggles the on-screen render statistics overlay.
func (r *Renderer) ToggleDebugStats() {
	r.doRenderStats = !r.doRenderStats
}

// RenderStats draws render timing statistics onto the buffer.
func (r *Renderer) RenderStats(buffer *core.NativeBuffer) {
	r.renderStats(buffer, r.renderTimeMs)
}

func (r *Renderer) renderStats(buffer *core.NativeBuffer, renderTimeMs float64) {
	stats := []string{
		fmt.Sprintf("Three.js Renderer Stats:"),
		fmt.Sprintf("  Render: %.2fms", renderTimeMs),
		fmt.Sprintf("  Size: %dx%d", r.width, r.height),
	}
	fg := color.RGBA{R: 230, G: 230, B: 230, A: 255}
	for i, line := range stats {
		for x, ch := range []byte(line) {
			if x >= buffer.Width() {
				break
			}
			buffer.SetCell(x, 2+i, uint32(ch), uint16(fg.R), uint16(fg.G), 0)
		}
	}
}

// Destroy releases renderer resources.
func (r *Renderer) Destroy() {
	r.destroyed = true
	r.scene = nil
	r.camera = nil
	r.spriteRenderer = nil
}

// SetSize resizes the renderer output dimensions.
func (r *Renderer) SetSize(width, height int) {
	r.width = width
	r.height = height
	if r.camera != nil {
		r.camera.Aspect = float64(width) / float64(height)
	}
}

// Resize is an alias for SetSize.
func (r *Renderer) Resize(width, height int) {
	r.SetSize(width, height)
}

// Render renders the 3D scene to the terminal buffer using camera, scene, lighting, and materials.
func (r *Renderer) Render(buffer *core.NativeBuffer, deltaTime float64) {
	if r.destroyed {
		return
	}
	if r.scene == nil || r.camera == nil {
		return
	}

	start := float64(time.Now().UnixNano()) / 1e6

	// Clear buffer with background / clear color
	buffer.Clear()
	bg := r.scene.GetBackgroundColor()
	if r.clearColor != nil {
		bg = r.clearColor
	}
	if cr, cg, cb, ca := bg.RGBA(); ca > 0 {
		r.fillBuffer(buffer, uint8(cr>>8), uint8(cg>>8), uint8(cb>>8), uint8(ca>>8))
	}

	// Update physics if available and apply resulting transforms to scene objects
	if r.physics != nil {
		r.physics.Step(deltaTime)
		r.applyPhysicsToObjects(deltaTime)
	}

	// Get scene objects
	objects := r.scene.GetObjects()
	if len(objects) == 0 {
		return
	}

	// Compute lighting from scene lights
	lights := r.scene.GetLights()

	// Simple CPU rasterization: project 3D points to 2D screen
	for _, obj := range objects {
		if !r.isObjectVisible(obj) {
			continue
		}
		// Use enhanced lighting when a material is present, otherwise simple flat shading.
		if r.material != nil {
			r.renderObjectEnhanced(obj, buffer, lights)
		} else {
			r.renderObject(obj, buffer, lights)
		}
	}

	// Render sprites
	if r.spriteRenderer != nil {
		r.spriteRenderer.Render(buffer)
	}

	// Render debug stats if enabled
	if r.doRenderStats {
		r.renderStats(buffer, float64(time.Now().UnixNano())/1e6-start)
	}
	r.renderTimeMs = float64(time.Now().UnixNano())/1e6 - start
}

// fillBuffer fills the entire buffer with a character and color.
func (r *Renderer) fillBuffer(buffer *core.NativeBuffer, r_, g_, b_, a_ uint8) {
	for y := 0; y < buffer.Height(); y++ {
		for x := 0; x < buffer.Width(); x++ {
			buffer.SetCell(x, y, ' ', 0, uint16(r_), 0)
		}
	}
}

// isObjectVisible returns false if the object is a Mesh with Visible == false.
func (r *Renderer) isObjectVisible(obj Object) bool {
	if m, ok := obj.(*Mesh); ok {
		return m.Visible
	}
	return true
}

// applyPhysicsToObjects applies physics body transforms to scene objects that expose a physics body.
func (r *Renderer) applyPhysicsToObjects(deltaTime float64) {
	for _, obj := range r.scene.GetObjects() {
		if m, ok := obj.(*Mesh); ok && m.PhysicsBody != nil {
			pos := m.PhysicsBody.GetTranslation()
			rot := m.PhysicsBody.GetRotation()
			m.Position = Vector3{X: pos.X, Y: pos.Y, Z: m.Position.Z}
			q := IdentityQuaternion()
			q.SetFromAxisAngle(Vector3{X: 0, Y: 0, Z: 1}, rot)
			m.RotationQuat = q
		}
	}
}

func (r *Renderer) renderObject(obj Object, buffer *core.NativeBuffer, lights []*Light) {
	vertices := r.transformedVertices(obj)
	indices := obj.GetIndices()

	material := r.material
	objColor := obj.GetColor()
	if material != nil && material.Color != nil {
		objColor = material.Color
	}

	// Apply ambient light contribution
	finalColor := objColor
	for _, l := range lights {
		if l.Type == LightAmbient {
			rCol, gCol, bCol, _ := l.Color.RGBA()
			cr, cg, cb, _ := objColor.RGBA()
			finalColor = color.RGBA{
				R: uint8(minInt(255, int(cr>>8)+int(rCol>>8)*int(l.Intensity))),
				G: uint8(minInt(255, int(cg>>8)+int(gCol>>8)*int(l.Intensity))),
				B: uint8(minInt(255, int(cb>>8)+int(bCol>>8)*int(l.Intensity))),
				A: 255,
			}
		}
	}

	for i := 0; i < len(indices); i += 3 {
		if i+2 >= len(indices) {
			break
		}
		v0 := vertices[indices[i]]
		v1 := vertices[indices[i+1]]
		v2 := vertices[indices[i+2]]

		// Project 3D to 2D using camera
		x0, y0 := r.projectWithCamera(v0.X, v0.Y, v0.Z)
		x1, y1 := r.projectWithCamera(v1.X, v1.Y, v1.Z)
		x2, y2 := r.projectWithCamera(v2.X, v2.Y, v2.Z)

		// Draw triangle outline
		r.drawLine(x0, y0, x1, y1, buffer, finalColor)
		r.drawLine(x1, y1, x2, y2, buffer, finalColor)
		r.drawLine(x2, y2, x0, y0, buffer, finalColor)
	}
}

func (r *Renderer) project(x, y, z float64) (int, int) {
	// Simple perspective projection
	focalLength := float64(r.width) / 2.0
	scale := focalLength / (z + focalLength + 5.0)

	screenX := int(float64(r.width)/2 + x*scale*float64(r.width)/2)
	screenY := int(float64(r.height)/2 - y*scale*float64(r.height)/2)

	return screenX, screenY
}

// projectWithCamera projects a 3D point using the active camera (perspective or orthographic).
func (r *Renderer) projectWithCamera(x, y, z float64) (int, int) {
	cam := r.camera
	if cam == nil {
		return r.project(x, y, z)
	}
	var scale float64
	if cam.Mode == CameraOrthographic {
		scale = cam.OrthoSize
		screenX := int(float64(r.width)/2 + x*scale/cam.Aspect)
		screenY := int(float64(r.height)/2 - y*scale)
		return screenX, screenY
	}
	// Perspective projection using FOV and near plane.
	focalLength := 1.0 / math.Tan(cam.FOV/2)
	// depth from camera (camera looks down -Z; z is world-space depth relative to camera)
	depth := z + cam.Near + 2.0
	if depth < 0.001 {
		depth = 0.001
	}
	scale = focalLength / depth
	// Apply aspect ratio to X so circles stay circular on non-square terminals.
	screenX := int(float64(r.width)/2 + x*scale*float64(r.width)/2/cam.Aspect)
	screenY := int(float64(r.height)/2 - y*scale*float64(r.height)/2)
	return screenX, screenY
}

func (r *Renderer) drawLine(x0, y0, x1, y1 int, buffer *core.NativeBuffer, c color.Color) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	sy := 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		if x0 >= 0 && x0 < buffer.Width() && y0 >= 0 && y0 < buffer.Height() {
			rCol, gCol, _, _ := c.RGBA()
			buffer.SetCell(x0, y0, '#', uint16(rCol>>8), uint16(gCol>>8), 0)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// RenderSprite renders a single sprite to the buffer.
func (r *Renderer) RenderSprite(sprite *Sprite, buffer *core.NativeBuffer) {
	if !sprite.Visible || sprite.Texture == nil {
		return
	}
	// Project sprite center to screen
	x, y := r.projectWithCamera(sprite.Position.X, sprite.Position.Y, sprite.Position.Z)
	// Compute screen size based on scale and texture aspect
	texAspect := float64(sprite.Texture.Width) / float64(sprite.Texture.Height)
	if texAspect == 0 {
		texAspect = 1
	}
	screenW := int(sprite.Scale.X * float64(r.width) * 0.1)
	screenH := int(float64(screenW) / texAspect * sprite.Scale.Y)
	if screenW < 1 {
		screenW = 1
	}
	if screenH < 1 {
		screenH = 1
	}
	chars := []byte{' ', '.', ':', '-', '=', '+', '#', '%', '@'}
	rCol, gCol, _, _ := sprite.Color.RGBA()
	for dy := 0; dy < screenH; dy++ {
		for dx := 0; dx < screenW; dx++ {
			sx := x - screenW/2 + dx
			sy := y - screenH/2 + dy
			if sx < 0 || sx >= buffer.Width() || sy < 0 || sy >= buffer.Height() {
				continue
			}
			tx := int(float64(dx) / float64(screenW) * float64(sprite.Texture.Width))
			ty := int(float64(dy) / float64(screenH) * float64(sprite.Texture.Height))
			if tx < 0 {
				tx = 0
			}
			if tx >= sprite.Texture.Width {
				tx = sprite.Texture.Width - 1
			}
			if ty < 0 {
				ty = 0
			}
			if ty >= sprite.Texture.Height {
				ty = sprite.Texture.Height - 1
			}
			idx := (ty*sprite.Texture.Width + tx) * 4
			if idx+3 < len(sprite.Texture.Data) {
				a := sprite.Texture.Data[idx+3]
				if a < 128 {
					continue
				}
				lum := (int(sprite.Texture.Data[idx]) + int(sprite.Texture.Data[idx+1]) + int(sprite.Texture.Data[idx+2])) / 3
				ch := chars[min(lum*len(chars)/256, len(chars)-1)]
				buffer.SetCell(sx, sy, uint32(ch), uint16(rCol>>8), uint16(gCol>>8), 0)
			}
		}
	}
}

// RenderParticleSystem renders all active particles of a particle system.
func (r *Renderer) RenderParticleSystem(ps *ParticleSystem, buffer *core.NativeBuffer) {
	ps.Render(r, buffer)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Device represents a WebGPU device abstraction.
type Device struct {
	features []string
	limits   map[string]float64
}

// NewDevice creates a new device.
func NewDevice() *Device {
	return &Device{
		features: []string{},
		limits:   make(map[string]float64),
	}
}

// ============================================================
// NEW: Shader, Material, Geometry
// ============================================================

// Shader represents a WebGPU-like shader with vertex and fragment source.
type Shader struct {
	ID             string
	VertexSource   string
	FragmentSource string
}

// NewShader creates a new shader.
func NewShader(id, vertexSource, fragmentSource string) *Shader {
	return &Shader{
		ID:             id,
		VertexSource:   vertexSource,
		FragmentSource: fragmentSource,
	}
}

// BlendMode defines how a material blends with the framebuffer.
type BlendMode int

const (
	BlendNone BlendMode = iota
	BlendAlpha
	BlendAdditive
	BlendMultiply
)

// Material describes surface appearance: shader, uniforms, blend mode.
type Material struct {
	ID          string
	Shader      *Shader
	Uniforms    map[string]interface{}
	BlendMode   BlendMode
	Color       color.Color
	Transparent bool
	AlphaTest   float64
}

// NewMaterial creates a new material.
func NewMaterial(shader *Shader) *Material {
	return &Material{
		ID:          "",
		Shader:      shader,
		Uniforms:    make(map[string]interface{}),
		BlendMode:   BlendAlpha,
		Color:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Transparent: false,
		AlphaTest:   0.0,
	}
}

// SetUniform sets a uniform value on the material.
func (m *Material) SetUniform(name string, value interface{}) {
	if m.Uniforms == nil {
		m.Uniforms = make(map[string]interface{})
	}
	m.Uniforms[name] = value
}

// GetUniform returns a uniform value from the material.
func (m *Material) GetUniform(name string) (interface{}, bool) {
	v, ok := m.Uniforms[name]
	return v, ok
}

// Geometry holds vertex data, indices, and per-instance attributes.
type Geometry struct {
	ID         string
	Vertices   []Vector3
	Indices    []int
	Attributes map[string][]float32
}

// NewGeometry creates a new empty geometry.
func NewGeometry() *Geometry {
	return &Geometry{
		Vertices:   make([]Vector3, 0),
		Indices:    make([]int, 0),
		Attributes: make(map[string][]float32),
	}
}

// SetAttribute sets a per-vertex or per-instance attribute.
func (g *Geometry) SetAttribute(name string, data []float32) {
	if g.Attributes == nil {
		g.Attributes = make(map[string][]float32)
	}
	g.Attributes[name] = data
}

// GetAttribute returns a per-vertex or per-instance attribute.
func (g *Geometry) GetAttribute(name string) ([]float32, bool) {
	d, ok := g.Attributes[name]
	return d, ok
}

// ============================================================
// NEW: Vector2, Rect, Texture
// ============================================================

// Vector2 is a 2D vector.
type Vector2 struct {
	X, Y float64
}

// Rect represents a rectangle (used for UV regions / sprite sheets).
type Rect struct {
	X, Y, Width, Height float64
}

// Texture holds image data for sprites.
type Texture struct {
	ID     string
	Width  int
	Height int
	Data   []uint8 // RGBA bytes
	Format string
}

// NewTexture creates a new texture from RGBA data.
func NewTexture(id string, width, height int, data []uint8) *Texture {
	return &Texture{
		ID:     id,
		Width:  width,
		Height: height,
		Data:   data,
		Format: "rgba",
	}
}

// ============================================================
// NEW: Enhanced Mesh (uses Geometry + Material)
// ============================================================

// Mesh is a renderable object combining Geometry and Material.
type Mesh struct {
	ID           string
	Geometry     *Geometry
	Material     *Material
	Position     Vector3
	Rotation     Vector3
	RotationQuat Quaternion
	Scale        Vector3
	Visible      bool
	PhysicsBody  PhysicsRigidBody
}

// NewMesh creates a new mesh from geometry and material.
func NewMesh(geometry *Geometry, material *Material) *Mesh {
	return &Mesh{
		Geometry:     geometry,
		Material:     material,
		Position:     Vector3{X: 0, Y: 0, Z: 0},
		Rotation:     Vector3{X: 0, Y: 0, Z: 0},
		RotationQuat: IdentityQuaternion(),
		Scale:        Vector3{X: 1, Y: 1, Z: 1},
		Visible:      true,
	}
}

// GetVertices returns the mesh vertices.
func (m *Mesh) GetVertices() []Vector3 {
	if m.Geometry != nil {
		return m.Geometry.Vertices
	}
	return nil
}

// GetIndices returns the mesh indices.
func (m *Mesh) GetIndices() []int {
	if m.Geometry != nil {
		return m.Geometry.Indices
	}
	return nil
}

// GetColor returns the mesh color.
func (m *Mesh) GetColor() color.Color {
	if m.Material != nil && m.Material.Color != nil {
		return m.Material.Color
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

// ============================================================
// NEW: Sprite
// ============================================================

// Sprite is a textured billboard rendered by SpriteRenderer.
type Sprite struct {
	ID       string
	Texture  *Texture
	Position Vector3
	Rotation float64
	Scale    Vector2
	UVRect   Rect
	Visible  bool
	Color    color.Color
}

// NewSprite creates a new sprite.
func NewSprite(texture *Texture) *Sprite {
	return &Sprite{
		Texture:  texture,
		Position: Vector3{X: 0, Y: 0, Z: 0},
		Rotation: 0,
		Scale:    Vector2{X: 1, Y: 1},
		UVRect:   Rect{X: 0, Y: 0, Width: 1, Height: 1},
		Visible:  true,
		Color:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
}

// ============================================================
// NEW: Lighting, Fog
// ============================================================

// LightType identifies the kind of light source.
type LightType int

const (
	LightAmbient LightType = iota
	LightDirectional
	LightPoint
)

// Light represents a light source in the scene.
type Light struct {
	ID        string
	Type      LightType
	Color     color.Color
	Intensity float64
	Position  Vector3
	Direction Vector3
	Range     float64
}

// NewLight creates a new light.
func NewLight(t LightType, color color.Color, intensity float64) *Light {
	return &Light{
		ID:        "",
		Type:      t,
		Color:     color,
		Intensity: intensity,
		Position:  Vector3{X: 0, Y: 0, Z: 0},
		Direction: Vector3{X: 0, Y: -1, Z: 0},
		Range:     10.0,
	}
}

// Fog describes atmospheric fog.
type Fog struct {
	Enabled bool
	Color   color.Color
	Near    float64
	Far     float64
	Density float64
}

// NewFog creates default fog settings.
func NewFog() Fog {
	return Fog{
		Enabled: false,
		Color:   color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Near:    1.0,
		Far:     100.0,
		Density: 0.01,
	}
}

// ============================================================
// NEW: Enhanced Camera (perspective + orthographic)
// ============================================================

// CameraMode selects the projection type.
type CameraMode int

const (
	CameraPerspective CameraMode = iota
	CameraOrthographic
)

// Camera represents a 3D camera with perspective or orthographic projection.
type Camera struct {
	Position         Vector3
	Target           Vector3
	Up               Vector3
	Mode             CameraMode
	FOV              float64
	Aspect           float64
	Near             float64
	Far              float64
	OrthoSize        float64
	ProjectionMatrix Matrix4
}

// NewCamera creates a new perspective camera.
func NewCamera() *Camera {
	c := &Camera{
		Position:  Vector3{X: 0, Y: 0, Z: 5},
		Target:    Vector3{X: 0, Y: 0, Z: 0},
		Up:        Vector3{X: 0, Y: 1, Z: 0},
		Mode:      CameraPerspective,
		FOV:       60.0 * math.Pi / 180.0,
		Aspect:    1.0,
		Near:      0.1,
		Far:       1000.0,
		OrthoSize: 5.0,
	}
	c.UpdateProjectionMatrix()
	return c
}

// NewPerspectiveCamera creates a new perspective camera.
func NewPerspectiveCamera(fovDegrees, aspect, near, far float64) *Camera {
	c := &Camera{
		Position: Vector3{X: 0, Y: 0, Z: 5},
		Target:   Vector3{X: 0, Y: 0, Z: 0},
		Up:       Vector3{X: 0, Y: 1, Z: 0},
		Mode:     CameraPerspective,
		FOV:      fovDegrees * math.Pi / 180.0,
		Aspect:   aspect,
		Near:     near,
		Far:      far,
	}
	c.UpdateProjectionMatrix()
	return c
}

// NewOrthographicCamera creates a new orthographic camera.
func NewOrthographicCamera(size, aspect, near, far float64) *Camera {
	c := &Camera{
		Position:  Vector3{X: 0, Y: 0, Z: 5},
		Target:    Vector3{X: 0, Y: 0, Z: 0},
		Up:        Vector3{X: 0, Y: 1, Z: 0},
		Mode:      CameraOrthographic,
		FOV:       0,
		Aspect:    aspect,
		Near:      near,
		Far:       far,
		OrthoSize: size,
	}
	c.UpdateProjectionMatrix()
	return c
}

// UpdateProjectionMatrix builds the projection matrix based on the camera mode and parameters.
func (c *Camera) UpdateProjectionMatrix() {
	if c.Mode == CameraPerspective {
		f := 1.0 / math.Tan(c.FOV/2)
		c.ProjectionMatrix = Matrix4{
			Elements: [16]float64{
				f / c.Aspect, 0, 0, 0,
				0, f, 0, 0,
				0, 0, (c.Far + c.Near) / (c.Near - c.Far), -1,
				0, 0, (2 * c.Far * c.Near) / (c.Near - c.Far), 0,
			},
		}
	} else {
		// Orthographic projection
		o := c.OrthoSize
		// aspect-corrected orthographic bounds
		c.ProjectionMatrix = Matrix4{
			Elements: [16]float64{
				1 / (o * c.Aspect), 0, 0, 0,
				0, 1 / o, 0, 0,
				0, 0, -2 / (c.Far - c.Near), 0,
				0, 0, -(c.Far + c.Near) / (c.Far - c.Near), 1,
			},
		}
	}
}

// SetPosition sets the camera position.
func (c *Camera) SetPosition(x, y, z float64) {
	c.Position = Vector3{X: x, Y: y, Z: z}
}

// LookAt sets the camera target to look at.
func (c *Camera) LookAt(x, y, z float64) {
	c.Target = Vector3{X: x, Y: y, Z: z}
}

// Move moves the camera by the given offset.
func (c *Camera) Move(dx, dy, dz float64) {
	c.Position.X += dx
	c.Position.Y += dy
	c.Position.Z += dz
}

// GetPosition returns the camera position.
func (c *Camera) GetPosition() Vector3 {
	return c.Position
}

// GetTarget returns the camera target.
func (c *Camera) GetTarget() Vector3 {
	return c.Target
}

// GetUp returns the camera up vector.
func (c *Camera) GetUp() Vector3 {
	return c.Up
}

// SetMode sets the camera projection mode.
func (c *Camera) SetMode(mode CameraMode) {
	c.Mode = mode
}

// SetFOV sets the field of view in degrees (perspective only).
func (c *Camera) SetFOV(fovDegrees float64) {
	c.FOV = fovDegrees * math.Pi / 180.0
}

// SetAspect sets the aspect ratio.
func (c *Camera) SetAspect(aspect float64) {
	c.Aspect = aspect
}

// ============================================================
// NEW: Enhanced Scene (fog, background color, lighting)
// ============================================================

// Scene represents a 3D scene with objects, lights, fog, and background.
type Scene struct {
	objects         []Object
	lights          []*Light
	fog             Fog
	backgroundColor color.Color
	camera          *Camera
}

// NewScene creates a new 3D scene.
func NewScene() *Scene {
	return &Scene{
		objects:         make([]Object, 0),
		lights:          make([]*Light, 0),
		fog:             NewFog(),
		backgroundColor: color.RGBA{R: 0, G: 0, B: 0, A: 255},
	}
}

// AddObject adds an object to the scene.
func (s *Scene) AddObject(obj Object) {
	s.objects = append(s.objects, obj)
}

// AddObjects adds multiple objects to the scene at once.
func (s *Scene) AddObjects(objs ...Object) {
	s.objects = append(s.objects, objs...)
}

// GetObjects returns all objects in the scene.
func (s *Scene) GetObjects() []Object {
	return s.objects
}

// AddLight adds a light to the scene.
func (s *Scene) AddLight(light *Light) {
	s.lights = append(s.lights, light)
}

// GetLights returns all lights in the scene.
func (s *Scene) GetLights() []*Light {
	return s.lights
}

// SetCamera sets the scene camera.
func (s *Scene) SetCamera(camera *Camera) {
	s.camera = camera
}

// GetCamera returns the scene camera.
func (s *Scene) GetCamera() *Camera {
	return s.camera
}

// Add is an alias for AddObject.
func (s *Scene) Add(obj Object) {
	s.AddObject(obj)
}

// Remove removes an object from the scene by identity.
func (s *Scene) Remove(obj Object) {
	for i, o := range s.objects {
		if o == obj {
			s.objects = append(s.objects[:i], s.objects[i+1:]...)
			return
		}
	}
}

// SetFog sets the scene fog.
func (s *Scene) SetFog(fog Fog) {
	s.fog = fog
}

// GetFog returns the scene fog.
func (s *Scene) GetFog() Fog {
	return s.fog
}

// SetBackgroundColor sets the scene background color.
func (s *Scene) SetBackgroundColor(c color.Color) {
	s.backgroundColor = c
}

// GetBackgroundColor returns the scene background color.
func (s *Scene) GetBackgroundColor() color.Color {
	return s.backgroundColor
}

// ============================================================
// Object interface (existing, preserved)
// ============================================================

// Object represents a 3D object.
type Object interface {
	GetVertices() []Vector3
	GetIndices() []int
	GetColor() color.Color
}

// ============================================================
// NEW: Vector3 (existing, preserved)
// ============================================================

// Vector3 represents a 3D vector.
type Vector3 struct {
	X, Y, Z float64
}

// ============================================================
// NEW: PhysicsEngine interface and implementations
// ============================================================

// PhysicsRigidBodyDesc describes a rigid body to create.
type PhysicsRigidBodyDesc struct {
	Translation    Vector2
	LinearDamping  float64
	AngularDamping float64
}

// PhysicsColliderDesc describes a collider shape.
type PhysicsColliderDesc struct {
	Width       float64
	Height      float64
	Restitution float64
	Friction    float64
	Density     float64
}

// PhysicsRigidBody is a simulated rigid body.
type PhysicsRigidBody interface {
	ApplyImpulse(force Vector2)
	ApplyTorqueImpulse(torque float64)
	GetTranslation() Vector2
	GetRotation() float64
}

// PhysicsEngine is the physics simulation interface.
type PhysicsEngine interface {
	CreateRigidBody(desc PhysicsRigidBodyDesc) PhysicsRigidBody
	CreateCollider(desc PhysicsColliderDesc, body PhysicsRigidBody)
	RemoveRigidBody(body PhysicsRigidBody)
	Step(deltaTime float64)
	SetGravity(gx, gy float64)
	GetGravity() Vector2
}

type physicsBody struct {
	position    Vector2
	velocity    Vector2
	angle       float64
	angularVel  float64
	linearDamp  float64
	angularDamp float64
	active      bool
}

// PlanckPhysics is a simple 2D physics engine (Planck-like).
type PlanckPhysics struct {
	bodies  map[PhysicsRigidBody]*physicsBody
	gravity Vector2
}

// NewPlanckPhysics creates a new Planck-like physics world.
func NewPlanckPhysics(gravityX, gravityY float64) *PlanckPhysics {
	return &PlanckPhysics{
		bodies:  make(map[PhysicsRigidBody]*physicsBody),
		gravity: Vector2{X: gravityX, Y: gravityY},
	}
}

func (p *PlanckPhysics) CreateRigidBody(desc PhysicsRigidBodyDesc) PhysicsRigidBody {
	body := &physicsBody{
		position:    desc.Translation,
		linearDamp:  desc.LinearDamping,
		angularDamp: desc.AngularDamping,
		active:      true,
	}
	p.bodies[body] = body
	return body
}

func (p *PlanckPhysics) CreateCollider(desc PhysicsColliderDesc, body PhysicsRigidBody) {
	// Colliders are simple; associated implicitly with the body.
}

func (p *PlanckPhysics) RemoveRigidBody(body PhysicsRigidBody) {
	delete(p.bodies, body)
}

func (p *PlanckPhysics) Step(deltaTime float64) {
	for _, body := range p.bodies {
		if !body.active {
			continue
		}
		body.velocity.X += p.gravity.X * deltaTime
		body.velocity.Y += p.gravity.Y * deltaTime
		body.velocity.X *= (1.0 - body.linearDamp*deltaTime)
		body.velocity.Y *= (1.0 - body.linearDamp*deltaTime)
		body.angularVel *= (1.0 - body.angularDamp*deltaTime)
		body.position.X += body.velocity.X * deltaTime
		body.position.Y += body.velocity.Y * deltaTime
		body.angle += body.angularVel * deltaTime
	}
}

func (p *PlanckPhysics) SetGravity(gx, gy float64) {
	p.gravity = Vector2{X: gx, Y: gy}
}

func (p *PlanckPhysics) GetGravity() Vector2 {
	return p.gravity
}

func (b *physicsBody) ApplyImpulse(force Vector2) {
	b.velocity.X += force.X
	b.velocity.Y += force.Y
}

func (b *physicsBody) ApplyTorqueImpulse(torque float64) {
	b.angularVel += torque
}

func (b *physicsBody) GetTranslation() Vector2 {
	return b.position
}

func (b *physicsBody) GetRotation() float64 {
	return b.angle
}

// RapierPhysics is a simple 2D physics engine (Rapier-like stub implementation).
type RapierPhysics struct {
	bodies  map[PhysicsRigidBody]*physicsBody
	gravity Vector2
}

// NewRapierPhysics creates a new Rapier-like physics world.
func NewRapierPhysics(gravityX, gravityY float64) *RapierPhysics {
	return &RapierPhysics{
		bodies:  make(map[PhysicsRigidBody]*physicsBody),
		gravity: Vector2{X: gravityX, Y: gravityY},
	}
}

func (r *RapierPhysics) CreateRigidBody(desc PhysicsRigidBodyDesc) PhysicsRigidBody {
	body := &physicsBody{
		position:    desc.Translation,
		linearDamp:  desc.LinearDamping,
		angularDamp: desc.AngularDamping,
		active:      true,
	}
	r.bodies[body] = body
	return body
}

func (r *RapierPhysics) CreateCollider(desc PhysicsColliderDesc, body PhysicsRigidBody) {
}

func (r *RapierPhysics) RemoveRigidBody(body PhysicsRigidBody) {
	delete(r.bodies, body)
}

func (r *RapierPhysics) Step(deltaTime float64) {
	for _, body := range r.bodies {
		if !body.active {
			continue
		}
		body.velocity.X += r.gravity.X * deltaTime
		body.velocity.Y += r.gravity.Y * deltaTime
		body.velocity.X *= (1.0 - body.linearDamp*deltaTime)
		body.velocity.Y *= (1.0 - body.linearDamp*deltaTime)
		body.angularVel *= (1.0 - body.angularDamp*deltaTime)
		body.position.X += body.velocity.X * deltaTime
		body.position.Y += body.velocity.Y * deltaTime
		body.angle += body.angularVel * deltaTime
	}
}

func (r *RapierPhysics) SetGravity(gx, gy float64) {
	r.gravity = Vector2{X: gx, Y: gy}
}

func (r *RapierPhysics) GetGravity() Vector2 {
	return r.gravity
}

// ============================================================
// NEW: ParticleSystem
// ============================================================

// Particle is a single simulated particle.
type Particle struct {
	Position        Vector3
	Velocity        Vector3
	Acceleration    Vector3
	Life            float64
	MaxLife         float64
	Size            float64
	Color           color.Color
	Rotation        float64
	AngularVelocity float64
	Active          bool
}

// ParticleSystem manages a pool of particles for effects (explosions, etc.).
type ParticleSystem struct {
	ID           string
	Particles    []*Particle
	Emitter      Vector3
	Gravity      Vector3
	EmissionRate float64
	Accumulator  float64
	MaxParticles int
	Texture      *Texture
	Enabled      bool
}

// NewParticleSystem creates a new particle system.
func NewParticleSystem(maxParticles int) *ParticleSystem {
	return &ParticleSystem{
		Particles:    make([]*Particle, 0, maxParticles),
		MaxParticles: maxParticles,
		Gravity:      Vector3{X: 0, Y: -9.8, Z: 0},
		Emitter:      Vector3{X: 0, Y: 0, Z: 0},
		Enabled:      true,
	}
}

// Emit spawns count new particles.
func (ps *ParticleSystem) Emit(count int) {
	if !ps.Enabled {
		return
	}
	for i := 0; i < count; i++ {
		var p *Particle
		for _, existing := range ps.Particles {
			if !existing.Active {
				p = existing
				break
			}
		}
		if p == nil && len(ps.Particles) < ps.MaxParticles {
			p = &Particle{}
			ps.Particles = append(ps.Particles, p)
		}
		if p == nil {
			continue
		}
		p.Position = ps.Emitter
		p.Velocity = Vector3{
			X: (rand.Float64() - 0.5) * 2,
			Y: rand.Float64()*2 + 1,
			Z: (rand.Float64() - 0.5) * 2,
		}
		p.Acceleration = ps.Gravity
		p.Life = 1.0 + rand.Float64()
		p.MaxLife = p.Life
		p.Size = 0.5 + rand.Float64()
		p.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		p.Rotation = 0
		p.AngularVelocity = (rand.Float64() - 0.5) * 4
		p.Active = true
	}
}

// Update advances the particle simulation by deltaTime seconds.
func (ps *ParticleSystem) Update(deltaTime float64) {
	if !ps.Enabled {
		return
	}
	if ps.EmissionRate > 0 {
		ps.Accumulator += deltaTime
		toEmit := int(ps.Accumulator * ps.EmissionRate)
		if toEmit > 0 {
			ps.Accumulator -= float64(toEmit) / ps.EmissionRate
			ps.Emit(toEmit)
		}
	}
	for _, p := range ps.Particles {
		if !p.Active {
			continue
		}
		p.Velocity.X += p.Acceleration.X * deltaTime
		p.Velocity.Y += p.Acceleration.Y * deltaTime
		p.Velocity.Z += p.Acceleration.Z * deltaTime
		p.Position.X += p.Velocity.X * deltaTime
		p.Position.Y += p.Velocity.Y * deltaTime
		p.Position.Z += p.Velocity.Z * deltaTime
		p.Rotation += p.AngularVelocity * deltaTime
		p.Life -= deltaTime
		if p.Life <= 0 {
			p.Active = false
		}
	}
}

// Render renders all active particles to the buffer via the renderer.
func (ps *ParticleSystem) Render(r *Renderer, buffer *core.NativeBuffer) {
	if !ps.Enabled {
		return
	}
	for _, p := range ps.Particles {
		if !p.Active {
			continue
		}
		alpha := p.Life / p.MaxLife
		if alpha < 0 {
			alpha = 0
		}
		if alpha > 1 {
			alpha = 1
		}
		rCol, gCol, bCol, _ := p.Color.RGBA()
		c := color.RGBA{
			R: uint8(rCol >> 8),
			G: uint8(gCol >> 8),
			B: uint8(bCol >> 8),
			A: uint8(float64(255) * alpha),
		}
		x, y := r.projectWithCamera(p.Position.X, p.Position.Y, p.Position.Z)
		if x >= 0 && x < buffer.Width() && y >= 0 && y < buffer.Height() {
			buffer.SetCell(x, y, '#', uint16(c.R), uint16(c.G), 0)
		}
	}
}

// ============================================================
// NEW: SpriteRenderer
// ============================================================

// SpriteRenderer renders sprites to the terminal buffer.
type SpriteRenderer struct {
	renderer *Renderer
	sprites  []*Sprite
}

// NewSpriteRenderer creates a new sprite renderer.
func NewSpriteRenderer(r *Renderer) *SpriteRenderer {
	return &SpriteRenderer{
		renderer: r,
		sprites:  make([]*Sprite, 0),
	}
}

// AddSprite adds a sprite to be rendered.
func (sr *SpriteRenderer) AddSprite(s *Sprite) {
	sr.sprites = append(sr.sprites, s)
}

// RemoveSprite removes a sprite from the renderer.
func (sr *SpriteRenderer) RemoveSprite(s *Sprite) {
	for i, s2 := range sr.sprites {
		if s2 == s {
			sr.sprites = append(sr.sprites[:i], sr.sprites[i+1:]...)
			break
		}
	}
}

// Render renders all sprites to the buffer.
func (sr *SpriteRenderer) Render(buffer *core.NativeBuffer) {
	for _, s := range sr.sprites {
		if !s.Visible || s.Texture == nil {
			continue
		}
		sr.renderer.RenderSprite(s, buffer)
	}
}

var _ = fmt.Sprintf
var _ = math.Pi

// ============================================================
// THREE.MATH: Color, Quaternion, Matrix4, Euler
// ============================================================

// Color represents an RGBA color with Three.js-like helpers.
type Color struct {
	R, G, B, A float64 // 0-1
}

// NewColor creates a Color from RGBA values (0-1).
func NewColor(r, g, b, a float64) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// NewColorFromRGBA creates a Color from an image/color.RGBA (0-255).
func NewColorFromRGBA(c color.RGBA) Color {
	return Color{
		R: float64(c.R) / 255.0,
		G: float64(c.G) / 255.0,
		B: float64(c.B) / 255.0,
		A: float64(c.A) / 255.0,
	}
}

// SetHex sets the color from a hexadecimal value (e.g. 0xff0000).
func (c *Color) SetHex(hex uint32) {
	c.R = float64((hex>>16)&0xFF) / 255.0
	c.G = float64((hex>>8)&0xFF) / 255.0
	c.B = float64(hex&0xFF) / 255.0
	c.A = 1.0
}

// ToRGBA converts to image/color.RGBA (0-255).
func (c Color) ToRGBA() color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, math.Max(0, c.R*255))),
		G: uint8(math.Min(255, math.Max(0, c.G*255))),
		B: uint8(math.Min(255, math.Max(0, c.B*255))),
		A: uint8(math.Min(255, math.Max(0, c.A*255))),
	}
}

// Clone returns a copy of the color.
func (c Color) Clone() Color {
	return Color{R: c.R, G: c.G, B: c.B, A: c.A}
}

// Quaternion represents a 3D rotation.
type Quaternion struct {
	X, Y, Z, W float64
}

// NewQuaternion creates a new quaternion.
func NewQuaternion(x, y, z, w float64) Quaternion {
	return Quaternion{X: x, Y: y, Z: z, W: w}
}

// IdentityQuaternion returns the identity quaternion.
func IdentityQuaternion() Quaternion {
	return Quaternion{X: 0, Y: 0, Z: 0, W: 1}
}

// SetFromAxisAngle sets the quaternion from a rotation axis and angle (radians).
func (q *Quaternion) SetFromAxisAngle(axis Vector3, angle float64) {
	halfAngle := angle / 2
	s := math.Sin(halfAngle)
	q.X = axis.X * s
	q.Y = axis.Y * s
	q.Z = axis.Z * s
	q.W = math.Cos(halfAngle)
	q.normalize()
}

// Multiply multiplies two quaternions: q = q * other.
func (q *Quaternion) Multiply(other Quaternion) {
	qx, qy, qz, qw := q.X, q.Y, q.Z, q.W
	ox, oy, oz, ow := other.X, other.Y, other.Z, other.W
	q.X = qx*ow + qw*ox + qy*oz - qz*oy
	q.Y = qy*ow + qw*oy + qz*ox - qx*oz
	q.Z = qz*ow + qw*oz + qx*oy - qy*ox
	q.W = qw*ow - qx*ox - qy*oy - qz*oz
}

// normalize normalizes the quaternion.
func (q *Quaternion) normalize() {
	len := math.Sqrt(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)
	if len > 0 {
		q.X /= len
		q.Y /= len
		q.Z /= len
		q.W /= len
	}
}

// Clone returns a copy of the quaternion.
func (q Quaternion) Clone() Quaternion {
	return Quaternion{X: q.X, Y: q.Y, Z: q.Z, W: q.W}
}

// Matrix4 represents a 4x4 transformation matrix.
type Matrix4 struct {
	Elements [16]float64
}

// NewMatrix4 creates an identity matrix.
func NewMatrix4() Matrix4 {
	return Matrix4{
		Elements: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
	}
}

// MakeTranslation creates a translation matrix.
func MakeTranslation(x, y, z float64) Matrix4 {
	m := NewMatrix4()
	m.Elements[12] = x
	m.Elements[13] = y
	m.Elements[14] = z
	return m
}

// MakeRotationFromQuaternion creates a rotation matrix from a quaternion.
func MakeRotationFromQuaternion(q Quaternion) Matrix4 {
	x, y, z, w := q.X, q.Y, q.Z, q.W
	x2 := x + x
	y2 := y + y
	z2 := z + z
	xx := x * x2
	xy := x * y2
	xz := x * z2
	yy := y * y2
	yz := y * z2
	zz := z * z2
	wx := w * x2
	wy := w * y2
	wz := w * z2

	m := NewMatrix4()
	m.Elements[0] = 1 - (yy + zz)
	m.Elements[1] = xy + wz
	m.Elements[2] = xz - wy
	m.Elements[3] = 0

	m.Elements[4] = xy - wz
	m.Elements[5] = 1 - (xx + zz)
	m.Elements[6] = yz + wx
	m.Elements[7] = 0

	m.Elements[8] = xz + wy
	m.Elements[9] = yz - wx
	m.Elements[10] = 1 - (xx + yy)
	m.Elements[11] = 0

	m.Elements[12] = 0
	m.Elements[13] = 0
	m.Elements[14] = 0
	m.Elements[15] = 1
	return m
}

// MakeScale creates a scale matrix.
func MakeScale(x, y, z float64) Matrix4 {
	m := NewMatrix4()
	m.Elements[0] = x
	m.Elements[5] = y
	m.Elements[10] = z
	return m
}

// Compose creates a matrix from position, quaternion, and scale.
func Compose(position Vector3, quat Quaternion, scale Vector3) Matrix4 {
	m := MakeRotationFromQuaternion(quat)
	m.Elements[0] *= scale.X
	m.Elements[1] *= scale.X
	m.Elements[2] *= scale.X
	m.Elements[4] *= scale.Y
	m.Elements[5] *= scale.Y
	m.Elements[6] *= scale.Y
	m.Elements[8] *= scale.Z
	m.Elements[9] *= scale.Z
	m.Elements[10] *= scale.Z
	m.Elements[12] = position.X
	m.Elements[13] = position.Y
	m.Elements[14] = position.Z
	return m
}

// Decompose extracts position, quaternion, and scale from a matrix.
func (m *Matrix4) Decompose() (position Vector3, quat Quaternion, scale Vector3) {
	position = Vector3{
		X: m.Elements[12],
		Y: m.Elements[13],
		Z: m.Elements[14],
	}

	scaleX := math.Sqrt(m.Elements[0]*m.Elements[0] + m.Elements[1]*m.Elements[1] + m.Elements[2]*m.Elements[2])
	scaleY := math.Sqrt(m.Elements[4]*m.Elements[4] + m.Elements[5]*m.Elements[5] + m.Elements[6]*m.Elements[6])
	scaleZ := math.Sqrt(m.Elements[8]*m.Elements[8] + m.Elements[9]*m.Elements[9] + m.Elements[10]*m.Elements[10])
	scale = Vector3{X: scaleX, Y: scaleY, Z: scaleZ}

	// Simple quaternion extraction from rotation part
	// (assumes no shear)
	rot := NewMatrix4()
	rot.Elements[0] = m.Elements[0] / scaleX
	rot.Elements[1] = m.Elements[1] / scaleX
	rot.Elements[2] = m.Elements[2] / scaleX
	rot.Elements[4] = m.Elements[4] / scaleY
	rot.Elements[5] = m.Elements[5] / scaleY
	rot.Elements[6] = m.Elements[6] / scaleY
	rot.Elements[8] = m.Elements[8] / scaleZ
	rot.Elements[9] = m.Elements[9] / scaleZ
	rot.Elements[10] = m.Elements[10] / scaleZ

	trace := rot.Elements[0] + rot.Elements[5] + rot.Elements[10]
	if trace > 0 {
		s := math.Sqrt(trace+1.0) * 2
		quat = Quaternion{
			X: (rot.Elements[6] - rot.Elements[9]) / s,
			Y: (rot.Elements[8] - rot.Elements[2]) / s,
			Z: (rot.Elements[1] - rot.Elements[4]) / s,
			W: 0.25 * s,
		}
	} else if rot.Elements[0] > rot.Elements[5] && rot.Elements[0] > rot.Elements[10] {
		s := math.Sqrt(1.0+rot.Elements[0]-rot.Elements[5]-rot.Elements[10]) * 2
		quat = Quaternion{
			X: 0.25 * s,
			Y: (rot.Elements[1] + rot.Elements[4]) / s,
			Z: (rot.Elements[8] + rot.Elements[2]) / s,
			W: (rot.Elements[6] - rot.Elements[9]) / s,
		}
	} else if rot.Elements[5] > rot.Elements[10] {
		s := math.Sqrt(1.0+rot.Elements[5]-rot.Elements[0]-rot.Elements[10]) * 2
		quat = Quaternion{
			X: (rot.Elements[1] + rot.Elements[4]) / s,
			Y: 0.25 * s,
			Z: (rot.Elements[6] + rot.Elements[9]) / s,
			W: (rot.Elements[8] - rot.Elements[2]) / s,
		}
	} else {
		s := math.Sqrt(1.0+rot.Elements[10]-rot.Elements[0]-rot.Elements[5]) * 2
		quat = Quaternion{
			X: (rot.Elements[8] + rot.Elements[2]) / s,
			Y: (rot.Elements[6] + rot.Elements[9]) / s,
			Z: 0.25 * s,
			W: (rot.Elements[1] - rot.Elements[4]) / s,
		}
	}
	return position, quat, scale
}

// Multiply multiplies two matrices: m = m * other.
func (m *Matrix4) Multiply(other Matrix4) {
	var result [16]float64
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i*4+j] = m.Elements[i*4+0]*other.Elements[0*4+j] +
				m.Elements[i*4+1]*other.Elements[1*4+j] +
				m.Elements[i*4+2]*other.Elements[2*4+j] +
				m.Elements[i*4+3]*other.Elements[3*4+j]
		}
	}
	m.Elements = result
}

// Clone returns a copy of the matrix.
func (m Matrix4) Clone() Matrix4 {
	return m
}

// Euler represents Euler angles for rotation.
type Euler struct {
	X, Y, Z float64 // radians
	Order   string  // rotation order, default "XYZ"
}

// NewEuler creates a new Euler.
func NewEuler(x, y, z float64, order ...string) Euler {
	o := "XYZ"
	if len(order) > 0 {
		o = order[0]
	}
	return Euler{X: x, Y: y, Z: z, Order: o}
}

// ToQuaternion converts Euler angles to a Quaternion.
func (e Euler) ToQuaternion() Quaternion {
	// XYZ order
	cx := math.Cos(e.X / 2)
	sx := math.Sin(e.X / 2)
	cy := math.Cos(e.Y / 2)
	sy := math.Sin(e.Y / 2)
	cz := math.Cos(e.Z / 2)
	sz := math.Sin(e.Z / 2)

	return Quaternion{
		X: sx*cy*cz - cx*sy*sz,
		Y: cx*sy*cz + sx*cy*sz,
		Z: cx*cy*sz - sx*sy*cz,
		W: cx*cy*cz + sx*sy*sz,
	}
}

// ============================================================
// BUILT-IN GEOMETRIES (Box, Plane, Sphere)
// ============================================================

// BoxGeometry creates a box (cube) geometry.
func BoxGeometry(width, height, depth float64) *Geometry {
	w2 := width / 2
	h2 := height / 2
	d2 := depth / 2

	vertices := []Vector3{
		// Front face
		{-w2, -h2, d2}, {w2, -h2, d2}, {w2, h2, d2}, {-w2, h2, d2},
		// Back face
		{-w2, -h2, -d2}, {-w2, h2, -d2}, {w2, h2, -d2}, {w2, -h2, -d2},
		// Left face
		{-w2, -h2, -d2}, {-w2, -h2, d2}, {-w2, h2, d2}, {-w2, h2, -d2},
		// Right face
		{w2, -h2, d2}, {w2, -h2, -d2}, {w2, h2, -d2}, {w2, h2, d2},
		// Top face
		{-w2, h2, -d2}, {-w2, h2, d2}, {w2, h2, d2}, {w2, h2, -d2},
		// Bottom face
		{-w2, -h2, d2}, {-w2, -h2, -d2}, {w2, -h2, -d2}, {w2, -h2, d2},
	}

	indices := []int{}
	for face := 0; face < 6; face++ {
		offset := face * 4
		indices = append(indices, offset, offset+1, offset+2)
		indices = append(indices, offset, offset+2, offset+3)
	}

	return &Geometry{
		ID:       "box",
		Vertices: vertices,
		Indices:  indices,
	}
}

// PlaneGeometry creates a plane geometry.
func PlaneGeometry(width, height float64) *Geometry {
	w2 := width / 2
	h2 := height / 2

	vertices := []Vector3{
		{-w2, h2, 0}, {w2, h2, 0}, {w2, -h2, 0}, {-w2, -h2, 0},
	}
	indices := []int{0, 1, 2, 0, 2, 3}

	return &Geometry{
		ID:       "plane",
		Vertices: vertices,
		Indices:  indices,
	}
}

// SphereGeometry creates a sphere geometry.
func SphereGeometry(radius float64, widthSegments, heightSegments int) *Geometry {
	vertices := make([]Vector3, 0)
	indices := make([]int, 0)

	for i := 0; i <= heightSegments; i++ {
		v := float64(i) / float64(heightSegments)
		phi := v * math.Pi
		for j := 0; j <= widthSegments; j++ {
			u := float64(j) / float64(widthSegments)
			theta := u * 2 * math.Pi
			x := radius * math.Sin(phi) * math.Cos(theta)
			y := radius * math.Cos(phi)
			z := radius * math.Sin(phi) * math.Sin(theta)
			vertices = append(vertices, Vector3{X: x, Y: y, Z: z})
		}
	}

	for i := 0; i < heightSegments; i++ {
		for j := 0; j < widthSegments; j++ {
			a := i*(widthSegments+1) + j
			b := i*(widthSegments+1) + j + 1
			c := (i+1)*(widthSegments+1) + j + 1
			d := (i+1)*(widthSegments+1) + j
			indices = append(indices, a, b, d)
			indices = append(indices, b, c, d)
		}
	}

	return &Geometry{
		ID:       "sphere",
		Vertices: vertices,
		Indices:  indices,
	}
}

// ============================================================
// BUILT-IN GEOMETRIES (Cylinder, Cone, Torus, Ring)
// ============================================================

// CylinderGeometry creates a cylinder geometry.
func CylinderGeometry(radiusTop, radiusBottom, height float64, radialSegments int) *Geometry {
	if radialSegments < 3 {
		radialSegments = 8
	}
	vertices := make([]Vector3, 0)
	indices := make([]int, 0)

	h2 := height / 2
	// Top cap
	for j := 0; j <= radialSegments; j++ {
		theta := float64(j) / float64(radialSegments) * 2 * math.Pi
		x := radiusTop * math.Cos(theta)
		z := radiusTop * math.Sin(theta)
		vertices = append(vertices, Vector3{X: x, Y: h2, Z: z})
	}
	topCenter := len(vertices)
	vertices = append(vertices, Vector3{X: 0, Y: h2, Z: 0})
	// Bottom cap
	for j := 0; j <= radialSegments; j++ {
		theta := float64(j) / float64(radialSegments) * 2 * math.Pi
		x := radiusBottom * math.Cos(theta)
		z := radiusBottom * math.Sin(theta)
		vertices = append(vertices, Vector3{X: x, Y: -h2, Z: z})
	}
	bottomCenter := len(vertices)
	vertices = append(vertices, Vector3{X: 0, Y: -h2, Z: 0})

	// Side quads
	for j := 0; j < radialSegments; j++ {
		a := j
		b := j + 1
		c := radialSegments + 1 + j + 1
		d := radialSegments + 1 + j
		indices = append(indices, a, b, c)
		indices = append(indices, a, c, d)
	}
	// Top cap fan
	for j := 0; j < radialSegments; j++ {
		a := topCenter
		b := j
		c := j + 1
		indices = append(indices, a, c, b)
	}
	// Bottom cap fan
	for j := 0; j < radialSegments; j++ {
		a := bottomCenter
		b := radialSegments + 1 + j + 1
		c := radialSegments + 1 + j
		indices = append(indices, a, c, b)
	}

	return &Geometry{
		ID:       "cylinder",
		Vertices: vertices,
		Indices:  indices,
	}
}

// ConeGeometry creates a cone geometry.
func ConeGeometry(radius, height float64, radialSegments int) *Geometry {
	return CylinderGeometry(radius, 0, height, radialSegments)
}

// TorusGeometry creates a torus (donut) geometry.
func TorusGeometry(radius, tube float64, radialSegments, tubularSegments int) *Geometry {
	if radialSegments < 3 {
		radialSegments = 16
	}
	if tubularSegments < 3 {
		tubularSegments = 16
	}
	vertices := make([]Vector3, 0)
	indices := make([]int, 0)

	for i := 0; i <= radialSegments; i++ {
		u := float64(i) / float64(radialSegments) * 2 * math.Pi
		for j := 0; j <= tubularSegments; j++ {
			v := float64(j) / float64(tubularSegments) * 2 * math.Pi
			cu, su := math.Cos(u), math.Sin(u)
			cv, sv := math.Cos(v), math.Sin(v)
			x := (radius + tube*cv) * cu
			y := tube * sv
			z := (radius + tube*cv) * su
			vertices = append(vertices, Vector3{X: x, Y: y, Z: z})
		}
	}

	stride := tubularSegments + 1
	for i := 0; i < radialSegments; i++ {
		for j := 0; j < tubularSegments; j++ {
			a := i*stride + j
			b := i*stride + j + 1
			c := (i+1)*stride + j + 1
			d := (i+1)*stride + j
			indices = append(indices, a, b, d)
			indices = append(indices, b, c, d)
		}
	}

	return &Geometry{
		ID:       "torus",
		Vertices: vertices,
		Indices:  indices,
	}
}

// RingGeometry creates a flat ring (annulus) geometry.
func RingGeometry(innerRadius, outerRadius float64, thetaSegments int) *Geometry {
	if thetaSegments < 3 {
		thetaSegments = 16
	}
	vertices := make([]Vector3, 0)
	indices := make([]int, 0)

	for i := 0; i <= thetaSegments; i++ {
		theta := float64(i) / float64(thetaSegments) * 2 * math.Pi
		// inner vertex
		vertices = append(vertices, Vector3{X: innerRadius * math.Cos(theta), Y: innerRadius * math.Sin(theta), Z: 0})
		// outer vertex
		vertices = append(vertices, Vector3{X: outerRadius * math.Cos(theta), Y: outerRadius * math.Sin(theta), Z: 0})
	}

	stride := 2
	for i := 0; i < thetaSegments; i++ {
		a := i * stride
		b := i*stride + 1
		c := (i+1)*stride + 1
		d := (i + 1) * stride
		indices = append(indices, a, d, c)
		indices = append(indices, a, c, b)
	}

	return &Geometry{
		ID:       "ring",
		Vertices: vertices,
		Indices:  indices,
	}
}

// ============================================================
// MATERIAL TYPES (MeshBasic, MeshLambert, MeshPhong)
// ============================================================

// MaterialType identifies the shading model.
type MaterialType int

const (
	MaterialBasic   MaterialType = iota // Unlit
	MaterialLambert                     // Diffuse lighting
	MaterialPhong                       // Diffuse + specular
)

// MaterialTypeName returns the string name of a MaterialType.
func MaterialTypeName(mt MaterialType) string {
	switch mt {
	case MaterialBasic:
		return "basic"
	case MaterialLambert:
		return "lambert"
	case MaterialPhong:
		return "phong"
	default:
		return "unknown"
	}
}

// EnhancedMaterial extends Material with Three.js-like material types.
type EnhancedMaterial struct {
	*Material
	Type         MaterialType
	Specular     color.Color
	Shininess    float64
	Side         int // 0=front, 1=back, 2=double
	Wireframe    bool
	VertexColors bool
}

// NewEnhancedMaterial creates a new enhanced material.
func NewEnhancedMaterial(shader *Shader, mt MaterialType) *EnhancedMaterial {
	return &EnhancedMaterial{
		Material:  NewMaterial(shader),
		Type:      mt,
		Specular:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Shininess: 30,
		Side:      0,
		Wireframe: false,
	}
}

// NewMeshBasicMaterial creates an unlit material.
func NewMeshBasicMaterial(shader *Shader) *EnhancedMaterial {
	return NewEnhancedMaterial(shader, MaterialBasic)
}

// NewMeshLambertMaterial creates a diffuse material.
func NewMeshLambertMaterial(shader *Shader) *EnhancedMaterial {
	return NewEnhancedMaterial(shader, MaterialLambert)
}

// NewMeshPhongMaterial creates a diffuse+specular material.
func NewMeshPhongMaterial(shader *Shader) *EnhancedMaterial {
	return NewEnhancedMaterial(shader, MaterialPhong)
}

// ============================================================
// INSTANCED MESH (for particles / sprites)
// ============================================================

// InstancedMesh renders multiple instances of the same geometry.
type InstancedMesh struct {
	ID       string
	Geometry *Geometry
	Material *Material
	Matrices []Matrix4
	Visible  bool
	Count    int
}

// NewInstancedMesh creates a new instanced mesh.
func NewInstancedMesh(geometry *Geometry, material *Material, count int) *InstancedMesh {
	return &InstancedMesh{
		ID:       "",
		Geometry: geometry,
		Material: material,
		Matrices: make([]Matrix4, count),
		Visible:  true,
		Count:    count,
	}
}

// SetMatrixAt sets the transformation matrix for an instance.
func (im *InstancedMesh) SetMatrixAt(index int, m Matrix4) {
	if index >= 0 && index < len(im.Matrices) {
		im.Matrices[index] = m
	}
}

// GetMatrixAt returns the transformation matrix for an instance.
func (im *InstancedMesh) GetMatrixAt(index int) Matrix4 {
	if index >= 0 && index < len(im.Matrices) {
		return im.Matrices[index]
	}
	return NewMatrix4()
}

// GetVertices returns the mesh vertices (from geometry).
func (im *InstancedMesh) GetVertices() []Vector3 {
	if im.Geometry != nil {
		return im.Geometry.Vertices
	}
	return nil
}

// GetIndices returns the mesh indices (from geometry).
func (im *InstancedMesh) GetIndices() []int {
	if im.Geometry != nil {
		return im.Geometry.Indices
	}
	return nil
}

// GetColor returns the mesh color.
func (im *InstancedMesh) GetColor() color.Color {
	if im.Material != nil && im.Material.Color != nil {
		return im.Material.Color
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

// ============================================================
// ENHANCED RENDERER: directional/point lighting, depth, culling
// ============================================================

// FaceCulling determines which faces are culled.
type FaceCulling int

const (
	CullNone FaceCulling = iota
	CullFront
	CullBack
)

// DepthWrite controls whether depth is written.
type DepthWrite bool

const (
	DepthWriteEnabled  DepthWrite = true
	DepthWriteDisabled DepthWrite = false
)

// renderObjectEnhanced renders an object with full lighting and material support.
func (r *Renderer) renderObjectEnhanced(obj Object, buffer *core.NativeBuffer, lights []*Light) {
	vertices := r.transformedVertices(obj)
	indices := obj.GetIndices()
	if len(vertices) == 0 || len(indices) == 0 {
		return
	}

	material := r.material
	objColor := obj.GetColor()
	if material != nil && material.Color != nil {
		objColor = material.Color
	}

	// Compute lighting
	for i := 0; i < len(indices); i += 3 {
		if i+2 >= len(indices) {
			break
		}
		v0 := vertices[indices[i]]
		v1 := vertices[indices[i+1]]
		v2 := vertices[indices[i+2]]

		// Compute normal
		nx, ny, nz := computeTriangleNormal(v0, v1, v2)

		// Project 3D to 2D using camera
		x0, y0 := r.projectWithCamera(v0.X, v0.Y, v0.Z)
		x1, y1 := r.projectWithCamera(v1.X, v1.Y, v1.Z)
		x2, y2 := r.projectWithCamera(v2.X, v2.Y, v2.Z)

		// Compute lighting at triangle center
		cx := (v0.X + v1.X + v2.X) / 3
		cy := (v0.Y + v1.Y + v2.Y) / 3
		cz := (v0.Z + v1.Z + v2.Z) / 3

		finalColor := r.computeLighting(objColor, nx, ny, nz, cx, cy, cz, lights)

		r.drawLine(x0, y0, x1, y1, buffer, finalColor)
		r.drawLine(x1, y1, x2, y2, buffer, finalColor)
		r.drawLine(x2, y2, x0, y0, buffer, finalColor)
	}
}

// transformedVertices returns the object's vertices transformed by its model matrix
// (position, rotation, scale) when the object is a Mesh. Otherwise returns the raw vertices.
func (r *Renderer) transformedVertices(obj Object) []Vector3 {
	vertices := obj.GetVertices()
	if len(vertices) == 0 {
		return vertices
	}
	mesh, ok := obj.(*Mesh)
	if !ok {
		return vertices
	}
	m := Compose(mesh.Position, mesh.RotationQuat, mesh.Scale)
	result := make([]Vector3, len(vertices))
	for i, v := range vertices {
		result[i] = multiplyMatrixVector(m, v)
	}
	return result
}

// multiplyMatrixVector transforms a 3D vector by a 4x4 matrix (treating w=1).
func multiplyMatrixVector(m Matrix4, v Vector3) Vector3 {
	e := m.Elements
	return Vector3{
		X: e[0]*v.X + e[4]*v.Y + e[8]*v.Z + e[12],
		Y: e[1]*v.X + e[5]*v.Y + e[9]*v.Z + e[13],
		Z: e[2]*v.X + e[6]*v.Y + e[10]*v.Z + e[14],
	}
}

// computeLighting calculates the final color from all lights.
func (r *Renderer) computeLighting(baseColor color.Color, nx, ny, nz, cx, cy, cz float64, lights []*Light) color.Color {
	// Convert base to 0-1
	br, bg, bb, _ := baseColor.RGBA()
	rf := float64(br>>8) / 255.0
	gf := float64(bg>>8) / 255.0
	bf := float64(bb>>8) / 255.0

	// Ambient contribution
	ar, ag, ab := rf, gf, bf

	for _, l := range lights {
		switch l.Type {
		case LightAmbient:
			lr, _, _, _ := l.Color.RGBA()
			li := float64(lr>>8) / 255.0 * l.Intensity
			ar += rf * li
			ag += gf * li
			ab += bf * li
		case LightDirectional:
			lr, _, _, _ := l.Color.RGBA()
			li := float64(lr>>8) / 255.0 * l.Intensity
			// Diffuse: max(0, N·L)
			dl := nx*l.Direction.X + ny*l.Direction.Y + nz*l.Direction.Z
			if dl < 0 {
				dl = 0
			}
			ar += rf * li * dl
			ag += gf * li * dl
			ab += bf * li * dl
		case LightPoint:
			lr, _, _, _ := l.Color.RGBA()
			li := float64(lr>>8) / 255.0 * l.Intensity
			// Direction from point to light
			dx := l.Position.X - cx
			dy := l.Position.Y - cy
			dz := l.Position.Z - cz
			dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
			if dist == 0 {
				dist = 0.001
			}
			// Attenuation
			att := 1.0 / (1.0 + 0.01*dist + 0.001*dist*dist)
			if l.Range > 0 && dist > l.Range {
				att = 0
			}
			// Diffuse
			dl := (dx*nx + dy*ny + dz*nz) / dist
			if dl < 0 {
				dl = 0
			}
			ar += rf * li * dl * att
			ag += gf * li * dl * att
			ab += bf * li * dl * att
		}
	}

	// Clamp
	clamp := func(v float64) float64 {
		if v > 1 {
			return 1
		}
		if v < 0 {
			return 0
		}
		return v
	}

	return color.RGBA{
		R: uint8(clamp(ar) * 255),
		G: uint8(clamp(ag) * 255),
		B: uint8(clamp(ab) * 255),
		A: 255,
	}
}

// computeTriangleNormal computes the normal vector of a triangle.
func computeTriangleNormal(v0, v1, v2 Vector3) (float64, float64, float64) {
	ax := v1.X - v0.X
	ay := v1.Y - v0.Y
	az := v1.Z - v0.Z
	bx := v2.X - v0.X
	by := v2.Y - v0.Y
	bz := v2.Z - v0.Z

	nx := ay*bz - az*by
	ny := az*bx - ax*bz
	nz := ax*by - ay*bx

	len := math.Sqrt(nx*nx + ny*ny + nz*nz)
	if len > 0 {
		return nx / len, ny / len, nz / len
	}
	return 0, 1, 0
}

// ============================================================
// SPRITE ANIMATION: SpriteAnimator, TiledSprite (ported from TypeScript)
// ============================================================

// SpriteSheetProperties describes a sprite sheet layout.
type SpriteSheetProperties struct {
	ImagePath     string
	TilesetWidth  int
	TilesetHeight int
	NumFrames     int
	UVTileSize    Vector2
}

// SpriteResource holds a texture and its sheet properties.
type SpriteResource struct {
	ID              string
	Texture         *Texture
	SheetProperties SpriteSheetProperties
}

// NewSpriteResource creates a new sprite resource.
func NewSpriteResource(id string, texture *Texture, tilesetWidth, tilesetHeight, numFrames int) *SpriteResource {
	uvTileWidth := 1.0
	uvTileHeight := 1.0
	if tilesetWidth > 0 && numFrames > 0 {
		uvTileWidth = 1.0 / float64(numFrames)
	}
	return &SpriteResource{
		ID:      id,
		Texture: texture,
		SheetProperties: SpriteSheetProperties{
			ImagePath:     id,
			TilesetWidth:  tilesetWidth,
			TilesetHeight: tilesetHeight,
			NumFrames:     numFrames,
			UVTileSize:    Vector2{X: uvTileWidth, Y: uvTileHeight},
		},
	}
}

// ResolvedAnimationState holds resolved animation configuration.
type ResolvedAnimationState struct {
	ImagePath       string
	TilesetWidth    int
	TilesetHeight   int
	NumFrames       int
	AnimNumFrames   int
	AnimFrameOffset int
	FrameDuration   float64
	Loop            bool
	InitialFrame    int
	FlipX           bool
	FlipY           bool
	Texture         *Texture
}

// DefaultAnimationState returns default values for an animation state.
func DefaultAnimationState() ResolvedAnimationState {
	return ResolvedAnimationState{
		AnimNumFrames:   1,
		AnimFrameOffset: 0,
		FrameDuration:   100,
		Loop:            true,
		InitialFrame:    0,
		FlipX:           false,
		FlipY:           false,
	}
}

// SpriteAnimation represents a single sprite sheet animation.
type SpriteAnimation struct {
	Name           string
	State          ResolvedAnimationState
	resource       *SpriteResource
	instanceIndex  int
	frameAttribute []float32
	flipAttribute  []float32

	CurrentLocalFrame int
	TimeAccumulator   float64
	IsPlaying         bool
	IsActive          bool
}

// NewSpriteAnimation creates a new sprite animation.
func NewSpriteAnimation(name string, state ResolvedAnimationState, resource *SpriteResource, instanceIndex int) *SpriteAnimation {
	return &SpriteAnimation{
		Name:              name,
		State:             state,
		resource:          resource,
		instanceIndex:     instanceIndex,
		frameAttribute:    make([]float32, 1),
		flipAttribute:     make([]float32, 2),
		CurrentLocalFrame: state.InitialFrame,
		TimeAccumulator:   0,
		IsPlaying:         true,
		IsActive:          false,
	}
}

// Activate activates the animation at a world transform.
func (a *SpriteAnimation) Activate(transform Matrix4) {
	a.IsActive = true
	a.IsPlaying = true
	a.CurrentLocalFrame = a.State.InitialFrame
	a.TimeAccumulator = 0
	a.updateFrameAttribute()
}

// Deactivate deactivates the animation.
func (a *SpriteAnimation) Deactivate() {
	a.IsActive = false
	a.IsPlaying = false
}

// UpdateTime advances the animation by deltaTime (ms). Returns true if frame changed.
func (a *SpriteAnimation) UpdateTime(deltaTimeMs float64) bool {
	if !a.IsPlaying || !a.IsActive {
		return false
	}

	a.TimeAccumulator += deltaTimeMs
	if a.TimeAccumulator < a.State.FrameDuration {
		return false
	}

	framesToAdvance := int(a.TimeAccumulator / a.State.FrameDuration)
	a.TimeAccumulator = math.Mod(a.TimeAccumulator, a.State.FrameDuration)

	oldFrame := a.CurrentLocalFrame
	nextFrame := a.CurrentLocalFrame + framesToAdvance

	if nextFrame >= a.State.AnimNumFrames {
		if a.State.Loop {
			a.CurrentLocalFrame = nextFrame % a.State.AnimNumFrames
		} else {
			a.CurrentLocalFrame = a.State.AnimNumFrames - 1
			a.IsPlaying = false
		}
	} else {
		a.CurrentLocalFrame = nextFrame
	}

	a.updateFrameAttribute()
	return a.CurrentLocalFrame != oldFrame || !a.IsPlaying
}

func (a *SpriteAnimation) updateFrameAttribute() {
	if a.State.AnimNumFrames <= 0 {
		return
	}
	absoluteFrame := float32(a.State.AnimFrameOffset + a.CurrentLocalFrame)
	if len(a.frameAttribute) > 0 {
		a.frameAttribute[0] = absoluteFrame
	}
	flipX := float32(0)
	flipY := float32(0)
	if a.State.FlipX {
		flipX = 1
	}
	if a.State.FlipY {
		flipY = 1
	}
	if len(a.flipAttribute) > 1 {
		a.flipAttribute[0] = flipX
		a.flipAttribute[1] = flipY
	}
}

// Play resumes the animation.
func (a *SpriteAnimation) Play() {
	if !a.IsActive {
		return
	}
	a.IsPlaying = true
}

// Stop pauses the animation.
func (a *SpriteAnimation) Stop() {
	a.IsPlaying = false
}

// GoToFrame sets the current frame.
func (a *SpriteAnimation) GoToFrame(frame int) {
	if !a.IsActive {
		return
	}
	if a.State.AnimNumFrames <= 0 {
		return
	}
	target := frame
	if target < 0 {
		target = 0
	}
	if target >= a.State.AnimNumFrames {
		target = a.State.AnimNumFrames - 1
	}
	if a.CurrentLocalFrame != target {
		a.CurrentLocalFrame = target
		a.TimeAccumulator = 0
		a.updateFrameAttribute()
	}
}

// SetFrameDuration sets the per-frame duration in ms.
func (a *SpriteAnimation) SetFrameDuration(duration float64) {
	if duration > 0 {
		a.State.FrameDuration = duration
	}
}

// TiledSprite is a sprite with multiple animations (like Three.js SkinnedMesh concept).
type TiledSprite struct {
	ID                 string
	animator           *SpriteAnimator
	animations         map[string]*SpriteAnimation
	currentAnimation   *SpriteAnimation
	transformObject    Transform
	visibleState       bool
	originalDefinition SpriteDefinition
}

// Transform represents position, rotation, scale.
type Transform struct {
	Position Vector3
	Rotation Quaternion
	Scale    Vector3
}

// NewTransform creates an identity transform.
func NewTransform() Transform {
	return Transform{
		Position: Vector3{X: 0, Y: 0, Z: 0},
		Rotation: IdentityQuaternion(),
		Scale:    Vector3{X: 1, Y: 1, Z: 1},
	}
}

// SpriteDefinition defines a sprite with its animations.
type SpriteDefinition struct {
	ID               string
	RenderOrder      int
	DepthWrite       bool
	MaxInstances     int
	Scale            float64
	InitialAnimation string
	Animations       map[string]AnimationDefinition
}

// AnimationDefinition defines a single animation within a sprite.
type AnimationDefinition struct {
	Resource        *SpriteResource
	AnimNumFrames   int
	AnimFrameOffset int
	FrameDuration   float64
	Loop            bool
	InitialFrame    int
	FlipX           bool
	FlipY           bool
}

// NewSpriteDefinition creates a new sprite definition.
func NewSpriteDefinition(initialAnimation string) SpriteDefinition {
	return SpriteDefinition{
		RenderOrder:      0,
		DepthWrite:       true,
		MaxInstances:     1024,
		Scale:            1.0,
		InitialAnimation: initialAnimation,
		Animations:       make(map[string]AnimationDefinition),
	}
}

// SpriteAnimator manages sprite sheet animations (ported from TypeScript).
type SpriteAnimator struct {
	scene            *Scene
	instances        map[string]*TiledSprite
	idCounter        int
	instanceManagers map[string]*SpriteInstanceManager
}

// SpriteInstanceManager manages instanced rendering for a sprite resource.
type SpriteInstanceManager struct {
	InstanceManager *InstancedMesh
	FrameAttribute  []float32
	FlipAttribute   []float32
	UVTileSize      Vector2
}

// NewSpriteAnimator creates a new SpriteAnimator.
func NewSpriteAnimator(scene *Scene) *SpriteAnimator {
	return &SpriteAnimator{
		scene:            scene,
		instances:        make(map[string]*TiledSprite),
		instanceManagers: make(map[string]*SpriteInstanceManager),
	}
}

// CreateSprite creates a new TiledSprite from a definition.
func (sa *SpriteAnimator) CreateSprite(def SpriteDefinition) (*TiledSprite, error) {
	id := def.ID
	if id == "" {
		sa.idCounter++
		id = "sprite_" + itoaInt(sa.idCounter)
	}

	animationInstanceParams := []struct {
		Name      string
		State     ResolvedAnimationState
		Resource  *SpriteResource
		Index     int
		FrameAttr []float32
		FlipAttr  []float32
	}{}

	for animName, animDef := range def.Animations {
		resource := animDef.Resource
		if resource == nil {
			return nil, fmt.Errorf("[SpriteAnimator] animation %q has nil resource", animName)
		}

		key := resource.ID
		manager, ok := sa.instanceManagers[key]
		if !ok {
			maxInstances := def.MaxInstances
			if maxInstances <= 0 {
				maxInstances = 1024
			}
			frameAttr := make([]float32, maxInstances)
			flipAttr := make([]float32, maxInstances*2)
			for i := 0; i < maxInstances; i++ {
				flipAttr[i*2] = 0
				flipAttr[i*2+1] = 0
			}

			geometry := NewGeometry()
			geometry.Vertices = []Vector3{
				{-0.5, 0.5, 0}, {0.5, 0.5, 0}, {0.5, -0.5, 0}, {-0.5, -0.5, 0},
			}
			geometry.Indices = []int{0, 1, 2, 0, 2, 3}
			geometry.SetAttribute("a_frameIndex", frameAttr)
			geometry.SetAttribute("a_flip", flipAttr)

			shader := NewShader("sprite", "", "")
			material := NewMaterial(shader)
			material.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}

			im := NewInstancedMesh(geometry, material, maxInstances)

			manager = &SpriteInstanceManager{
				InstanceManager: im,
				FrameAttribute:  frameAttr,
				FlipAttribute:   flipAttr,
				UVTileSize:      resource.SheetProperties.UVTileSize,
			}
			sa.instanceManagers[key] = manager
		}

		index := len(animationInstanceParams) // simplified: sequential index
		if index >= len(manager.FrameAttribute) {
			return nil, fmt.Errorf("[SpriteAnimator] out of instance slots for %s", resource.ID)
		}

		state := ResolvedAnimationState{
			ImagePath:       resource.SheetProperties.ImagePath,
			TilesetWidth:    resource.SheetProperties.TilesetWidth,
			TilesetHeight:   resource.SheetProperties.TilesetHeight,
			NumFrames:       resource.SheetProperties.NumFrames,
			AnimNumFrames:   animDef.AnimNumFrames,
			AnimFrameOffset: animDef.AnimFrameOffset,
			FrameDuration:   animDef.FrameDuration,
			Loop:            animDef.Loop,
			InitialFrame:    animDef.InitialFrame,
			FlipX:           animDef.FlipX,
			FlipY:           animDef.FlipY,
			Texture:         resource.Texture,
		}

		if state.AnimNumFrames <= 0 {
			state.AnimNumFrames = resource.SheetProperties.NumFrames
		}
		if state.FrameDuration <= 0 {
			state.FrameDuration = 100
		}

		animationInstanceParams = append(animationInstanceParams, struct {
			Name      string
			State     ResolvedAnimationState
			Resource  *SpriteResource
			Index     int
			FrameAttr []float32
			FlipAttr  []float32
		}{
			Name:      animName,
			State:     state,
			Resource:  resource,
			Index:     index,
			FrameAttr: manager.FrameAttribute[index : index+1],
			FlipAttr:  manager.FlipAttribute[index*2 : index*2+2],
		})
	}

	if def.InitialAnimation == "" {
		_, ok := def.Animations[def.InitialAnimation]
		if !ok {
			if len(animationInstanceParams) == 0 {
				return nil, fmt.Errorf("[SpriteAnimator] no animations defined for sprite %q", id)
			}
			def.InitialAnimation = animationInstanceParams[0].Name
		}
	}

	sprite := &TiledSprite{
		ID:                 id,
		animator:           sa,
		animations:         make(map[string]*SpriteAnimation),
		transformObject:    NewTransform(),
		visibleState:       true,
		originalDefinition: def,
	}

	for _, params := range animationInstanceParams {
		anim := NewSpriteAnimation(params.Name, params.State, params.Resource, params.Index)
		anim.frameAttribute = params.FrameAttr
		anim.flipAttribute = params.FlipAttr
		sprite.animations[params.Name] = anim
	}

	initialAnim := sprite.animations[def.InitialAnimation]
	if initialAnim == nil {
		return nil, fmt.Errorf("[TiledSprite] initial animation %q not found", def.InitialAnimation)
	}
	sprite.currentAnimation = initialAnim
	initialAnim.Activate(sprite.getWorldTransform())

	sa.instances[id] = sprite
	return sprite, nil
}

// Update advances all sprites by deltaTime (ms).
func (sa *SpriteAnimator) Update(deltaTimeMs float64) {
	for _, sprite := range sa.instances {
		sprite.Update(deltaTimeMs)
	}
}

// RemoveSprite removes a sprite by ID.
func (sa *SpriteAnimator) RemoveSprite(id string) {
	if sprite, ok := sa.instances[id]; ok {
		sprite.Destroy()
		delete(sa.instances, id)
	}
}

// RemoveAllSprites removes all sprites.
func (sa *SpriteAnimator) RemoveAllSprites() {
	for id := range sa.instances {
		sa.RemoveSprite(id)
	}
}

// GetSprite returns a sprite by ID.
func (sa *SpriteAnimator) GetSprite(id string) (*TiledSprite, bool) {
	s, ok := sa.instances[id]
	return s, ok
}

// Update advances the tiled sprite by deltaTime (ms).
func (ts *TiledSprite) Update(deltaTimeMs float64) {
	if ts.currentAnimation != nil {
		ts.currentAnimation.UpdateTime(deltaTimeMs)
	}
}

// Play plays the current animation.
func (ts *TiledSprite) Play() {
	if ts.currentAnimation != nil {
		ts.currentAnimation.Play()
	}
}

// Stop stops the current animation.
func (ts *TiledSprite) Stop() {
	if ts.currentAnimation != nil {
		ts.currentAnimation.Stop()
	}
}

// GoToFrame sets the current animation frame.
func (ts *TiledSprite) GoToFrame(frame int) {
	if ts.currentAnimation != nil {
		ts.currentAnimation.GoToFrame(frame)
	}
}

// SetAnimation switches to a different animation by name.
func (ts *TiledSprite) SetAnimation(name string) error {
	newAnim, ok := ts.animations[name]
	if !ok {
		return fmt.Errorf("[TiledSprite] animation %q not found", name)
	}
	if ts.currentAnimation != nil && ts.currentAnimation.Name != name {
		ts.currentAnimation.Deactivate()
	}
	ts.currentAnimation = newAnim
	if ts.visibleState {
		newAnim.Activate(ts.getWorldTransform())
	} else {
		newAnim.Deactivate()
	}
	return nil
}

// Destroy cleans up the tiled sprite.
func (ts *TiledSprite) Destroy() {
	if ts.currentAnimation != nil {
		ts.currentAnimation.Deactivate()
	}
	ts.visibleState = false
}

// GetPosition returns the sprite position.
func (ts *TiledSprite) GetPosition() Vector3 {
	return ts.transformObject.Position
}

// SetPosition sets the sprite position.
func (ts *TiledSprite) SetPosition(x, y, z float64) {
	ts.transformObject.Position = Vector3{X: x, Y: y, Z: z}
	if ts.currentAnimation != nil && ts.visibleState {
		ts.currentAnimation.Activate(ts.getWorldTransform())
	}
}

// GetScale returns the sprite scale.
func (ts *TiledSprite) GetScale() Vector3 {
	return ts.transformObject.Scale
}

// SetScale sets the sprite scale.
func (ts *TiledSprite) SetScale(s Vector3) {
	ts.transformObject.Scale = s
	if ts.currentAnimation != nil && ts.visibleState {
		ts.currentAnimation.Activate(ts.getWorldTransform())
	}
}

// SetRotation sets the sprite rotation from a quaternion.
func (ts *TiledSprite) SetRotation(q Quaternion) {
	ts.transformObject.Rotation = q
	if ts.currentAnimation != nil && ts.visibleState {
		ts.currentAnimation.Activate(ts.getWorldTransform())
	}
}

// GetWorldTransform returns the current world transform matrix.
func (ts *TiledSprite) GetWorldTransform() Matrix4 {
	return ts.getWorldTransform()
}

func (ts *TiledSprite) getWorldTransform() Matrix4 {
	return Compose(ts.transformObject.Position, ts.transformObject.Rotation, ts.transformObject.Scale)
}

// IsVisible returns the visibility state.
func (ts *TiledSprite) IsVisible() bool {
	return ts.visibleState
}

// SetVisible sets the visibility state.
func (ts *TiledSprite) SetVisible(v bool) {
	if ts.visibleState == v {
		return
	}
	ts.visibleState = v
	if ts.currentAnimation != nil {
		if v {
			ts.currentAnimation.Activate(ts.getWorldTransform())
		} else {
			ts.currentAnimation.Deactivate()
		}
	}
}

// itoaInt converts an integer to a string.
func itoaInt(n int) string {
	if n == 0 {
		return "0"
	}
	var result string
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// ============================================================
// EXPLOSION EFFECTS (ported from TypeScript)
// ============================================================

// ExplosionEffectParameters describes an explosion effect.
type ExplosionEffectParameters struct {
	NumRows               int
	NumCols               int
	DurationMs            float64
	Strength              float64
	StrengthVariation     float64
	Gravity               float64
	GravityScale          float64
	FadeOut               bool
	AngularVelocityMin    Vector3
	AngularVelocityMax    Vector3
	InitialVelocityYBoost float64
	ZVariationStrength    float64
}

// DefaultExplosionEffectParameters returns default explosion parameters.
func DefaultExplosionEffectParameters() ExplosionEffectParameters {
	return ExplosionEffectParameters{
		NumRows:               5,
		NumCols:               5,
		DurationMs:            2000,
		Strength:              5,
		StrengthVariation:     0.5,
		Gravity:               9.8,
		GravityScale:          0.15,
		FadeOut:               true,
		AngularVelocityMin:    Vector3{X: -math.Pi, Y: -math.Pi, Z: -math.Pi},
		AngularVelocityMax:    Vector3{X: math.Pi, Y: math.Pi, Z: math.Pi},
		InitialVelocityYBoost: 1.0,
		ZVariationStrength:    0.3,
	}
}

// ExplosionHandle is a handle to an active explosion effect.
type ExplosionHandle struct {
	Effect          *ExplodingSpriteEffect
	HasBeenRestored bool
}

// ExplodingSpriteEffect creates an explosion of sprite particles.
type ExplodingSpriteEffect struct {
	Scene          *Scene
	Resource       *SpriteResource
	FrameUVOffset  Vector2
	FrameUVSize    Vector2
	WorldTransform Matrix4
	Params         ExplosionEffectParameters

	InstancedMesh *InstancedMesh
	NumParticles  int
	IsActive      bool
	TimeElapsedMs float64
}

// NewExplodingSpriteEffect creates a new explosion effect.
func NewExplodingSpriteEffect(scene *Scene, resource *SpriteResource, frameUvOffset, frameUvSize Vector2, worldTransform Matrix4, params ExplosionEffectParameters) *ExplodingSpriteEffect {
	effect := &ExplodingSpriteEffect{
		Scene:          scene,
		Resource:       resource,
		FrameUVOffset:  frameUvOffset,
		FrameUVSize:    frameUvSize,
		WorldTransform: worldTransform,
		Params:         params,
		NumParticles:   params.NumRows * params.NumCols,
		IsActive:       true,
	}
	effect.createParticles()
	return effect
}

func (e *ExplodingSpriteEffect) createParticles() {
	if e.NumParticles == 0 {
		return
	}

	particleUnitWidth := 1.0 / float64(e.Params.NumCols)
	particleUnitHeight := 1.0 / float64(e.Params.NumRows)

	geometry := PlaneGeometry(particleUnitWidth, particleUnitHeight)
	shader := NewShader("explosion", "", "")
	material := NewMaterial(shader)
	material.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}

	e.InstancedMesh = NewInstancedMesh(geometry, material, e.NumParticles)

	spriteWorldCenter := Vector3{
		X: e.WorldTransform.Elements[12],
		Y: e.WorldTransform.Elements[13],
		Z: e.WorldTransform.Elements[14],
	}

	for r := 0; r < e.Params.NumRows; r++ {
		for c := 0; c < e.Params.NumCols; c++ {
			localX := (float64(c)+0.5)*particleUnitWidth - 0.5
			localY := (float64(r)+0.5)*particleUnitHeight - 0.5
			localPos := Vector3{X: localX, Y: localY, Z: 0}

			// Apply world transform (simplified: just translate)
			worldPos := Vector3{
				X: localPos.X + e.WorldTransform.Elements[12],
				Y: localPos.Y + e.WorldTransform.Elements[13],
				Z: localPos.Z + e.WorldTransform.Elements[14],
			}

			velocityDir := Vector3{
				X: worldPos.X - spriteWorldCenter.X,
				Y: worldPos.Y - spriteWorldCenter.Y,
				Z: worldPos.Z - spriteWorldCenter.Z,
			}
			if velocityDir.X*velocityDir.X+velocityDir.Y*velocityDir.Y+velocityDir.Z*velocityDir.Z < 0.0001 {
				velocityDir = Vector3{X: rand.Float64() - 0.5, Y: rand.Float64() - 0.5, Z: rand.Float64() - 0.5}
			}
			vLen := math.Sqrt(velocityDir.X*velocityDir.X + velocityDir.Y*velocityDir.Y + velocityDir.Z*velocityDir.Z)
			if vLen > 0 {
				velocityDir.X /= vLen
				velocityDir.Y /= vLen
				velocityDir.Z /= vLen
			}

			strengthFactor := 1.0 - e.Params.StrengthVariation*0.5 + rand.Float64()*e.Params.StrengthVariation
			strength := e.Params.Strength * strengthFactor * 0.1
			velocity := Vector3{
				X: velocityDir.X * strength,
				Y: velocityDir.Y * strength,
				Z: velocityDir.Z * strength,
			}

			velocity.Z += (rand.Float64() - 0.5) * strength * e.Params.ZVariationStrength
			velocity.Y += e.Params.Strength * e.Params.InitialVelocityYBoost * rand.Float64()

			lifeVariation := 0.8 + rand.Float64()*0.4

			u0 := e.FrameUVOffset.X + float64(c)/float64(e.Params.NumCols)*e.FrameUVSize.X
			v0 := e.FrameUVOffset.Y + float64(r)/float64(e.Params.NumRows)*e.FrameUVSize.Y
			uSize := e.FrameUVSize.X / float64(e.Params.NumCols)
			vSize := e.FrameUVSize.Y / float64(e.Params.NumRows)

			_ = velocity
			_ = lifeVariation
			_ = u0
			_ = v0
			_ = uSize
			_ = vSize

			index := r*e.Params.NumCols + c
			e.InstancedMesh.SetMatrixAt(index, e.WorldTransform)
		}
	}

	e.Scene.AddObject(e.InstancedMesh)
	e.TimeElapsedMs = 0
}

// Update advances the explosion effect by deltaTime (ms).
func (e *ExplodingSpriteEffect) Update(deltaTimeMs float64) {
	if !e.IsActive {
		return
	}
	e.TimeElapsedMs += deltaTimeMs
	if e.TimeElapsedMs >= e.Params.DurationMs {
		e.Dispose()
	}
}

// Dispose removes the explosion effect.
func (e *ExplodingSpriteEffect) Dispose() {
	if !e.IsActive {
		return
	}
	e.IsActive = false
}

// ExplosionManager manages multiple explosion effects.
type ExplosionManager struct {
	Scene            *Scene
	ActiveExplosions []*ExplodingSpriteEffect
}

// NewExplosionManager creates a new explosion manager.
func NewExplosionManager(scene *Scene) *ExplosionManager {
	return &ExplosionManager{
		Scene:            scene,
		ActiveExplosions: make([]*ExplodingSpriteEffect, 0),
	}
}

// CreateExplosionForSprite creates an explosion effect for a tiled sprite.
func (em *ExplosionManager) CreateExplosionForSprite(sprite *TiledSprite, userParams ...ExplosionEffectParameters) *ExplosionHandle {
	params := DefaultExplosionEffectParameters()
	if len(userParams) > 0 {
		params = userParams[0]
	}

	// Use current frame UV
	anim := sprite.currentAnimation
	if anim == nil {
		return nil
	}

	frameUvOffset := Vector2{
		X: float64(anim.State.AnimFrameOffset+anim.CurrentLocalFrame) * float64(anim.State.Texture.Width) / 16, // simplified
		Y: 0,
	}
	frameUvSize := Vector2{
		X: 1.0 / float64(anim.State.AnimNumFrames),
		Y: 1.0,
	}

	spriteRes := NewSpriteResource("explosion", anim.State.Texture, 1, 1, 1)
	effect := NewExplodingSpriteEffect(em.Scene, spriteRes, frameUvOffset, frameUvSize, sprite.getWorldTransform(), params)
	em.ActiveExplosions = append(em.ActiveExplosions, effect)

	return &ExplosionHandle{
		Effect:          effect,
		HasBeenRestored: false,
	}
}

// Update advances all explosions by deltaTime (ms).
func (em *ExplosionManager) Update(deltaTimeMs float64) {
	for i := len(em.ActiveExplosions) - 1; i >= 0; i-- {
		em.ActiveExplosions[i].Update(deltaTimeMs)
		if !em.ActiveExplosions[i].IsActive {
			em.ActiveExplosions = append(em.ActiveExplosions[:i], em.ActiveExplosions[i+1:]...)
		}
	}
}

// DisposeAll removes all active explosions.
func (em *ExplosionManager) DisposeAll() {
	for _, exp := range em.ActiveExplosions {
		exp.Dispose()
	}
	em.ActiveExplosions = make([]*ExplodingSpriteEffect, 0)
}

// ============================================================
// PARTICLE GENERATOR (simplified port from TypeScript)
// ============================================================

// SpriteParticleGenerator generates particles from a sprite texture.
type SpriteParticleGenerator struct {
	Scene        *Scene
	Resource     *SpriteResource
	MaxParticles int
	Particles    []*Particle
	Emitter      Vector3
	EmissionRate float64
	Accumulator  float64
	Gravity      Vector3
	Enabled      bool
}

// NewSpriteParticleGenerator creates a new particle generator.
func NewSpriteParticleGenerator(scene *Scene, resource *SpriteResource, maxParticles int) *SpriteParticleGenerator {
	return &SpriteParticleGenerator{
		Scene:        scene,
		Resource:     resource,
		MaxParticles: maxParticles,
		Particles:    make([]*Particle, 0, maxParticles),
		Emitter:      Vector3{X: 0, Y: 0, Z: 0},
		Gravity:      Vector3{X: 0, Y: -9.8, Z: 0},
		Enabled:      true,
	}
}

// Emit spawns count new particles.
func (pg *SpriteParticleGenerator) Emit(count int) {
	if !pg.Enabled {
		return
	}
	for i := 0; i < count; i++ {
		var p *Particle
		for _, existing := range pg.Particles {
			if !existing.Active {
				p = existing
				break
			}
		}
		if p == nil && len(pg.Particles) < pg.MaxParticles {
			p = &Particle{}
			pg.Particles = append(pg.Particles, p)
		}
		if p == nil {
			continue
		}
		p.Position = pg.Emitter
		p.Velocity = Vector3{
			X: (rand.Float64() - 0.5) * 2,
			Y: rand.Float64()*2 + 1,
			Z: (rand.Float64() - 0.5) * 2,
		}
		p.Acceleration = pg.Gravity
		p.Life = 1.0 + rand.Float64()
		p.MaxLife = p.Life
		p.Size = 0.5 + rand.Float64()
		p.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		p.Active = true
	}
}

// Update advances the particle simulation.
func (pg *SpriteParticleGenerator) Update(deltaTime float64) {
	if !pg.Enabled {
		return
	}
	if pg.EmissionRate > 0 {
		pg.Accumulator += deltaTime
		toEmit := int(pg.Accumulator * pg.EmissionRate)
		if toEmit > 0 {
			pg.Accumulator -= float64(toEmit) / pg.EmissionRate
			pg.Emit(toEmit)
		}
	}
	for _, p := range pg.Particles {
		if !p.Active {
			continue
		}
		p.Velocity.X += p.Acceleration.X * deltaTime
		p.Velocity.Y += p.Acceleration.Y * deltaTime
		p.Velocity.Z += p.Acceleration.Z * deltaTime
		p.Position.X += p.Velocity.X * deltaTime
		p.Position.Y += p.Velocity.Y * deltaTime
		p.Position.Z += p.Velocity.Z * deltaTime
		p.Life -= deltaTime
		if p.Life <= 0 {
			p.Active = false
		}
	}
}

// Render renders all active particles.
func (pg *SpriteParticleGenerator) Render(buffer *core.NativeBuffer, r *Renderer) {
	if !pg.Enabled {
		return
	}
	for _, p := range pg.Particles {
		if !p.Active {
			continue
		}
		alpha := p.Life / p.MaxLife
		if alpha < 0 {
			alpha = 0
		}
		if alpha > 1 {
			alpha = 1
		}
		rCol, gCol, bCol, _ := p.Color.RGBA()
		c := color.RGBA{
			R: uint8(rCol >> 8),
			G: uint8(gCol >> 8),
			B: uint8(bCol >> 8),
			A: uint8(float64(255) * alpha),
		}
		x, y := r.projectWithCamera(p.Position.X, p.Position.Y, p.Position.Z)
		if x >= 0 && x < buffer.Width() && y >= 0 && y < buffer.Height() {
			buffer.SetCell(x, y, '#', uint16(c.R), uint16(c.G), 0)
		}
	}
}

// ============================================================
// PHYSICS EXPLOSION EFFECTS (ported from TypeScript)
// ============================================================

// PhysicsExplosionEffectParameters describes a physics-based explosion.
type PhysicsExplosionEffectParameters struct {
	NumRows        int
	NumCols        int
	DurationMs     float64
	ExplosionForce float64
	ForceVariation float64
	TorqueStrength float64
	GravityScale   float64
	FadeOut        bool
	LinearDamping  float64
	AngularDamping float64
	Restitution    float64
	Friction       float64
	Density        float64
}

// DefaultPhysicsExplosionEffectParameters returns default physics explosion parameters.
func DefaultPhysicsExplosionEffectParameters() PhysicsExplosionEffectParameters {
	return PhysicsExplosionEffectParameters{
		NumRows:        5,
		NumCols:        5,
		DurationMs:     3000,
		ExplosionForce: 25.0,
		ForceVariation: 0.4,
		TorqueStrength: 15.0,
		GravityScale:   1.0,
		FadeOut:        true,
		LinearDamping:  0.8,
		AngularDamping: 0.5,
		Restitution:    0.3,
		Friction:       0.7,
		Density:        1.0,
	}
}

// PhysicsExplosionHandle is a handle to a physics explosion.
type PhysicsExplosionHandle struct {
	Effect          *PhysicsExplodingSpriteEffect
	HasBeenRestored bool
}

// PhysicsExplodingSpriteEffect creates a physics-based explosion.
type PhysicsExplodingSpriteEffect struct {
	Scene          *Scene
	PhysicsWorld   PhysicsEngine
	Resource       *SpriteResource
	FrameUVOffset  Vector2
	FrameUVSize    Vector2
	WorldTransform Matrix4
	Params         PhysicsExplosionEffectParameters

	Particles     []*physicsBody
	NumParticles  int
	InstancedMesh *InstancedMesh
	IsActive      bool
	TimeElapsedMs float64
}

// NewPhysicsExplodingSpriteEffect creates a new physics explosion effect.
func NewPhysicsExplodingSpriteEffect(scene *Scene, physicsWorld PhysicsEngine, resource *SpriteResource, frameUvOffset, frameUvSize Vector2, worldTransform Matrix4, params PhysicsExplosionEffectParameters) *PhysicsExplodingSpriteEffect {
	effect := &PhysicsExplodingSpriteEffect{
		Scene:          scene,
		PhysicsWorld:   physicsWorld,
		Resource:       resource,
		FrameUVOffset:  frameUvOffset,
		FrameUVSize:    frameUvSize,
		WorldTransform: worldTransform,
		Params:         params,
		NumParticles:   params.NumRows * params.NumCols,
		IsActive:       true,
	}
	effect.createPhysicsParticles()
	return effect
}

func (e *PhysicsExplodingSpriteEffect) createPhysicsParticles() {
	if e.NumParticles == 0 {
		return
	}

	particleUnitWidth := 1.0 / float64(e.Params.NumCols)
	particleUnitHeight := 1.0 / float64(e.Params.NumRows)

	spriteWorldCenter := Vector3{
		X: e.WorldTransform.Elements[12],
		Y: e.WorldTransform.Elements[13],
	}

	geometry := PlaneGeometry(particleUnitWidth, particleUnitHeight)
	shader := NewShader("physics-explosion", "", "")
	material := NewMaterial(shader)
	material.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}

	e.InstancedMesh = NewInstancedMesh(geometry, material, e.NumParticles)

	for r := 0; r < e.Params.NumRows; r++ {
		for c := 0; c < e.Params.NumCols; c++ {
			localX := (float64(c)+0.5)*particleUnitWidth - 0.5
			localY := (float64(r)+0.5)*particleUnitHeight - 0.5

			worldPos := Vector3{
				X: localX + e.WorldTransform.Elements[12],
				Y: localY + e.WorldTransform.Elements[13],
			}

			body := e.PhysicsWorld.CreateRigidBody(PhysicsRigidBodyDesc{
				Translation:    Vector2{X: worldPos.X, Y: worldPos.Y},
				LinearDamping:  e.Params.LinearDamping,
				AngularDamping: e.Params.AngularDamping,
			})

			e.PhysicsWorld.CreateCollider(PhysicsColliderDesc{
				Width:       particleUnitWidth * 0.8,
				Height:      particleUnitHeight * 0.8,
				Restitution: e.Params.Restitution,
				Friction:    e.Params.Friction,
				Density:     e.Params.Density,
			}, body)

			explosionDir := Vector2{
				X: worldPos.X - spriteWorldCenter.X,
				Y: worldPos.Y - spriteWorldCenter.Y,
			}
			if explosionDir.X*explosionDir.X+explosionDir.Y*explosionDir.Y < 0.0001 {
				explosionDir = Vector2{X: rand.Float64() - 0.5, Y: rand.Float64() - 0.5}
			}
			dLen := math.Sqrt(explosionDir.X*explosionDir.X + explosionDir.Y*explosionDir.Y)
			if dLen > 0 {
				explosionDir.X /= dLen
				explosionDir.Y /= dLen
			}

			forceFactor := 1.0 - e.Params.ForceVariation*0.5 + rand.Float64()*e.Params.ForceVariation
			explosionForce := e.Params.ExplosionForce * forceFactor
			body.ApplyImpulse(Vector2{
				X: explosionDir.X * explosionForce,
				Y: explosionDir.Y*explosionForce + explosionForce*0.3,
			})

			torque := (rand.Float64() - 0.5) * e.Params.TorqueStrength
			body.ApplyTorqueImpulse(torque)

			index := r*e.Params.NumCols + c
			e.InstancedMesh.SetMatrixAt(index, e.WorldTransform)
			e.Particles = append(e.Particles, body.(*physicsBody))
		}
	}

	e.Scene.AddObject(e.InstancedMesh)
	e.TimeElapsedMs = 0
}

// Update advances the physics explosion by deltaTime (ms).
func (e *PhysicsExplodingSpriteEffect) Update(deltaTimeMs float64) {
	if !e.IsActive {
		return
	}
	e.TimeElapsedMs += deltaTimeMs

	axis := Vector3{X: 0, Y: 0, Z: 1}
	for i, body := range e.Particles {
		pos := body.GetTranslation()
		rot := body.GetRotation()
		q := IdentityQuaternion()
		q.SetFromAxisAngle(axis, rot)

		m := Compose(Vector3{X: pos.X, Y: pos.Y, Z: 0}, q, Vector3{X: 1, Y: 1, Z: 1})
		e.InstancedMesh.SetMatrixAt(i, m)
	}

	if e.TimeElapsedMs >= e.Params.DurationMs {
		e.Dispose()
	}
}

// Dispose removes the physics explosion effect.
func (e *PhysicsExplodingSpriteEffect) Dispose() {
	if !e.IsActive {
		return
	}
	e.IsActive = false
	for _, body := range e.Particles {
		e.PhysicsWorld.RemoveRigidBody(body)
	}
	e.Particles = make([]*physicsBody, 0)
}

// PhysicsExplosionManager manages physics explosions.
type PhysicsExplosionManager struct {
	Scene            *Scene
	PhysicsWorld     PhysicsEngine
	ActiveExplosions []*PhysicsExplodingSpriteEffect
}

// NewPhysicsExplosionManager creates a new physics explosion manager.
func NewPhysicsExplosionManager(scene *Scene, physicsWorld PhysicsEngine) *PhysicsExplosionManager {
	return &PhysicsExplosionManager{
		Scene:            scene,
		PhysicsWorld:     physicsWorld,
		ActiveExplosions: make([]*PhysicsExplodingSpriteEffect, 0),
	}
}

// CreateExplosionForSprite creates a physics explosion for a tiled sprite.
func (pem *PhysicsExplosionManager) CreateExplosionForSprite(sprite *TiledSprite, userParams ...PhysicsExplosionEffectParameters) *PhysicsExplosionHandle {
	params := DefaultPhysicsExplosionEffectParameters()
	if len(userParams) > 0 {
		params = userParams[0]
	}

	anim := sprite.currentAnimation
	if anim == nil {
		return nil
	}

	frameUvOffset := Vector2{X: 0, Y: 0}
	frameUvSize := Vector2{X: 1.0 / float64(anim.State.AnimNumFrames), Y: 1.0}

	spriteRes := NewSpriteResource("physics-explosion", anim.State.Texture, 1, 1, 1)
	effect := NewPhysicsExplodingSpriteEffect(pem.Scene, pem.PhysicsWorld, spriteRes, frameUvOffset, frameUvSize, sprite.getWorldTransform(), params)
	pem.ActiveExplosions = append(pem.ActiveExplosions, effect)

	return &PhysicsExplosionHandle{
		Effect:          effect,
		HasBeenRestored: false,
	}
}

// Update advances all physics explosions by deltaTime (ms).
func (pem *PhysicsExplosionManager) Update(deltaTimeMs float64) {
	for i := len(pem.ActiveExplosions) - 1; i >= 0; i-- {
		pem.ActiveExplosions[i].Update(deltaTimeMs)
		if !pem.ActiveExplosions[i].IsActive {
			pem.ActiveExplosions = append(pem.ActiveExplosions[:i], pem.ActiveExplosions[i+1:]...)
		}
	}
}

// DisposeAll removes all active physics explosions.
func (pem *PhysicsExplosionManager) DisposeAll() {
	for _, exp := range pem.ActiveExplosions {
		exp.Dispose()
	}
	pem.ActiveExplosions = make([]*PhysicsExplodingSpriteEffect, 0)
}

// ============================================================
// RENDERER RENDER METHOD: use enhanced rendering with lighting
// ============================================================

// renderObject is the original simple renderer; kept for backward compatibility.
// The Render method now uses enhanced lighting when materials are present.
