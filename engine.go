package molecular

import (
	"sync"

	"github.com/google/uuid"
)

type Config struct {
	// MinSpeed means the minimum positive speed
	MinSpeed float64
	// MaxSpeed means the maximum positive speed
	MaxSpeed float64
}

// Engine includes a sync.RWMutex which should be locked when operating global things inside a tick
type Engine struct {
	sync.RWMutex

	// the config should not change while engine running
	cfg Config
	// the main anchor object must be invincible and unmovable
	mainAnchor *Object
	objects    map[uuid.UUID]*Object
}

func NewEngine(cfg Config) (e *Engine) {
	e = &Engine{
		cfg: cfg,
		mainAnchor: &Object{
			id:      uuid.Nil,
			attachs: make(set[*Object], 10),
		},
		objects: make(map[uuid.UUID]*Object, 10),
	}
	e.objects[uuid.Nil] = e.mainAnchor
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

	return e.newObjectLocked(anchor, pos)
}

func (e *Engine) newObjectLocked(anchor *Object, pos Vec3) *Object {
	for i := 20; i > 0; i-- {
		if id, err := uuid.NewV7(); err == nil {
			if _, ok := e.objects[id]; !ok {
				return e.newAndPutObject(id, anchor, pos)
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

// Tick will call tick on the main anchor
func (e *Engine) Tick(dt float64) {
	var wg sync.WaitGroup
	wg.Add(1)
	e.mainAnchor.tick(&wg, dt, e)
	wg.Wait()
}
