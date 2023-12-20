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

import (
	"sync"
)

// objPool wrapped a sync.Pool
type objPool[T any] struct {
	pool sync.Pool // internal sync pool
}

func newObjPool[T any]() (p *objPool[T]) {
	return &objPool[T]{
		pool: sync.Pool{
			New: func() any {
				return new(T)
			},
		},
	}
}

func (p *objPool[T]) Get() (ptr *T) {
	ptr = p.pool.Get().(*T)
	return
}

func (p *objPool[T]) Put(ptr *T) {
	p.pool.Put(ptr)
}

// growToLen will set the slice's length to n
// if the slice's cap is less than n, it will create a new slice and copy the data
func growToLen[T any](slice []T, n int) []T {
	diff := cap(slice) - n
	if diff >= 0 {
		return slice[:n]
	}
	res := make([]T, n, n+5)
	copy(res, slice)
	return res
}
