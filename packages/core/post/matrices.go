// Package post provides post-processing pipeline (filters, matrices, effects)
// for the PixTui port. It mirrors the TypeScript implementation found in
// packages/core/src/post/ of the original repository.
//
// To avoid an import cycle (core -> post -> core), the post package does not
// import core. Instead it operates on the PostEffectBuffer interface (defined
// in effects.go), which is satisfied by *core.NativeBuffer.
package post

import (
	"math"
)

// Matrix is a 4x4 float64 matrix. It can represent 2D affine transforms
// (using the top-left 2x2 plus translation column) as well as 4x4 RGBA color
// matrices used by color filters.
type Matrix [4][4]float64

// NewMatrix returns a zero-initialized 4x4 matrix.
func NewMatrix() Matrix {
	return Matrix{}
}

// Identity returns the 4x4 identity matrix.
func Identity() Matrix {
	return Matrix{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

// Translate returns a 2D translation matrix.
func Translate(x, y float64) Matrix {
	m := Identity()
	m[0][3] = x
	m[1][3] = y
	return m
}

// Scale returns a 2D scaling matrix.
func Scale(sx, sy float64) Matrix {
	m := Identity()
	m[0][0] = sx
	m[1][1] = sy
	return m
}

// Rotate returns a 2D rotation matrix (angle in radians).
func Rotate(angle float64) Matrix {
	m := Identity()
	c := math.Cos(angle)
	s := math.Sin(angle)
	m[0][0] = c
	m[0][1] = -s
	m[1][0] = s
	m[1][1] = c
	return m
}

// Multiply returns the matrix product a*b.
func Multiply(a, b Matrix) Matrix {
	var result Matrix
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return result
}

// TransformPoint transforms a 2D point (x, y) by the matrix m and returns the
// transformed coordinates. Homogeneous coordinate w is assumed to be 1.
func TransformPoint(m Matrix, x, y float64) (float64, float64) {
	tx := m[0][0]*x + m[0][1]*y + m[0][3]
	ty := m[1][0]*x + m[1][1]*y + m[1][3]
	return tx, ty
}

// ApplyMatrix applies a 2D affine matrix transform to the buffer cells.
// Cells are remapped by inverse-mapping each destination cell to a source
// cell; out-of-bounds sources are left as blank space.
func ApplyMatrix(buffer PostEffectBuffer, m Matrix) {
	width := buffer.Width()
	height := buffer.Height()
	if width == 0 || height == 0 {
		return
	}

	access := buffer.Access()
	chars := access.Char
	fgs := access.Fg
	bgs := access.Bg
	attrs := access.Attributes

	// Blank out every destination cell first.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			buffer.SetCell(x, y, ' ', 0, 0, 0)
		}
	}

	inv := invertAffine(m)
	if inv == nil {
		return
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sx, sy := TransformPoint(*inv, float64(x), float64(y))
			sxR := int(math.Round(sx))
			syR := int(math.Round(sy))
			if sxR < 0 || sxR >= width || syR < 0 || syR >= height {
				continue
			}
			srcIdx := syR*width + sxR
			buffer.SetCell(x, y, chars[srcIdx], fgs[srcIdx], bgs[srcIdx], attrs[srcIdx])
		}
	}
}

// invertAffine returns the inverse of the 2D affine transform encoded in m,
// or nil if the matrix is singular (non-invertible).
func invertAffine(m Matrix) *Matrix {
	a := m[0][0]
	b := m[0][1]
	c := m[1][0]
	d := m[1][1]
	det := a*d - b*c
	if math.Abs(det) < 1e-12 {
		return nil
	}
	invDet := 1.0 / det
	tx := m[0][3]
	ty := m[1][3]

	result := Identity()
	result[0][0] = d * invDet
	result[0][1] = -b * invDet
	result[1][0] = -c * invDet
	result[1][1] = a * invDet
	result[0][3] = -(result[0][0]*tx + result[0][1]*ty)
	result[1][3] = -(result[1][0]*tx + result[1][1]*ty)
	return &result
}
