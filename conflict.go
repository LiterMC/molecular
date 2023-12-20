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

func (e *Engine) appendObjsInsideRange(objs []*Object, pos Vec3, radius float64) []*Object {
	radius2 := radius * radius
	for _, o := range e.objects {
		if o.AbsPos().Subbed(pos).SqLen() <= radius2 {
			objs = append(objs, o)
		}
	}
	return objs
}

func (e *Engine) appendObjsInsideRing(objs []*Object, pos Vec3, minR, maxR float64) []*Object {
	minR2 := minR * minR
	maxR2 := maxR * maxR
	for _, o := range e.objects {
		l := o.AbsPos().Subbed(pos).SqLen()
		if minR2 <= l && l <= maxR2 {
			objs = append(objs, o)
		}
	}
	return objs
}
