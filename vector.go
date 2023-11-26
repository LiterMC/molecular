package molecular

import (
	"fmt"
	"math"
)

type Vec struct {
	X, Y, Z float64
}

var (
	ZeroVec   Vec
	OneVec    Vec = Vec{1, 1, 1}
	NegOneVec Vec = Vec{-1, -1, -1}
)

func (v *Vec) Clone() *Vec {
	return &Vec{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}
}

func (v Vec) String() string {
	return fmt.Sprintf("Vec(%v, %v, %v)", v.X, v.Y, v.Z)
}

func (v Vec) XYZ() (x, y, z float64) {
	return v.X, v.Y, v.Z
}

func (v Vec) IsZero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0
}

func (v Vec) Len() float64 {
	return math.Sqrt(v.SqLen())
}

// Squared length
func (v Vec) SqLen() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec) Abs() Vec {
	return Vec{
		X: math.Abs(v.X),
		Y: math.Abs(v.Y),
		Z: math.Abs(v.Z),
	}
}

func (v *Vec) Map(m func(float64) float64) *Vec {
	v.X = m(v.X)
	v.Y = m(v.Y)
	v.Z = m(v.Z)
	return v
}

func (v Vec) Mapped(m func(float64) float64) Vec {
	return Vec{
		X: m(v.X),
		Y: m(v.Y),
		Z: m(v.Z),
	}
}

// Negate is a shortcut of ScaleN(-1)
func (v *Vec) Negate() *Vec {
	v.X = -v.X
	v.Y = -v.Y
	v.Z = -v.Z
	return v
}

// Negated is a shortcut of ScaledN(-1)
func (v Vec) Negated() Vec {
	return Vec{
		X: -v.X,
		Y: -v.Y,
		Z: -v.Z,
	}
}

func (v *Vec) Add(u Vec) *Vec {
	v.X += u.X
	v.Y += u.Y
	v.Z += u.Z
	return v
}

func (v Vec) Added(u Vec) Vec {
	return Vec{
		X: v.X + u.X,
		Y: v.Y + u.Y,
		Z: v.Z + u.Z,
	}
}

func (v *Vec) Sub(u Vec) *Vec {
	v.X -= u.X
	v.Y -= u.Y
	v.Z -= u.Z
	return v
}

func (v Vec) Subbed(u Vec) Vec {
	return Vec{
		X: v.X - u.X,
		Y: v.Y - u.Y,
		Z: v.Z - u.Z,
	}
}

func (v *Vec) Scale(u Vec) *Vec {
	v.X *= u.X
	v.Y *= u.Y
	v.Z *= u.Z
	return v
}

func (v Vec) Scaled(u Vec) Vec {
	return Vec{
		X: v.X * u.X,
		Y: v.Y * u.Y,
		Z: v.Z * u.Z,
	}
}

func (v *Vec) ScaleN(n float64) *Vec {
	v.X *= n
	v.Y *= n
	v.Z *= n
	return v
}

func (v Vec) ScaledN(n float64) Vec {
	return Vec{
		X: v.X * n,
		Y: v.Y * n,
		Z: v.Z * n,
	}
}

// Normalize make the length of the vector to 1 and keep the current direction.
func (v *Vec) Normalize() *Vec {
	if v.IsZero() {
		v.X = 1
	} else {
		v.ScaleN(1 / v.Len())
	}
	return v
}

// Normalized returns a vector of length 1 facing the direction of u with the same angle.
func (v Vec) Normalized() Vec {
	if v.IsZero() {
		return Vec{1, 0, 0}
	}
	return v.ScaledN(1 / v.Len())
}

func (v Vec) Dot(u Vec) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z
}

// AngleX returns the angle between the vector and y-axis, about z-axis
//
//	Z ^
//	  |/
//	--+-->
//	  |  Y
func (v Vec) AngleX() float64 {
	return math.Atan2(v.Z, v.Y)
}

// AngleY returns the angle between the vector and z-axis, about x-axis
//
//	X ^
//	  |/
//	--+-->
//	  |  Z
func (v Vec) AngleY() float64 {
	return math.Atan2(v.X, v.Z)
}

// AngleZ returns the angle between the vector and x-axis, about y-axis
//
//	Y ^
//	  |/
//	--+-->
//	  |  X
func (v Vec) AngleZ() float64 {
	return math.Atan2(v.Y, v.X)
}

// Rotate around x-axis
func (v *Vec) RotateX(angle float64) *Vec {
	s, c := math.Sincos(angle)
	v.Y = v.Y*c - v.Z*s
	v.Z = v.Y*s + v.Z*c
	return v
}

// Rotate around y-axis
func (v *Vec) RotateY(angle float64) *Vec {
	s, c := math.Sincos(angle)
	v.X = v.Y*s + v.Z*c
	v.Z = v.Y*c - v.Z*s
	return v
}

// Rotate around z-axis
func (v *Vec) RotateZ(angle float64) *Vec {
	s, c := math.Sincos(angle)
	v.X = v.X*c - v.Y*s
	v.Y = v.X*s + v.Y*c
	return v
}

// Rotate around x-axis
func (v Vec) RotatedX(angle float64) Vec {
	s, c := math.Sincos(angle)
	return Vec{
		X: v.X,
		Y: v.Y*c - v.Z*s,
		Z: v.Y*s + v.Z*c,
	}
}

// Rotate around y-axis
func (v Vec) RotatedY(angle float64) Vec {
	s, c := math.Sincos(angle)
	return Vec{
		X: v.Y*s + v.Z*c,
		Y: v.Y,
		Z: v.Y*c - v.Z*s,
	}
}

// Rotate around z-axis
func (v Vec) RotatedZ(angle float64) Vec {
	s, c := math.Sincos(angle)
	return Vec{
		X: v.X*c - v.Y*s,
		Y: v.X*s + v.Y*c,
		Z: v.Z,
	}
}

// Volume returns X * Y * Z
func (v Vec) Volume() float64 {
	return v.X * v.Y * v.Z
}
