package molecular

import (
	"math"
	"sync"
)

const (
	G = 6.674e-11 // The gravitational constant is 6.674×10−11 N⋅m2/kg2
)

type GravityField struct {
	mass    float64
	radius, rSq  float64
}

func NewGravityField(mass float64, radius float64) *GravityField {
	return &GravityField{
		mass:     mass,
		radius:   radius,
	}
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
}

// FieldAt returns the acceleration at the distance due to the gravity field
func (f *GravityField) FieldAt(distance Vec3) Vec3 {
	lSq := distance.SqLen()
	if lSq == 0 {
		return ZeroVec
	}
	distance.Negate()
	if lSq < f.rSq {
		distance.ScaleN(G * f.mass / lSq)
	} else {
		l := math.Sqrt(lSq)
		// normalize 1 / l and G * m / l ^ 2
		distance.ScaleN(G * f.mass / (lSq * l))
	}
	return distance
}

type gravityStatus struct {
	f   GravityField
	pos Vec3

	gone bool
	c    int
}

var gravityStatusPool = sync.Pool{
	New: func() any {
		return new(gravityStatus)
	},
}

func (s *gravityStatus) FieldAt(pos Vec3) Vec3 {
	return s.f.FieldAt(pos.Subbed(s.pos))
}

func (s *gravityStatus) clone() (g *gravityStatus) {
	g = gravityStatusPool.Get().(*gravityStatus)
	*g = *s
	g.c = 1
	return
}

func (s *gravityStatus) count() {
	s.c++
}

func (s *gravityStatus) release() {
	s.c--
	if s.c < 0 {
		panic("gravityStatus.count become less than one")
	}
	if s.c == 0 {
		gravityStatusPool.Put(s)
	}
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
