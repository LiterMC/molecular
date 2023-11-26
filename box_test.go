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
