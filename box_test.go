// molecular is a 3D physics engine written in Go
// Copyright (C) 2023  Kevin Z <zyxkad@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package molecular_test

import (
	. "github.com/LiterMC/molecular"
	"testing"
)

var zeroCube = NewCube(ZeroVec, ZeroVec)

func TestBoxOverlap(t *testing.T) {
	type T struct {
		A, B *Cube
		Area *Cube
	}
	datas := []T{
		{zeroCube, zeroCube, zeroCube},
		{NewCube(ZeroVec, OneVec), zeroCube, zeroCube},
		{NewCube(ZeroVec, OneVec), NewCube(ZeroVec, NegOneVec), zeroCube},
		{NewCube(ZeroVec, NegOneVec), zeroCube, NewCube(OneVec, ZeroVec)},
		{NewCube(ZeroVec, NegOneVec), NewCube(ZeroVec, OneVec), NewCube(OneVec, ZeroVec)},
		{NewCube(OneVec, ZeroVec), zeroCube, nil},
		{NewCube(NegOneVec, ZeroVec), zeroCube, nil},
		{NewCube(OneVec, NegOneVec), zeroCube, zeroCube},
		{NewCube(ZeroVec, OneVec.ScaledN(2)), NewCube(OneVec, OneVec), NewCube(OneVec, OneVec)},
		{NewCube(ZeroVec, OneVec.ScaledN(2)), NewCube(OneVec, OneVec.ScaledN(2)), NewCube(OneVec, OneVec)},
		{NewCube(ZeroVec, NegOneVec.ScaledN(2)), NewCube(OneVec, NegOneVec.ScaledN(2)), NewCube(OneVec, OneVec)},
	}
	area := new(Cube)
	for _, d := range datas {
		o := d.A.OverlapBox(d.B, area)
		if o2 := d.A.Overlap(d.B); o != o2 {
			t.Fatalf("Overlap results not synced for cubes %v & %v. OverlapBox=%v, Overlap=%v", d.A, d.B, o, o2)
		} else if o {
			if d.Area == nil {
				t.Errorf("Unexpected overlap area %v for cubes %v & %v", area, d.A, d.B)
			} else if !d.Area.Equals(area) {
				t.Errorf("Incorrect overlap area %v for cubes %v & %v, expect %v", area, d.A, d.B, d.Area)
			}
		} else if d.Area != nil {
			t.Errorf("Expect overlap area for cubes %v & %v, but got none", d.A, d.B)
		}
	}
}
