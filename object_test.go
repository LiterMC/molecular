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
	"testing"

	. "github.com/LiterMC/molecular"
)

var eng = NewEngine(Config{})

func TestObjectAbsPosRotate(t *testing.T) {
	o1 := eng.NewObject(NaturalObj, nil, UnitX)
	o1.SetHeading(Vec3{math.Pi, 0, 0})
	o2 := eng.NewObject(NaturalObj, o1, UnitZ)
	o2.SetHeading(Vec3{0, 0, 0})
	o3 := eng.NewObject(NaturalObj, o2, UnitY)
	o3.SetHeading(Vec3{0, 0, 0})
	if !o1.AbsPos().Equals(Vec3{1, 0, 0}) {
		t.Errorf("o1 assert failed: %v", o1.AbsPos())
	} else if !o2.AbsPos().Equals(Vec3{1, 0, 1}) {
		t.Errorf("o2 assert failed: %v", o2.AbsPos())
	} else if !o3.AbsPos().Equals(Vec3{1, 1, 1}) {
		t.Errorf("o3 assert failed: %v", o3.AbsPos())
	}
}
