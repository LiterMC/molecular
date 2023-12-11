// molecular is a 3D physics engine written in Go
// Copyright (C) 2023  Kevin Z <zyxkad@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
	t.Logf("%v", UnitX.RotatedY(90./180.*math.Pi))
	t.Logf("%v", UnitX.RotatedZ(90./180.*math.Pi))
	t.Logf("%v", UnitY.RotatedX(90./180.*math.Pi))
	t.Logf("%v", UnitY.RotatedZ(-90./180.*math.Pi))
	t.Logf("%v", UnitZ.RotatedX(-90./180.*math.Pi))
	t.Logf("%v", UnitZ.RotatedY(-90./180.*math.Pi))

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
