package molecular

import (
	"math"
)

const (
	C   = 299792458.0 // The speed of light
	cSq = C * C
)

// See <https://en.wikipedia.org/wiki/Lorentz_factor>
func (e *Engine) LorentzFactor(speed float64) float64 {
	return 1 / e.ReLorentzFactor(speed)
}

// ReLorentzFactor is the reciprocal of the Lorentz Factor
// It's used for faster calculate in some specific cases
// See <https://en.wikipedia.org/wiki/Lorentz_factor>
func (e *Engine) ReLorentzFactor(speed float64) float64 {
	return e.ReLorentzFactorSq(speed * speed)
}

// ReLorentzFactorSq is same as ReLorentzFactor, but require squared speed as input
func (e *Engine) ReLorentzFactorSq(speedSq float64) float64 {
	if speedSq == 0 {
		return 1
	}
	if speedSq > e.maxSpeedSq {
		speedSq = e.maxSpeedSq
	}
	return math.Sqrt(1 - speedSq/cSq)
}

// Note: F = dP / dt
func (e *Engine) Momentum(mass float64, velocity Vec3) Vec3 {
	return velocity.ScaledN(e.ReLorentzFactor(velocity.Len()) * mass)
}

// AccFromForce calculate the acceleration from force
// TODO: Not sure if this is correct in SR
func (e *Engine) AccFromForce(mass float64, speed float64, force Vec3) Vec3 {
	return force.ScaledN(e.ReLorentzFactor(speed) / mass)
}

// ProperTime returns the delta time that relative to the moving object,
// with given speed and the delta time relative to the observer (or the server)
func (e *Engine) ProperTime(t float64, speed float64) float64 {
	if t == 0 {
		return 0
	}
	if t < 0 {
		panic("molecular.Engine: delta time cannot be negative")
	}
	if speed < 0 {
		panic("molecular.Engine: speed cannot be negative")
	}
	if speed >= C {
		return math.SmallestNonzeroFloat64
	}
	t2 := t * e.ReLorentzFactor(speed)
	if t2 == 0 {
		return math.SmallestNonzeroFloat64
	}
	return t2
}
