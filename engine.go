package molecular

import (
	"sync"

	"github.com/google/uuid"
)

const (
	defaultMinAcc = 1e-3
)

type Config struct {
	// MinSpeed means the minimum positive speed
	MinSpeed float64
	// MaxSpeed means the maximum positive speed
	MaxSpeed float64
	// MinAccel means the minimum positive acceleration
	MinAccel float64
}

// Engine includes a sync.RWMutex which should be locked when operating global things inside a tick
type Engine struct {
	sync.RWMutex

	// the config should not change while engine running
	cfg                    Config
	minSpeedSq, maxSpeedSq float64
	minAccelSq             float64

	// the main anchor object must be invincible and unmovable
	mainAnchor *Object
	// objects save all the Object instance but not mainAnchor
	objects        map[uuid.UUID]*Object
	events, queued []eventWave
}

func NewEngine(cfg Config) (e *Engine) {
	e = &Engine{
		cfg: cfg,
		mainAnchor: &Object{
			id: uuid.Nil,
		},
		objects: make(map[uuid.UUID]*Object, 10),
	}
	e.maxSpeedSq = cfg.MaxSpeed * cfg.MaxSpeed
	if e.maxSpeedSq <= 0 || e.maxSpeedSq > cSq {
		e.maxSpeedSq = cSq
	}
	e.minSpeedSq = cfg.MinSpeed * cfg.MinSpeed
	if cfg.MinAccel > 0 {
		e.minAccelSq = cfg.MinAccel * cfg.MinAccel
	} else if cfg.MinAccel == 0 {
		e.cfg.MinAccel = defaultMinAcc
		e.minAccelSq = defaultMinAcc * defaultMinAcc
	} else {
		e.minAccelSq = cfg.MinAccel
	}
	return
}

func (e *Engine) Config() Config {
	return e.cfg
}

func (e *Engine) MainAnchor() *Object {
	return e.mainAnchor
}

// NewObject will create an object use v7 UUID
func (e *Engine) NewObject(anchor *Object, pos Vec3) *Object {
	e.Lock()
	defer e.Unlock()

	stat := makeObjStatus()
	stat.anchor = anchor
	stat.pos = pos
	return e.newObjectLocked(stat)
}

func (e *Engine) newObjectLocked(stat objStatus) *Object {
	for i := 20; i > 0; i-- {
		if id, err := uuid.NewV7(); err == nil {
			if _, ok := e.objects[id]; !ok {
				return e.newAndPutObject(id, stat)
			}
		}
	}
	panic("molecular.Engine: Too many UUID generation failures")
}

func (e *Engine) GetObject(id uuid.UUID) *Object {
	e.RLock()
	defer e.RUnlock()

	return e.objects[id]
}

func (e *Engine) Events() []eventWave {
	return e.events
}

func (e *Engine) queueEvent(event eventWave) {
	e.Lock()
	defer e.Unlock()
	e.queued = append(e.queued, event)
}

// Tick will call tick on the main anchor
func (e *Engine) Tick(dt float64) {
	e.Lock()
	e.events = append(e.events, e.queued...)
	e.queued = e.queued[:0]
	e.Unlock()

	var wg sync.WaitGroup
	// tick objects
	e.tickLocked(&wg, dt)
	wg.Wait()

	// save object status
	e.saveStatusLocked(&wg)
	wg.Wait()
}

func (e *Engine) tickLocked(wg *sync.WaitGroup, dt float64) {
	e.RLock()
	defer e.RUnlock()

	wg.Add(len(e.objects))
	for _, o := range e.objects {
		go func(o *Object) {
			defer wg.Done()
			o.tick(dt)
		}(o)
	}
	for _, event := range e.events {
		if event.Heavy() {
			go func(event eventWave) {
				defer wg.Done()
				event.Tick(dt, e)
			}(event)
		}else{
			event.Tick(dt, e)
		}
	}
}

func (e *Engine) saveStatusLocked(wg *sync.WaitGroup) {
	e.Lock()
	defer e.Unlock()

	wg.Add(len(e.objects))
	for _, o := range e.objects {
		go func(o *Object) {
			defer wg.Done()
			o.saveStatus()
		}(o)
	}
	for i := 0; i < len(e.events); {
		event := e.events[i]
		if event.AliveTime() == 0 {
			event.OnRemoved()
			e.events[i] = e.events[len(e.events)-1]
			e.events = e.events[:len(e.events)-1]
		} else {
			i++
		}
	}
}
