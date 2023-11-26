package molecular

import (
	"math"
)

const (
	C = 299792458.0 // The speed of light
)

// RelativeDeltaTime returns the actual delta time that relative to the moving object,
// with given speed and the delta time relative to the observer (or the server)
func (e *Engine) RelativeDeltaTime(t float64, speed float64) float64 {
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
	t2 := t * math.Sqrt(1-speed*speed/(C*C))
	if t2 == 0 {
		return math.SmallestNonzeroFloat64
	}
	return t2
}
