# PixTui Three

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

3D renderer for PixTui.

## About

PixTui Three provides a CPU-based 3D renderer for PixTui applications. It supports scene management, camera controls, meshes, geometries, materials, lighting, and physics.

## Features

- **Renderer** — CPU rasterization with clear color, render stats, and destroy
- **Scene** — Object management, lights, fog, background color
- **Camera** — Perspective and orthographic modes with projection matrix
- **Mesh** — Geometry + Material combination with transforms
- **Geometry** — Box, Plane, Sphere, Cylinder, Cone, Torus, Ring
- **Material** — Basic, Lambert, Phong with shaders and uniforms
- **Light** — Ambient, directional, and point lights
- **Fog** — Linear fog effect
- **Sprite** — 2D sprite rendering with texture support
- **ParticleSystem** — Particle emission and rendering
- **PhysicsEngine** — Physics simulation interface
- **ExplosionEffect** — Particle explosion effects

## Installation

```bash
go get github.com/Gomobra/PixTui/packages/three
```

## Quick Start

```go
package main

import (
	"github.com/Gomobra/PixTui/packages/three"
)

func main() {
	renderer := three.NewRenderer(80, 24)
	scene := three.NewScene()
	camera := three.NewCamera()

	mesh := three.NewMesh(
		three.NewBoxGeometry(1, 1, 1),
		three.NewMaterial(three.NewShader("basic", "", "")),
	)
	scene.Add(mesh)

	renderer.Render(scene, camera)
}
```

## Components

| Component | Description |
|---|---|
| `Renderer` | 3D renderer with clear color and render stats |
| `Scene` | Scene graph with objects, lights, and fog |
| `Camera` | Perspective/orthographic camera |
| `Mesh` | Geometry + Material renderable |
| `Geometry` | Vertex and index data |
| `Material` | Shader, uniforms, and blend modes |
| `Light` | Ambient, directional, point lights |
| `Fog` | Linear fog |
| `Sprite` | Textured 2D sprite |
| `ParticleSystem` | Particle emission and rendering |
| `PhysicsEngine` | Physics simulation |

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

Gomobra — [@Gomobra](https://github.com/Gomobra)

Project Link: [https://github.com/Gomobra/PixTui](https://github.com/Gomobra/PixTui)