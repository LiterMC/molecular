package molecular

import (
	"fmt"
	"math"
	"sync"

	"github.com/google/uuid"
)

type ObjType uint8

const (
	ManMadeObj ObjType = 0
	NaturalObj ObjType = 1
	LivingObj  ObjType = 2
)

func (t ObjType) String() string {
	switch t {
	case ManMadeObj:
		return "man-made"
	case NaturalObj:
		return "natural"
	case LivingObj:
		return "living"
	default:
		panic("Unknown object type value")
	}
}

// Object represents an object in the physics engine.
type Object struct {
	mux              sync.RWMutex
	e                *Engine
	id               uuid.UUID // a v7 UUID
	anchor           *Object
	attachs          set[*Object]
	typ              ObjType
	blocks           []Block
	gcenter          Vec3 // the gravity center
	gfield           *GravityField
	pitch, yaw, roll float64
	pos              Vec3 // the position relative to the anchor
	tickForce        Vec3
	velocity         Vec3
}

func (e *Engine) newAndPutObject(id uuid.UUID, anchor *Object, pos Vec3) (o *Object) {
	if anchor == nil {
		anchor = e.mainAnchor
	}
	o = &Object{
		e:       e,
		id:      id,
		anchor:  anchor,
		attachs: make(set[*Object], 10),
		pos:     pos,
	}
	anchor.attachs.Put(o)
	if _, ok := e.objects[id]; ok {
		panic("molecular.Engine: Object id " + id.String() + " is already exists")
	}
	e.objects[id] = o
	return
}

func (o *Object) String() string {
	return fmt.Sprintf(`Object[%s]{
	anchor=%s,
	pos=%v,
	facing=(pitch=%v, yaw=%v, roll=%v),
	type=%s,
}`, o.id, o.anchor, o.pos, o.pitch, o.yaw, o.roll, o.typ)
}

// An object's id will never be changed
func (o *Object) Id() uuid.UUID {
	return o.id
}

// Engine returns the engine of the object
func (o *Object) Engine() *Engine {
	return o.e
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
func (o *Object) Pos() Vec3 {
	return o.pos
}

// SetPos sets the position relative to the anchor
func (o *Object) SetPos(pos Vec3) {
	o.pos = pos
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
	o.pos = p1.Subbed(p2)
	o.anchor.mux.Lock()
	o.anchor.attachs.Remove(o)
	o.anchor.mux.Unlock()
	anchor.mux.Lock()
	anchor.attachs.Put(o)
	anchor.mux.Unlock()
	o.anchor = anchor
}

// forEachAnchor will invoke the callback function on each anchor object.
// forEachAnchor will lock the reader locker for the anchors.
func (o *Object) forEachAnchor(cb func(*Object)) {
	o.mux.RLock()
	defer o.mux.RUnlock()

	if o.anchor == nil {
		return
	}
	cb(o.anchor)
	o.anchor.forEachAnchor(cb)
}

// AbsPos returns the position relative to the main anchor
func (o *Object) AbsPos() (p Vec3) {
	o.mux.RLock()
	if o.anchor == nil {
		o.mux.RUnlock()
		return
	}
	p = o.pos
	o.mux.RUnlock()

	o.forEachAnchor(func(a *Object) {
		p.Add(a.pos)
	})
	return
}

func (o *Object) Velocity() Vec3 {
	return o.velocity
}

func (o *Object) SetVelocity(velocity Vec3) {
	o.velocity = velocity
}

func (o *Object) Type() ObjType {
	return o.typ
}

func (o *Object) SetType(t ObjType) {
	o.typ = t
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

// TickForce returns the force vector that can be edit during a tick.
// You should never read/write the vector concurrently or outside a tick.
func (o *Object) TickForce() *Vec3 {
	return &o.tickForce
}

// GField returns the gravitational field
func (o *Object) GField() *GravityField {
	return o.gfield
}

// SetGField sets the gravitational field
func (o *Object) SetGField(field *GravityField) {
	o.gfield = field
}

func (o *Object) _tick(dt float64) {
	vel := o.velocity
	vl := vel.Len()
	if o.e.cfg.MaxSpeed > 0 {
		vl = math.Min(o.e.cfg.MaxSpeed, vl)
	}
	pt := o.e.ProperTime(dt, vl)

	// reset the state
	o.tickForce = ZeroVec
	{ // apply the gravity
		p := o.pos
		o.forEachAnchor(func(a *Object) {
			if a.gfield != nil {
				f := a.gfield.FieldAt(p.Subbed(a.gcenter))
				f.ScaledN(dt)
				o.velocity.Add(f)
			}
			p.Add(a.pos)
		})
	}

	// TODO: update gcenter

	mass := 0.0
	for _, b := range o.blocks {
		b.Tick(pt, o)
		mass += b.Mass()
	}
	if mass > 0 {
		o.velocity.Add(o.e.AccFromForce(mass, o.velocity.Len(), o.tickForce))
	}
	if o.gfield != nil && o.typ == ManMadeObj {
		o.gfield.SetMass(mass)
	}

	if vl > o.e.cfg.MinSpeed {
		o.pos.Add(vel.ScaledN(dt))
	}
}

// tick will tick the object itself and it's attachments concurrently
// it will call wg.Done when exit
func (o *Object) tick(wg *sync.WaitGroup, dt float64, e *Engine) {
	defer wg.Done()

	o.mux.RLock()
	for a := range o.attachs {
		wg.Add(1)
		go a.tick(wg, dt, e)
	}
	o.mux.RUnlock()

	if o.anchor != nil { // only tick on non-main anchor
		o.mux.Lock()
		defer o.mux.Unlock()
		o._tick(dt)
	}
}
