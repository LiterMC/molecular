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
	"encoding/binary"
)

type Bitset struct {
	data []uint32
}

// NewBitset creates a Bitset with at least `n` bits
func NewBitset(n int) *Bitset {
	return &Bitset{
		data: make([]uint32, (n+31)/32),
	}
}

// Cap returns the bit slots of the Bitset
func (b *Bitset) Cap() int {
	return len(b.data) * 32
}

// Get will return true if the target bit is one
func (b *Bitset) Get(i int) bool {
	n := i % 32
	i /= 32
	if i >= len(b.data) {
		return false
	}
	return b.data[i]&(1<<n) != 0
}

// Set set the bit to one
func (b *Bitset) Set(i int) {
	n := i % 32
	i /= 32
	if i >= len(b.data) {
		b.data = growToLen(b.data, i+1)
	}
	b.data[i] |= 1 << n
}

// Clear set the bit to zero
func (b *Bitset) Clear(i int) {
	n := i % 32
	i /= 32
	if i >= len(b.data) {
		b.data = growToLen(b.data, i+1)
	}
	b.data[i] &^= 1 << n
}

// Flip toggle the bit, and returns the old value
func (b *Bitset) Flip(i int) bool {
	n := i % 32
	i /= 32
	if i >= len(b.data) {
		b.data = growToLen(b.data, i+1)
	} else if b.data[i]&(1<<n) != 0 {
		b.data[i] &^= 1 << n
		return true
	}
	b.data[i] |= 1 << n
	return false
}

func (b *Bitset) String() string {
	const prefix = "bitset:"
	buf := make([]byte, len(prefix)+len(b.data)*32)
	copy(buf, prefix)
	for i, v := range b.data {
		for j := 0; j < 32; j++ {
			n := i*32 + j + len(prefix)
			if v&(1<<j) == 0 {
				buf[n] = '0'
			} else {
				buf[n] = '1'
			}
		}
	}
	return (string)(buf)
}

// Bytes encode the Bitset to bytes use LittleEndian mode
func (b *Bitset) Bytes() (buf []byte) {
	buf = make([]byte, len(b.data)*4)
	for i, v := range b.data {
		binary.LittleEndian.PutUint32(buf[i*4:], v)
	}
	return
}
