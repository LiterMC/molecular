package molecular

import (
	"fmt"
	"math"
)

type Vec3 struct {
	X, Y, Z float64
}

var (
	UnitX     Vec3 = Vec3{1, 0, 0}
	UnitY     Vec3 = Vec3{0, 1, 0}
	UnitZ     Vec3 = Vec3{0, 0, 1}
	ZeroVec   Vec3
	OneVec    Vec3 = Vec3{1, 1, 1}
	NegOneVec Vec3 = Vec3{-1, -1, -1}
)

func (v *Vec3) Clone() *Vec3 {
	return &Vec3{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}
}

func (v Vec3) String() string {
	return fmt.Sprintf("Vec3(%v, %v, %v)", v.X, v.Y, v.Z)
}

func (v Vec3) XYZ() (x, y, z float64) {
	return v.X, v.Y, v.Z
}

func (v Vec3) IsZero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0
}

func (v Vec3) Equals(u Vec3) bool {
	return v.X == u.X && v.Y == u.Y && v.Z == u.Z
}

func (v Vec3) Len() float64 {
	return math.Sqrt(v.SqLen())
}

// Squared length
func (v Vec3) SqLen() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3) Abs() Vec3 {
	return Vec3{
		X: math.Abs(v.X),
		Y: math.Abs(v.Y),
		Z: math.Abs(v.Z),
	}
}

func (v *Vec3) Map(m func(float64) float64) *Vec3 {
	v.X = m(v.X)
	v.Y = m(v.Y)
	v.Z = m(v.Z)
	return v
}

func (v Vec3) Mapped(m func(float64) float64) Vec3 {
	return Vec3{
		X: m(v.X),
		Y: m(v.Y),
		Z: m(v.Z),
	}
}

// Negate is a shortcut of ScaleN(-1)
func (v *Vec3) Negate() *Vec3 {
	v.X = -v.X
	v.Y = -v.Y
	v.Z = -v.Z
	return v
}

// Negated is a shortcut of ScaledN(-1)
func (v Vec3) Negated() Vec3 {
	return Vec3{
		X: -v.X,
		Y: -v.Y,
		Z: -v.Z,
	}
}

func (v *Vec3) Add(u Vec3) *Vec3 {
	v.X += u.X
	v.Y += u.Y
	v.Z += u.Z
	return v
}

func (v Vec3) Added(u Vec3) Vec3 {
	return Vec3{
		X: v.X + u.X,
		Y: v.Y + u.Y,
		Z: v.Z + u.Z,
	}
}

func (v *Vec3) Sub(u Vec3) *Vec3 {
	v.X -= u.X
	v.Y -= u.Y
	v.Z -= u.Z
	return v
}

func (v Vec3) Subbed(u Vec3) Vec3 {
	return Vec3{
		X: v.X - u.X,
		Y: v.Y - u.Y,
		Z: v.Z - u.Z,
	}
}

func (v *Vec3) Scale(u Vec3) *Vec3 {
	v.X *= u.X
	v.Y *= u.Y
	v.Z *= u.Z
	return v
}

func (v Vec3) Scaled(u Vec3) Vec3 {
	return Vec3{
		X: v.X * u.X,
		Y: v.Y * u.Y,
		Z: v.Z * u.Z,
	}
}

func (v *Vec3) ScaleN(n float64) *Vec3 {
	v.X *= n
	v.Y *= n
	v.Z *= n
	return v
}

func (v Vec3) ScaledN(n float64) Vec3 {
	return Vec3{
		X: v.X * n,
		Y: v.Y * n,
		Z: v.Z * n,
	}
}

func (v *Vec3) Mod(u Vec3) *Vec3 {
	v.X = math.Mod(v.X, u.X)
	v.Y = math.Mod(v.Y, u.Y)
	v.Z = math.Mod(v.Z, u.Z)
	return v
}

func (v Vec3) Moded(u Vec3) Vec3 {
	return Vec3{
		X: math.Mod(v.X, u.X),
		Y: math.Mod(v.Y, u.Y),
		Z: math.Mod(v.Z, u.Z),
	}
}

func (v *Vec3) ModN(n float64) *Vec3 {
	v.X = math.Mod(v.X, n)
	v.Y = math.Mod(v.Y, n)
	v.Z = math.Mod(v.Z, n)
	return v
}

func (v Vec3) ModedN(n float64) Vec3 {
	return Vec3{
		X: math.Mod(v.X, n),
		Y: math.Mod(v.Y, n),
		Z: math.Mod(v.Z, n),
	}
}

// Normalize make the length of the vector to 1 and keep the current direction.
func (v *Vec3) Normalize() *Vec3 {
	if v.IsZero() {
		v.X = 1
	} else {
		v.ScaleN(1 / v.Len())
	}
	return v
}

// Normalized returns a vector of length 1 facing the direction of u with the same angle.
func (v Vec3) Normalized() Vec3 {
	if v.IsZero() {
		return Vec3{1, 0, 0}
	}
	return v.ScaledN(1 / v.Len())
}

func (v Vec3) Dot(u Vec3) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z
}

// AngleX returns the angle between the vector and y-axis, about z-axis
//
//	Z ^
//	  |/
//	--+-->
//	  |  Y
func (v Vec3) AngleX() float64 {
	return math.Atan2(v.Z, v.Y)
}

// AngleY returns the angle between the vector and z-axis, about x-axis
//
//	X ^
//	  |/
//	--+-->
//	  |  Z
func (v Vec3) AngleY() float64 {
	return math.Atan2(v.X, v.Z)
}

// AngleZ returns the angle between the vector and x-axis, about y-axis
//
//	Y ^
//	  |/
//	--+-->
//	  |  X
func (v Vec3) AngleZ() float64 {
	return math.Atan2(v.Y, v.X)
}

