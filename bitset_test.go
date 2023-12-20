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

func TestBitset(t *testing.T) {
	b := NewBitset(0)
	b.Set(1)
	if b.Get(0) {
		t.Errorf("Bitset(1).Get(0) is true")
	}
	if b.Get(33) {
		t.Errorf("Bitset(1).Get(33) is true")
	}
	if b.Get(32) {
		t.Errorf("Bitset(1).Get(32) is true")
	}
	if b.Get(31) {
		t.Errorf("Bitset(1).Get(31) is true")
	}
	if !b.Get(1) {
		t.Errorf("Bitset(1).Get(1) is false")
	}
	if b.Bytes()[0] != 0x02 {
		t.Errorf("Bitset(1).Bytes()[0] != 0x02")
	}
	if b.Bytes()[3] != 0x00 {
		t.Errorf("Bitset(1).Bytes()[3] != 0x00")
	}
	b.Set(32)
	if b.Get(0) {
		t.Errorf("Bitset(1).Get(0) is true")
	}
	if b.Get(33) {
		t.Errorf("Bitset(1).Get(33) is true")
	}
	if !b.Get(32) {
		t.Errorf("Bitset(1).Get(32) is false")
	}
	if b.Get(31) {
		t.Errorf("Bitset(1).Get(31) is true")
	}
	if !b.Get(1) {
		t.Errorf("Bitset(1).Get(1) is false")
	}
	b.Clear(1)
	if b.Get(1) {
		t.Errorf("Bitset(1).Get(1) is true")
	}
}
