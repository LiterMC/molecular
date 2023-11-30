package molecular

type eventWave interface {
	Sender() *Object
	// Pos returns the absolute start position when the event was sent
	Pos() Vec3
	// If AliveTime returns zero, the event will be removed
	AliveTime() float64
	MaxSpeed() float64
	MaxRadius() float64
	Tick(dt float64, e *Engine)
}

type eventWaveFn struct {
	sender                   *Object
	pos                      Vec3
	alive                    float64
	speed                    float64
	radius, maxRadius        float64
	on                       func(receiver *Object)
	triggered, lastTriggered set[*Object]
	objsCache                []*Object
}

var _ eventWave = (*eventWaveFn)(nil)

func newEventWave(sender *Object, pos Vec3, radius float64, on func(receiver *Object)) eventWave {
	return &eventWaveFn{
		sender:        sender,
		pos:           pos,
		alive:         60 * 60, // 1 hour
		speed:         C,
		maxRadius:     radius,
		on:            on,
		triggered:     make(set[*Object], 10),
		lastTriggered: make(set[*Object], 10),
	}
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
		if f.lastTriggered.Has(o) {
			continue
		}
		f.triggered.Put(o)
		f.on(o)
	}
}
