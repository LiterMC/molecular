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
	// SetObject will be called when block is inited or it's moving between objects
	SetObject(o *Object)
	Mass() float64
	// Material returns the material of the face, nil is allowed
	Material(f Facing) *Material
	// Outline specific the position and the maximum space of the block
	Outline() *Cube
	// Tick will be called when the block need to update it's state
	Tick(dt float64)
}
