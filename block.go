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
	Material(f Facing) *Material
	Outline() *Cube
	Tick(dt float64)
}
