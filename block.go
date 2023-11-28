package molecular

type Facing uint8

//go:generate stringer -type=Facing
const (
	TOP Facing = iota
	BOTTOM
	LEFT
	RIGHT
	FRONT
	BACK
)

type Block interface {
	Mass() float64
	// Material returns the material of the face, nil is allowed
	Material(f Facing) *Material
	// Outline specific the position and the maximum space of the block
	Outline() *Cube
	// Tick will be called when the block need to update it's state
	Tick(dt float64, o *Object)
}
