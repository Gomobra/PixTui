package three

import (
	"image/color"
	"testing"
)

func TestRendererCreate(t *testing.T) {
	renderer := NewRenderer(80, 24)
	if renderer == nil {
		t.Fatal("expected non-nil renderer")
	}
	if renderer.width != 80 || renderer.height != 24 {
		t.Errorf("expected 80x24, got %dx%d", renderer.width, renderer.height)
	}
}

func TestSceneCreate(t *testing.T) {
	scene := NewScene()
	if scene == nil {
		t.Fatal("expected non-nil scene")
	}
	if len(scene.GetObjects()) != 0 {
		t.Error("expected empty scene")
	}
}

func TestCameraCreate(t *testing.T) {
	camera := NewCamera()
	if camera == nil {
		t.Fatal("expected non-nil camera")
	}
}

func TestMeshCreate(t *testing.T) {
	geometry := NewGeometry()
	geometry.Vertices = []Vector3{
		{X: -1, Y: -1, Z: 0},
		{X: 1, Y: -1, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	geometry.Indices = []int{0, 1, 2}
	material := NewMaterial(NewShader("test", "v", "f"))
	material.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	mesh := NewMesh(geometry, material)
	if mesh == nil {
		t.Fatal("expected non-nil mesh")
	}
	if len(mesh.GetVertices()) != 3 {
		t.Errorf("expected 3 vertices, got %d", len(mesh.GetVertices()))
	}
	if len(mesh.GetIndices()) != 3 {
		t.Errorf("expected 3 indices, got %d", len(mesh.GetIndices()))
	}
}

func TestShaderCreate(t *testing.T) {
	shader := NewShader("lit", "vertex src", "fragment src")
	if shader == nil {
		t.Fatal("expected non-nil shader")
	}
	if shader.ID != "lit" {
		t.Errorf("expected id 'lit', got %s", shader.ID)
	}
}

func TestMaterialCreate(t *testing.T) {
	shader := NewShader("lit", "v", "f")
	material := NewMaterial(shader)
	if material == nil {
		t.Fatal("expected non-nil material")
	}
	if material.Shader != shader {
		t.Error("expected material shader to match")
	}
	material.SetUniform("color", color.RGBA{R: 255, G: 0, B: 0, A: 255})
	v, ok := material.GetUniform("color")
	if !ok {
		t.Error("expected uniform to exist")
	}
	_ = v
}

func TestGeometryCreate(t *testing.T) {
	geom := NewGeometry()
	if geom == nil {
		t.Fatal("expected non-nil geometry")
	}
	geom.SetAttribute("a_position", []float32{1, 2, 3})
	val, ok := geom.GetAttribute("a_position")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if len(val) != 3 {
		t.Errorf("expected 3 values, got %d", len(val))
	}
}

func TestSpriteCreate(t *testing.T) {
	tex := NewTexture("tex1", 16, 16, make([]uint8, 16*16*4))
	sprite := NewSprite(tex)
	if sprite == nil {
		t.Fatal("expected non-nil sprite")
	}
	if sprite.Texture != tex {
		t.Error("expected sprite texture to match")
	}
}

func TestParticleSystemCreate(t *testing.T) {
	ps := NewParticleSystem(100)
	if ps == nil {
		t.Fatal("expected non-nil particle system")
	}
	if ps.MaxParticles != 100 {
		t.Errorf("expected max 100 particles, got %d", ps.MaxParticles)
	}
	ps.Emit(10)
	ps.Update(0.016)
}

func TestPlanckPhysicsCreate(t *testing.T) {
	physics := NewPlanckPhysics(0, -9.8)
	if physics == nil {
		t.Fatal("expected non-nil physics")
	}
	body := physics.CreateRigidBody(PhysicsRigidBodyDesc{
		Translation:    Vector2{X: 0, Y: 0},
		LinearDamping:  0.1,
		AngularDamping: 0.1,
	})
	if body == nil {
		t.Fatal("expected non-nil body")
	}
	physics.Step(0.016)
	_ = body.GetTranslation()
	_ = body.GetRotation()
}

func TestRapierPhysicsCreate(t *testing.T) {
	physics := NewRapierPhysics(0, -9.8)
	if physics == nil {
		t.Fatal("expected non-nil physics")
	}
	body := physics.CreateRigidBody(PhysicsRigidBodyDesc{
		Translation:    Vector2{X: 0, Y: 0},
		LinearDamping:  0.1,
		AngularDamping: 0.1,
	})
	if body == nil {
		t.Fatal("expected non-nil body")
	}
	physics.Step(0.016)
	_ = body.GetTranslation()
	_ = body.GetRotation()
}

func TestSpriteRendererCreate(t *testing.T) {
	renderer := NewRenderer(80, 24)
	sr := NewSpriteRenderer(renderer)
	if sr == nil {
		t.Fatal("expected non-nil sprite renderer")
	}
	tex := NewTexture("tex1", 8, 8, make([]uint8, 8*8*4))
	sprite := NewSprite(tex)
	sr.AddSprite(sprite)
	if len(sr.sprites) != 1 {
		t.Errorf("expected 1 sprite, got %d", len(sr.sprites))
	}
}

func TestSceneLighting(t *testing.T) {
	scene := NewScene()
	light := NewLight(LightAmbient, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 1)
	scene.AddLight(light)
	if len(scene.GetLights()) != 1 {
		t.Errorf("expected 1 light, got %d", len(scene.GetLights()))
	}
	scene.SetFog(Fog{Enabled: true, Near: 1, Far: 100})
	if !scene.GetFog().Enabled {
		t.Error("expected fog enabled")
	}
}

func TestCameraModes(t *testing.T) {
	cam := NewPerspectiveCamera(60, 16.0/9.0, 0.1, 1000)
	if cam == nil {
		t.Fatal("expected non-nil camera")
	}
	if cam.Mode != CameraPerspective {
		t.Error("expected perspective mode")
	}
	ortho := NewOrthographicCamera(10, 16.0/9.0, 0.1, 100)
	if ortho == nil {
		t.Fatal("expected non-nil ortho camera")
	}
	if ortho.Mode != CameraOrthographic {
		t.Error("expected orthographic mode")
	}
}

func TestNewGeometries(t *testing.T) {
	cyl := CylinderGeometry(1, 1, 2, 8)
	if cyl == nil || len(cyl.Vertices) == 0 || len(cyl.Indices) == 0 {
		t.Error("expected non-empty cylinder geometry")
	}
	cone := ConeGeometry(1, 2, 8)
	if cone == nil || len(cone.Vertices) == 0 || len(cone.Indices) == 0 {
		t.Error("expected non-empty cone geometry")
	}
	torus := TorusGeometry(2, 0.5, 16, 16)
	if torus == nil || len(torus.Vertices) == 0 || len(torus.Indices) == 0 {
		t.Error("expected non-empty torus geometry")
	}
	ring := RingGeometry(0.5, 1, 16)
	if ring == nil || len(ring.Vertices) == 0 || len(ring.Indices) == 0 {
		t.Error("expected non-empty ring geometry")
	}
}

func TestRendererClearColorAndCamera(t *testing.T) {
	renderer := NewRenderer(80, 24)
	if renderer.GetClearColor() == nil {
		t.Error("expected non-nil clear color")
	}
	custom := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	renderer.SetClearColor(custom)
	if renderer.GetClearColor() != custom {
		t.Error("expected clear color to match")
	}
	cam := NewPerspectiveCamera(60, 16.0/9.0, 0.1, 1000)
	renderer.SetActiveCamera(cam)
	if renderer.GetActiveCamera() != cam {
		t.Error("expected active camera to match")
	}
}

func TestSceneAddRemove(t *testing.T) {
	scene := NewScene()
	geom := NewGeometry()
	geom.Vertices = []Vector3{{X: 0, Y: 0, Z: 0}}
	geom.Indices = []int{}
	mat := NewMaterial(NewShader("t", "v", "f"))
	mesh := NewMesh(geom, mat)
	scene.Add(mesh)
	if len(scene.GetObjects()) != 1 {
		t.Errorf("expected 1 object after Add, got %d", len(scene.GetObjects()))
	}
	scene.Remove(mesh)
	if len(scene.GetObjects()) != 0 {
		t.Errorf("expected 0 objects after Remove, got %d", len(scene.GetObjects()))
	}
}
