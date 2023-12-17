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

var eventWavePool = newObjPool[eventWave]()

type eventWave struct {
	sender            *Object
	pos               Vec3
	alive             float64
	speed             float64
	radius, maxRadius float64
	heavy             bool
	on                func(receiver *Object)
	onBeforeTick      func(*eventWave) bool
	onRemove          func()
	objsCache         []*Object
	delay, tick       int
	skipped           float64
}

func newEventWave(sender *Object, pos Vec3, radius float64, on func(receiver *Object), heavy bool) (e *eventWave) {
	e = eventWavePool.Get()
	e.sender = sender
	e.pos = pos
	e.alive = 60 * 60 // 1 hour
	e.speed = C
	e.radius = 0
	e.maxRadius = radius
	e.on = on
	e.heavy = heavy
	return
}

func (f *eventWave) Sender() *Object {
	return f.sender
}

// Pos returns the absolute start position when the event was sent
func (f *eventWave) Pos() Vec3 {
	return f.pos
}

// If AliveTime returns zero, the event will be removed
func (f *eventWave) AliveTime() float64 {
	return f.alive
}

func (f *eventWave) MaxSpeed() float64 {
	return f.speed
}

func (f *eventWave) MaxRadius() float64 {
	return f.maxRadius
}

// should this event starts from a separate goroutine
func (f *eventWave) Heavy() bool {
	return f.heavy
}

func (f *eventWave) Tick(dt float64, e *Engine) {
	if f.delay > 0 {
		f.skipped += dt
		if f.tick++; f.tick < f.delay {
			return
		}
		dt = f.skipped
		f.skipped = 0
		f.tick = 0
	}
	if f.alive -= dt; f.alive < 0 {
		dt += f.alive
		f.alive = 0
	}
	if f.onBeforeTick != nil && f.onBeforeTick(f) {
		return
	}
	lastr := f.radius
	rd := f.speed * dt
	f.radius += rd
	if f.maxRadius >= 0 && f.radius >= f.maxRadius {
		f.radius = f.maxRadius
		f.alive = 0
	}
	f.objsCache = e.appendObjsInsideRing(f.objsCache[:0], f.pos, lastr, f.radius+rd/2)
	for _, o := range f.objsCache {
		if o == f.sender {
			continue
		}
		f.on(o)
	}
	f.objsCache = f.objsCache[:0]
}

// free will call onRemove and put this eventWave object into the pool
func (f *eventWave) free() {
	f.on = nil
	if f.onRemove != nil {
		f.onRemove()
		f.onRemove = nil
	}
	if f.onBeforeTick != nil {
		f.onBeforeTick = nil
	}
	eventWavePool.Put(f)
}
