package molecular

import (
	"math"
)

const (
	C = 299792458.0 // The speed of light
)

// RelativeDeltaTime returns the actuall time that relative to the obeserver (server),
// with given speed and the time relative to the moving object
func RelativeDeltaTime(t float64, speed float64) float64 {
	return t / math.Sqrt(1-speed*speed/C*C)
}
