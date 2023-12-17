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

 
package molecular

// MaterialProps saves the Material properties
type MaterialProps struct {
	Brittleness float64 // <https://en.wikipedia.org/wiki/Brittleness>
	COR         float64 // Coefficient of restitution <https://en.wikipedia.org/wiki/Coefficient_of_restitution>
	Density     float64 // kg / m^3
	Durability  int64   // -1 means never break
	HeatCap     float64 // J / (kg * K) <https://en.wikipedia.org/wiki/Specific_heat_capacity>
	FirePoint   float64 // The temperature that can cause fire, zero means none
}

// Material represents a unique material.
// You should only handle the pointer of the Material.
// Copying or cloning Material is an illegal operation.
type Material struct {
	id    string
	props MaterialProps
}

func NewMaterial(id string, props MaterialProps) *Material {
	return &Material{
		id:    id,
		props: props,
	}
}

func (m *Material) Id() string {
	return m.id
}

func (m *Material) Props() MaterialProps {
	return m.props
}

// MaterialPair represents the status that between two materials
type MaterialPair struct {
	// The id of the materials
	MatterA, MatterB *Material

	// Frictions see: <https://en.wikipedia.org/wiki/Friction>
	SCOF, KCOF float64 // the coefficients of static/kinetic friction
}

// CalcNetForce returns the net force of a object after canceled out the friction
// The first argument `natural` is the natural force acting on the material
// The second argument `app` is the application force acting on the object
// Note: All input forces **must be** zero or positive, but the net force may be negative
func (p *MaterialPair) CalcNetForce(natural, app float64, moving bool) float64 {
	var friction float64
	if moving {
		friction = p.KCOF * natural
	} else {
		friction = p.SCOF * natural
		if app <= friction {
			return 0
		}
	}
	return app - friction
}

// MaterialSet manages a set of Material and MaterialPair
type MaterialSet struct {
	set map[string]*Material
	// the key of pairs is smaller first
	pairs map[[2]*Material]*MaterialPair
}

func NewMaterialSet() (s *MaterialSet) {
	return &MaterialSet{
		set:   make(map[string]*Material),
		pairs: make(map[[2]*Material]*MaterialPair),
	}
}

// Add push a Material into the MaterialSet
// If the Material's id is already exists, Add will panic
func (s *MaterialSet) Add(m *Material) {
	if m.id == "" {
		panic("molecular.MaterialSet: Material's id cannot be empty")
	}
	if _, ok := s.set[m.id]; ok {
		panic("molecular.MaterialSet: Material " + m.id + " is already exists")
	}
	s.set[m.id] = m
}

func (s *MaterialSet) Get(id string) *Material {
	return s.set[id]
}

func (s *MaterialSet) GetPair(a, b *Material) *MaterialPair {
	var k [2]*Material
	if a.id < b.id {
		k[0] = a
		k[1] = b
	} else {
		k[0] = b
		k[1] = a
	}
	return s.pairs[k]
}

// AddPair push a MaterialPair into the MaterialSet
// If the MaterialPair already exists, AddPair will panic
func (s *MaterialSet) AddPair(pair *MaterialPair) {
	if pair.MatterA.id > pair.MatterB.id {
		pair.MatterA, pair.MatterB = pair.MatterB, pair.MatterA
	}
	k := [2]*Material{pair.MatterA, pair.MatterB}
	if _, ok := s.pairs[k]; ok {
		panic("molecular.MaterialSet: The MaterialPair of materials " +
			pair.MatterA.id + " and " + pair.MatterA.id + " is already exists")
	}
	s.pairs[k] = pair
}
