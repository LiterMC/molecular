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
	o2 := eng.NewObject(NaturalObj, o1, UnitX)
	o2.SetHeading(Vec3{0, 0, 0})
	o3 := eng.NewObject(NaturalObj, o2, UnitY)
	o3.SetHeading(Vec3{0, 0, 0})
	if !o1.AbsPos().Equals(Vec3{1, 0, 0}) {
		t.Errorf("o1 assert failed")
	} else if !o2.AbsPos().Equals(Vec3{1, 0, 1}) {
		t.Errorf("o2 assert failed")
	} else if !o3.AbsPos().Equals(Vec3{1, 1, 1}) {
		t.Errorf("o3 assert failed")
	}
}
