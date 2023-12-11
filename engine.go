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
	// TODO: should we use tree/map structure instead of flat?
	objects map[uuid.UUID]*Object
	events  []*eventWave
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

// NewObject will create an object use random v7 UUID
func (e *Engine) NewObject(typ ObjType, anchor *Object, pos Vec3, processors ...func(*Object)) (o *Object) {
	stat := makeObjStatus()
	stat.anchor = anchor
	stat.pos = pos

	e.Lock()
	defer e.Unlock()

	id := e.generateObjectId()
	o = e.newAndPutObject(id, stat)
	o.SetType(typ)
	for _, p := range processors {
		p(o)
	}
	return
}

func (e *Engine) newObjectFromStatus(id uuid.UUID, stat objStatus, processors ...func(*Object)) (o *Object) {
	e.Lock()
	defer e.Unlock()

	o = e.newAndPutObject(id, stat)
	for _, p := range processors {
		p(o)
	}
	return
}

func (e *Engine) generateObjectId() uuid.UUID {
	for i := 20; i > 0; i-- {
		if id, err := uuid.NewV7(); err == nil {
			if _, ok := e.objects[id]; !ok {
				return id
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

func (e *Engine) ForeachObject(cb func(o *Object)) {
	e.RLock()
	defer e.RUnlock()

	for _, o := range e.objects {
		cb(o)
	}
}

func (e *Engine) ForeachBlock(cb func(b Block)) {
	e.ForeachObject(func(o *Object) {
		for _, b := range o.blocks {
			cb(b)
		}
	})
}

// Events returns the length of event waves
func (e *Engine) Events() int {
	return len(e.events)
}

func (e *Engine) queueEvent(event *eventWave) {
	if event == nil {
		return
	}

	e.Lock()
	defer e.Unlock()
	e.events = append(e.events, event)
}

// Tick will call tick on the main anchor
func (e *Engine) Tick(dt float64) {
	var wg sync.WaitGroup
	// tick objects
	e.tickObjectLocked(&wg, dt)
	wg.Wait()

	// tick events
	e.tickEventLocked(&wg, dt)
	wg.Wait()

	// sync object status
	e.syncStatusLocked(&wg)
	wg.Wait()
}

func (e *Engine) tickObjectLocked(wg *sync.WaitGroup, dt float64) {
	e.RLock()
	defer e.RUnlock()

	for _, o := range e.objects {
		wg.Add(1)
		go func(o *Object) {
			defer wg.Done()
			o.tick(dt)
		}(o)
	}
}

func (e *Engine) tickEventLocked(wg *sync.WaitGroup, dt float64) {
	e.RLock()
	defer e.RUnlock()

	for _, event := range e.events {
		if event.Heavy() {
			wg.Add(1)
			go func(event *eventWave) {
				defer wg.Done()
				event.Tick(dt, e)
			}(event)
		} else {
			event.Tick(dt, e)
		}
	}
}

func (e *Engine) syncStatusLocked(wg *sync.WaitGroup) {
	e.Lock()
	defer e.Unlock()

	for _, o := range e.objects {
		wg.Add(1)
		go func(o *Object) {
			defer wg.Done()
			o.saveStatus()
		}(o)
	}
	// remove not alive events
	for i := 0; i < len(e.events); {
		event := e.events[i]
		if event.AliveTime() == 0 {
			event.free()
			e.events[i] = e.events[len(e.events)-1]
			e.events = e.events[:len(e.events)-1]
		} else {
			i++
		}
	}
}
