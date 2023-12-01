package molecular

import (
	"math"
	"sync"
)

type eventWave interface {
	Sender() *Object
	// Pos returns the absolute start position when the event was sent
	Pos() Vec3
	// If AliveTime returns zero, the event will be removed
	AliveTime() float64
	MaxSpeed() float64
	MaxRadius() float64
	// should this event starts from a separate goroutine
	Heavy() bool
	Tick(dt float64, e *Engine)
	OnRemoved()
}

var eventWaveFnPool = sync.Pool{
	New: func() any {
		return new(eventWaveFn)
	},
}

type eventWaveFn struct {
	sender                   *Object
	pos                      Vec3
	alive                    float64
	speed                    float64
	radius, maxRadius        float64
	heavy                    bool
	on                       func(receiver *Object)
	onRemove                 func()
	triggered, lastTriggered set[*Object]
	objsCache                []*Object
}

var _ eventWave = (*eventWaveFn)(nil)

func newEventWave(sender *Object, pos Vec3, radius float64, on func(receiver *Object), heavy bool) (e *eventWaveFn) {
	e = eventWaveFnPool.Get().(*eventWaveFn)
	e.sender = sender
	e.pos = pos
	e.alive = 60 * 60 // 1 hour
	e.speed = C
	e.radius = 0
	e.maxRadius = radius
	e.on = on
	e.heavy = heavy
	if e.triggered == nil {
		e.triggered = make(set[*Object], 10)
		e.lastTriggered = make(set[*Object], 10)
	} else {
		e.triggered.Clear()
	}
	return
}

func (f *eventWaveFn) Sender() *Object {
	return f.sender
}

func (f *eventWaveFn) Pos() Vec3 {
	return f.pos
}

func (f *eventWaveFn) AliveTime() float64 {
	return f.alive
}

func (f *eventWaveFn) MaxSpeed() float64 {
	return f.speed
}

func (f *eventWaveFn) MaxRadius() float64 {
	return f.maxRadius
}

func (f *eventWaveFn) Heavy() bool {
	return f.heavy
}

func (f *eventWaveFn) Tick(dt float64, e *Engine) {
	f.alive -= dt
	if f.alive < 0 {
		dt += f.alive
		f.alive = 0
	}
	lastr := f.radius
	f.radius += f.speed * dt
	if f.maxRadius >= 0 && f.radius >= f.maxRadius {
		f.radius = f.maxRadius
		f.alive = 0
	}
	f.objsCache = e.appendObjsInsideRing(f.objsCache[:0], f.pos, lastr, f.radius)
	f.triggered, f.lastTriggered = f.lastTriggered, f.triggered
	f.triggered.Clear()
	for _, o := range f.objsCache {
		if f.lastTriggered.Has(o) || o == f.sender {
			continue
		}
		f.triggered.Put(o)
		f.on(o)
	}
}

func (f *eventWaveFn) OnRemoved() {
	f.on = nil
	if f.onRemove != nil {
		f.onRemove()
		f.onRemove = nil
	}
	eventWaveFnPool.Put(f)
}

func (e *Engine) newGravityWave(sender *Object, center Vec3, f *GravityField) eventWave {
	g := gravityStatusPool.Get().(*gravityStatus)
	g.f = *f
	g.pos = center
	g.gone = false
	g.c = 1
	maxRadius := -1.0
	if e.cfg.MinAccel > 0 {
		maxRadius = math.Sqrt(G / e.cfg.MinAccel * f.Mass())
	}

	w := newEventWave(sender, center, maxRadius, func(r *Object) {
		r.Lock()
		defer r.Unlock()

		if last := r.passedGravity[sender]; last != nil {
			last.release()
		}
		g.count()
		r.passedGravity[sender] = g
	}, false)
	w.onRemove = g.release
	return w
}
