package molecular

import (
	"github.com/google/uuid"
)

type Config struct {
	// MinSpeed means the minimum positive speed
	MinSpeed float64
	// MaxSpeed means the maximum positive speed
	MaxSpeed float64
}

type Engine struct {
	cfg        Config
	mainAnchor *Object // the main anchor object, must be invincible and unmoveable
	objects    map[uuid.UUID]*Object
}

func NewEngine(cfg Config) *Engine {
	return &Engine{
		cfg: cfg,
		mainAnchor: &Object{
			id:      uuid.Nil,
			attachs: make(set[*Object], 10),
		},
	}
}

func (e *Engine) Config() Config {
	return e.cfg
}

func (e *Engine) MainAnchor() *Object {
	return e.mainAnchor
}

// NewObject will create an object use v7 UUID
func (e *Engine) NewObject(anchor *Object, pos Vec) *Object {
	for i := 20; i > 0; i-- {
		if id, err := uuid.NewV7(); err == nil {
			if _, ok := e.objects[id]; !ok {
				return newObject(id, anchor, pos)
			}
		}
	}
	panic("molecular.Engine: Too many UUID generation failures")
}

// Tick will call Tick on the main anchor
func (e *Engine) Tick(dt float64) {
	e.mainAnchor.Tick(dt, e)
}
