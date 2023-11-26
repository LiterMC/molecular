package molecular

import (
	"math"

	"github.com/google/uuid"
)

// Object represents an object in the physics engine
// If the object is a main anchor, it's position will always be zero,
// and it should not contains any block
// Only the objects that have GravityField can become an sub-anchor
type Object struct {
	id      uuid.UUID // a v7 UUID
	attachs set[*Object]
	anchor  *Object
	pos     Vec // the position relative to the anchor
	facing  Vec
	blocks  []Block
	speed   Vec
	field   *GravityField
}

func newObject(id uuid.UUID, anchor *Object, pos Vec) (o *Object) {
	o = &Object{
		id:      id,
		attachs: make(set[*Object], 10),
		anchor:  anchor,
		pos:     pos,
	}
	anchor.attachs.Put(o)
	return
}

func (o *Object) Id() uuid.UUID {
	return o.id
}

// Attachs returns all the objects that anchored to this object directly
func (o *Object) Attachs() []*Object {
	return o.attachs.AsSlice()
}

// Anchor returns this object's anchor object
// If Anchor returns nil, the object is the main anchor
func (o *Object) Anchor() *Object {
	return o.anchor
}

// Pos returns the position relative to the anchor
func (o *Object) Pos() Vec {
	if o.anchor == nil {
		return ZeroVec
	}
	return o.pos
}

// SetPos sets the position relative to the anchor
func (o *Object) SetPos(pos Vec) {
	o.pos = pos
}

func (o *Object) Facing() Vec {
	if o.anchor == nil {
		return ZeroVec
	}
	return o.facing
}

func (o *Object) SetFacing(f Vec) {
	o.facing = f
}

func (o *Object) Speed() Vec {
	return o.speed
}

func (o *Object) SetSpeed(speed Vec) {
	o.speed = speed
}

func (o *Object) Blocks() []Block {
	return o.blocks
}

func (o *Object) SetBlocks(blocks []Block) {
	o.blocks = blocks
}

func (o *Object) AddBlock(b Block) {
	o.blocks = append(o.blocks, b)
}

// AbsPos returns the position relative to the main anchor
func (o *Object) AbsPos() Vec {
	if o.anchor == nil {
		return ZeroVec
	}
	return o.anchor.AbsPos().Added(o.pos)
}

// AttachTo will change the object's anchor to another.
// The new position will be calculated at the same time.
func (o *Object) AttachTo(anchor *Object) {
	if anchor == nil {
		panic("molecular.Object: new anchor cannot be nil")
	}
	if o.anchor == nil {
		panic("molecular.Object: cannot attach main anchor")
	}
	p1, p2 := o.AbsPos(), anchor.AbsPos()
	o.anchor.attachs.Remove(o)
	o.anchor = anchor
	o.pos = p1.Subbed(p2)
	anchor.attachs.Put(o)
}

func (o *Object) tick(dt float64, e *Engine) {
	if o.anchor == nil {
		return
	}

	speed := o.speed.Len()
	if e.cfg.MaxSpeed > 0 {
		speed = math.Min(e.cfg.MaxSpeed, speed)
	}
	dt = e.RelativeDeltaTime(dt, speed)

	mass := 0.0
	for _, b := range o.blocks {
		b.Tick(dt)
		mass += b.Mass()
	}

	o.pos.Add(o.speed.ScaledN(dt))
}

// Tick will tick the object and it's attachments.
func (o *Object) Tick(dt float64, e *Engine) {
	o.tick(dt, e)

	for a := range o.attachs {
		a.Tick(dt, e)
	}
}
