package molecular_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	. "github.com/LiterMC/molecular"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func randPi() float64 {
	return (r.Float64() - 0.5) * 2 * math.Pi
}

func randAngle() Vec3 {
	return Vec3{randPi(), randPi(), randPi()}
}

func randVec3() Vec3 {
	return Vec3{
		(r.Float64() - 0.5) * 4,
		(r.Float64() - 0.5) * 4,
		(r.Float64() - 0.5) * 4,
	}
}

func TestVectorRotateXYZ(t *testing.T) {
	t.Logf("%v", UnitX.RotatedY(90. / 180. * math.Pi))
	t.Logf("%v", UnitX.RotatedZ(90. / 180. * math.Pi))
	t.Logf("%v", UnitY.RotatedX(90. / 180. * math.Pi))
	t.Logf("%v", UnitY.RotatedZ(-90. / 180. * math.Pi))
	t.Logf("%v", UnitZ.RotatedX(-90. / 180. * math.Pi))
	t.Logf("%v", UnitZ.RotatedY(-90. / 180. * math.Pi))

	pos := make([]Vec3, 16)
	for i, _ := range pos {
		pos[i] = randVec3()
	}
	for i := 0; i < 16; i++ {
		a := randAngle()
		for _, p := range pos {
			q := p
			q.RotateX(a.X).RotateY(a.Y).RotateZ(a.Z)
			o := p.RotatedXYZ(a)
			if !o.Equals(q) {
				t.Errorf("Rotated pos not equal: %v => (%v, %v)", p, q, o)
			}
		}
	}
}
