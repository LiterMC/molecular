package molecular

const (
	G = 6.674e-11 // The gravitational constant is 6.674×10−11 N⋅m2/kg2
)

type GravityField struct {
	mass float64
}

func NewGravityField(pos Vec, mass float64) *GravityField {
	return &GravityField{
		mass: mass,
	}
}

func (f *GravityField) Mass() float64 {
	return f.mass
}

func (f *GravityField) SetMass(mass float64) {
	f.mass = mass
}

// FieldAt returns the acceleration at the distance due to the gravity field
func (f *GravityField) FieldAt(distance Vec) Vec {
	l := distance.Len()
	if l == 0 {
		return ZeroVec
	}
	distance.Negate()
	// normalize 1 / l and G * m / l ^ 2
	distance.ScaleN(G * f.mass / (l * l * l))
	return distance
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

func (f *MagnetField) FieldAt(distance Vec) Vec {
	l := distance.Len()
	if l == 0 {
		return ZeroVec
	}
	// normalize will scale with factor 1 / l, so we merge two steps into one
	distance.ScaleN(f.power / (l * l * l))
	return distance
}
