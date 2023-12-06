package molecular

import (
	"math"
	"sync"
)

var eventWavePool = sync.Pool{
	New: func() any {
		return new(eventWave)
	},
}

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
	e = eventWavePool.Get().(*eventWave)
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
	f.radius += f.speed * dt
	if f.maxRadius >= 0 && f.radius >= f.maxRadius {
		f.radius = f.maxRadius
		f.alive = 0
	}
	f.objsCache = e.appendObjsInsideRing(f.objsCache[:0], f.pos, lastr, f.radius)
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

var maxRs = [...]float64{maxR0, maxR1, maxR2, maxR3, maxR4, maxR5, maxR6, maxR7}

// How does gravity wave works:
// - r < maxRs[n]: update per 1 << n*2 ticks
func (e *Engine) newGravityWave(sender *Object, center Vec3, f *GravityField, tick uint16) *eventWave {
	g := gravityStatusPool.Get().(*gravityStatus)
	g.f = *f
	g.pos = center
	g.gone = false
	g.c.Store(1)
	maxRadius := -1.0
	if e.cfg.MinAccel > 0 {
		maxRadius = math.Sqrt(G / e.cfg.MinAccel * f.Mass())
	}

	life := 0
	for i := (uint16)(2); i != 0 && tick&(i-1) == 0; i <<= 2 {
		life++
	}
	if life < len(maxRs) {
		if maxr := maxRs[life]; maxr < maxRadius {
			maxRadius = maxr
		}
	}

	w := newEventWave(sender, center, maxRadius, func(r *Object) {
		r.Lock()
		defer r.Unlock()

		if last := r.passedGravity[sender]; last != nil {
			last.release()
		}
		g.count()
		r.passedGravity[sender] = g
	}, true)
	w.onRemove = g.release
	w.onBeforeTick = gravityEventBeforeTick
	return w
}

func gravityEventBeforeTick(e *eventWave) bool {
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
