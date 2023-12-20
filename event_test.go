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
	"math"
	"testing"
)

const (
	maxR0 = (float64)(int(1)<<(iota*2)) * (C / 100.)
	maxR1
	maxR2
	maxR3
	maxR4
	maxR5
	maxR6
	maxR7
	maxR8
)

type eventWaveFn struct {
	radius float64
	delay  int
}

func beforeTickSwitch(e *eventWaveFn) bool {
	switch {
	case e.radius > maxR8:
		if e.delay != 1<<(8*2) {
			e.delay = 1 << (8 * 2)
		}
	case e.radius > maxR7:
		if e.delay != 1<<(7*2) {
			e.delay = 1 << (7 * 2)
		}
	case e.radius > maxR6:
		if e.delay != 1<<(6*2) {
			e.delay = 1 << (6 * 2)
		}
	case e.radius > maxR5:
		if e.delay != 1<<(5*2) {
			e.delay = 1 << (5 * 2)
		}
	case e.radius > maxR4:
		if e.delay != 1<<(4*2) {
			e.delay = 1 << (4 * 2)
		}
	case e.radius > maxR3:
		if e.delay != 1<<(3*2) {
			e.delay = 1 << (3 * 2)
		}
	case e.radius > maxR2:
		if e.delay != 1<<(2*2) {
			e.delay = 1 << (2 * 2)
		}
	case e.radius > maxR1:
		if e.delay != 1<<(1*2) {
			e.delay = 1 << (1 * 2)
		}
	case e.radius > maxR0:
		if e.delay != 1<<(0*2) {
			e.delay = 1 << (0 * 2)
		}
	}
	return false
}

func beforeTickMath(e *eventWaveFn) bool {
	if e.radius >= maxR8 {
		if e.delay != 1<<(8*2) {
			e.delay = 1 << (8 * 2)
		}
	} else {
		n := (int)(math.Log2(e.radius / (C / 100.)))
		if n > 0 {
			if e.delay != 1<<n {
				e.delay = 1 << n
			}
		}
	}
	return false
}

func BenchmarkEventWaveRadiusCheckSwitch(b *testing.B) {
	events := make([]*eventWaveFn, 10)
	for i, _ := range events {
		events[i] = &eventWaveFn{
			radius: r.Float64() * C * 2,
			delay:  0,
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, e := range events {
			beforeTickSwitch(e)
		}
	}
}

func BenchmarkEventWaveRadiusCheckMath(b *testing.B) {
	events := make([]*eventWaveFn, 10)
	for i, _ := range events {
		events[i] = &eventWaveFn{
			radius: r.Float64() * C * 2,
			delay:  0,
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, e := range events {
			beforeTickMath(e)
		}
	}
}
