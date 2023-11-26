package molecular

import (
	"math"
)

const (
	G = 6.674e-11   // The gravitational constant is 6.674×10−11 N⋅m2/kg2
	C = 299792458.0 // The speed of light
)

func RelativeDeltaTime(t float64, speed float64) float64 {
	return t / math.Sqrt(1-speed*speed/C*C)
}

type ForceField struct {
	pos  Vec
	mass float64
}

func (f *ForceField) Mass() float64 {
	return f.mass
}

func (f *ForceField) SetMass(mass float64) {
	f.mass = mass
}

// FieldAt returns the acceleration at the pos due to the force field
func (f *ForceField) FieldAt(pos Vec) Vec {
	acc := f.pos.Subbed(pos)
	l := acc.Len()
	if l == 0 {
		return ZeroVec
	}
	// normalize 1 / l and G * m / l ^ 2
	acc.ScaleN(G * f.mass / (l * l * l))
	return acc
}