// Rotate around x-axis
func (v *Vec3) RotateX(angle float64) *Vec3 {
	s, c := math.Sincos(angle)
	v.Y, v.Z = v.Y*c-v.Z*s, v.Y*s+v.Z*c
	return v
}

// Rotate around y-axis
func (v *Vec3) RotateY(angle float64) *Vec3 {
	s, c := math.Sincos(angle)
	v.X, v.Z = v.X*s+v.Z*c, v.X*c-v.Z*s
	return v
}

// Rotate around z-axis
func (v *Vec3) RotateZ(angle float64) *Vec3 {
	s, c := math.Sincos(angle)
	v.X, v.Y = v.X*c-v.Y*s, v.X*s+v.Y*c
	return v
}

// TODO: maybe we can do them once?
func (v *Vec3) RotateXYZ(angles Vec3) *Vec3 {
	return v.
		RotateX(angles.X).
		RotateY(angles.Y).
		RotateZ(angles.Z)
}

// Rotate around x-axis
func (v Vec3) RotatedX(angle float64) Vec3 {
	s, c := math.Sincos(angle)
	return Vec3{
		X: v.X,
		Y: v.Y*c - v.Z*s,
		Z: v.Y*s + v.Z*c,
	}
}

// Rotate around y-axis
func (v Vec3) RotatedY(angle float64) Vec3 {
	s, c := math.Sincos(angle)
	return Vec3{
		X: v.X*s + v.Z*c,
		Y: v.Y,
		Z: v.X*c - v.Z*s,
	}
}

// Rotate around z-axis
func (v Vec3) RotatedZ(angle float64) Vec3 {
	s, c := math.Sincos(angle)
	return Vec3{
		X: v.X*c - v.Y*s,
		Y: v.X*s + v.Y*c,
		Z: v.Z,
	}
}

func (v Vec3) RotatedXYZ(angles Vec3) Vec3 {
	w := v
	w.RotateXYZ(angles)
	return w
}

type Vec4 struct {
	T, X, Y, Z float64
}

func (v Vec4) String() string {
	return fmt.Sprintf("Vec4(%v, %v, %v, %v)", v.T, v.X, v.Y, v.Z)
}

func (v Vec4) XYZ() (x, y, z float64) {
	return v.X, v.Y, v.Z
}

func (v Vec4) IsZero() bool {
	return v.T == 0 && v.X == 0 && v.Y == 0 && v.Z == 0
}

func (v Vec4) Equals(u Vec4) bool {
	return v.T == u.T && v.X == u.X && v.Y == u.Y && v.Z == u.Z
}

func (v Vec4) Len() float64 {
	return math.Sqrt(v.SqLen())
}

// Squared length
func (v Vec4) SqLen() float64 {
	return v.T*v.T + v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec4) To3() Vec3 {
	return Vec3{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}
}

func (v Vec4) Abs() Vec4 {
	return Vec4{
		T: math.Abs(v.T),
		X: math.Abs(v.X),
		Y: math.Abs(v.Y),
		Z: math.Abs(v.Z),
	}
}

func (v *Vec4) Map(m func(float64) float64) *Vec4 {
	v.T = m(v.T)
	v.X = m(v.X)
	v.Y = m(v.Y)
	v.Z = m(v.Z)
	return v
}

func (v Vec4) Mapped(m func(float64) float64) Vec4 {
	return Vec4{
		T: m(v.T),
		X: m(v.X),
		Y: m(v.Y),
		Z: m(v.Z),
	}
}

// Negate is a shortcut of ScaleN(-1)
func (v *Vec4) Negate() *Vec4 {
	v.T = -v.T
	v.X = -v.X
	v.Y = -v.Y
	v.Z = -v.Z
	return v
}

// Negated is a shortcut of ScaledN(-1)
func (v Vec4) Negated() Vec4 {
	return Vec4{
		T: -v.T,
		X: -v.X,
		Y: -v.Y,
		Z: -v.Z,
	}
}

func (v *Vec4) Add(u Vec4) *Vec4 {
	v.T += u.T
	v.X += u.X
	v.Y += u.Y
	v.Z += u.Z
	return v
}

func (v Vec4) Added(u Vec4) Vec4 {
	return Vec4{
		T: v.T + u.T,
		X: v.X + u.X,
		Y: v.Y + u.Y,
		Z: v.Z + u.Z,
	}
}

func (v *Vec4) Sub(u Vec4) *Vec4 {
	v.T -= u.T
	v.X -= u.X
	v.Y -= u.Y
	v.Z -= u.Z
	return v
}

func (v Vec4) Subbed(u Vec4) Vec4 {
	return Vec4{
		T: v.T - u.T,
		X: v.X - u.X,
		Y: v.Y - u.Y,
		Z: v.Z - u.Z,
	}
}

func (v *Vec4) Scale(u Vec4) *Vec4 {
	v.T *= u.T
	v.X *= u.X
	v.Y *= u.Y
	v.Z *= u.Z
	return v
}

func (v Vec4) Scaled(u Vec4) Vec4 {
	return Vec4{
		T: v.T * u.T,
		X: v.X * u.X,
		Y: v.Y * u.Y,
		Z: v.Z * u.Z,
	}
}

func (v *Vec4) ScaleN(n float64) *Vec4 {
	v.T *= n
	v.X *= n
	v.Y *= n
	v.Z *= n
	return v
}

func (v Vec4) ScaledN(n float64) Vec4 {
	return Vec4{
		T: v.T * n,
		X: v.X * n,
		Y: v.Y * n,
		Z: v.Z * n,
	}
}
