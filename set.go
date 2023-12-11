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

package molecular

type set[T comparable] map[T]struct{}

func (s set[T]) Put(v T) {
	s[v] = struct{}{}
}

func (s set[T]) Has(v T) (ok bool) {
	_, ok = s[v]
	return
}

func (s set[T]) Remove(v T) (ok bool) {
	_, ok = s[v]
	delete(s, v)
	return
}

func (s set[T]) Clear() {
	clear(s)
}

func (s set[T]) AsSlice() (a []T) {
	a = make([]T, 0, len(s))
	for i := range s {
		a = append(a, i)
	}
	return
}

func (s set[T]) Clone() (o set[T]) {
	o = make(set[T], len(s))
	for a := range s {
		o.Put(a)
	}
	return
}
