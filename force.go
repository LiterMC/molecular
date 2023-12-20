// molecular is a 3D physics engine written in Go
// Copyright (C) 2023  Kevin Z <zyxkad@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package molecular

import (
	"math"
)

const (
	G = 6.674e-11 // The gravitational constant is 6.674×10−11 N⋅m2/kg2
)

var gravityFieldPool = newObjPool[GravityField]()

type GravityField struct {
	pos    Vec3
	mass   float64
	radius float64
	rSq    float64 // radius * radius
	rCube  float64 // 1 / (radius * radius * radius)
}

func NewGravityField(pos Vec3, mass float64, radius float64) (f *GravityField) {
	f = gravityFieldPool.Get()
	f.pos = pos
	f.mass = mass
	f.radius = radius
	f.rSq = radius * radius
	f.rCube = 1 / (radius * radius * radius)
	return
}

func (f *GravityField) Pos() Vec3 {
	return f.pos
}

func (f *GravityField) SetPos(pos Vec3) {
	f.pos = pos
}

func (f *GravityField) Mass() float64 {
	return f.mass
}

func (f *GravityField) SetMass(mass float64) {
	f.mass = mass
}

func (f *GravityField) Radius() float64 {
	return f.radius
}

func (f *GravityField) SetRadius(radius float64) {
	f.radius = radius
	f.rSq = radius * radius
	f.rCube = 1 / (radius * radius * radius)
}

func (f *GravityField) Clone() (g *GravityField) {
	g = gravityFieldPool.Get()
	*g = *f
	return
}

// FieldAt returns the acceleration at the position due to the gravity field
func (f *GravityField) FieldAt(pos Vec3) Vec3 {
	acc := f.pos.Subbed(pos)
	lSq := acc.SqLen()
	if lSq == 0 {
		return ZeroVec
	}
	if lSq < f.rSq {
		acc.ScaleN(G * f.mass * f.rCube)
	} else {
		l := math.Sqrt(lSq)
		// normalize 1 / l and G * m / l ^ 2
		acc.ScaleN(G * f.mass / (lSq * l))
	}
	return acc
}

// MagnetField represents a simulated magnetic field.
// For easier calculate, it's not the real magnetic field.
// Since the magnetic field disappears easily, the cubic distance is used
type MagnetField struct {
	power float64 // in m^3 / s^2
}

func NewMagnetField(power float64) *MagnetField {
	return &MagnetField{
		power: power,
	}
}

func (f *MagnetField) Power() float64 {
	return f.power
}

func (f *MagnetField) SetPower(power float64) {
	f.power = power
}

func (f *MagnetField) FieldAt(distance Vec3) Vec3 {
	l := distance.Len()
	if l == 0 {
		return ZeroVec
	}
	// normalize will scale with factor 1 / l, so we merge two steps into one
	distance.ScaleN(f.power / (l * l * l))
	return distance
}
