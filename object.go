package molecular

import (
	"fmt"

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

type objStatus struct {
	anchor           *Object
	gcenter          Vec3 // the gravity center
	gfield           *GravityField
	pitch, yaw, roll float64
	pos              Vec3 // the position relative to the anchor
	tickForce        Vec3
	velocity         Vec3
	passedGravity    map[*Object]Vec3
}

func makeObjStatus() objStatus {
	return objStatus{
		passedGravity: make(map[*Object]Vec3, 5),
	}
}

func (s *objStatus) from(a *objStatus) {
	s.anchor = a.anchor
	s.gcenter = a.gcenter
	s.pitch = a.pitch
	s.yaw = a.yaw
	s.roll = a.roll
	s.pos = a.pos
	s.tickForce = a.tickForce
	s.velocity = a.velocity
	clear(s.passedGravity)
	for k, v := range a.passedGravity {
		s.passedGravity[k] = v
	}
	if a.gfield != nil {
		if s.gfield == nil {
			s.gfield = new(GravityField)
		}
		*s.gfield = *a.gfield
	} else {
		s.gfield = nil
	}
}

func (s *objStatus) clone() (a objStatus) {
	a = *s
	if a.gfield != nil {
		a.gfield = new(GravityField)
		*a.gfield = *s.gfield
	}
	a.passedGravity = make(map[*Object]Vec3, len(s.passedGravity))
	for k, v := range s.passedGravity {
		a.passedGravity[k] = v
	}
	return
}

// Object represents an object in the physics engine.
type Object struct {
	e      *Engine
	id     uuid.UUID // a v7 UUID
	typ    ObjType
	blocks []Block
	objStatus

	lastStatus objStatus
}

func (e *Engine) newAndPutObject(id uuid.UUID, stat objStatus) (o *Object) {
	if stat.anchor == nil {
		stat.anchor = e.mainAnchor
	}
	o = &Object{
		e:          e,
		id:         id,
		objStatus:  stat,
		lastStatus: stat.clone(),
	}
	if _, ok := e.objects[id]; ok {
		panic("molecular.Engine: Object id " + id.String() + " is already exists")
	}
	e.objects[id] = o
	return
}

func (o *Object) String() string {
	if o.anchor == nil {
		return "Object[MainAnchor]"
	}
	return fmt.Sprintf(`Object[%s]{
	anchor=%s,
	pos=%v,
	facing=(pitch=%v, yaw=%v, roll=%v),
	type=%s,
}`, o.id, o.anchor.id, o.pos, o.pitch, o.yaw, o.roll, o.typ)
}

// An object's id will never be changed
func (o *Object) Id() uuid.UUID {
	return o.id
}

// Engine returns the engine of the object
func (o *Object) Engine() *Engine {
	return o.e
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
// AttachTo must be called inside the object's tick
func (o *Object) AttachTo(anchor *Object) {
	if anchor == nil {
		panic("molecular.Object: new anchor cannot be nil")
	}
	if o.anchor == nil {
		panic("molecular.Object: cannot attach main anchor")
	}

	if anchor == o.anchor {
		return
	}

	var (
		p = o.lastStatus.pos
		v = o.lastStatus.velocity
	)
	p.Sub(anchor.lastStatus.pos)
	v.Sub(anchor.lastStatus.velocity)
	o.forEachAnchor(func(a *Object) {
		p.Add(a.lastStatus.pos)
		v.Add(a.lastStatus.velocity)
	})
	anchor.forEachAnchor(func(a *Object) {
		p.Sub(a.lastStatus.pos)
		v.Sub(a.lastStatus.velocity)
	})
	o.anchor = anchor
	o.pos = p
	o.velocity = v
}

// forEachAnchor will invoke the callback function on each anchor object
func (o *Object) forEachAnchor(cb func(*Object)) {
	if o.lastStatus.anchor == nil {
		return
	}
	cb(o.lastStatus.anchor)
	o.lastStatus.anchor.forEachAnchor(cb)
}

// AbsPos returns the position relative to the main anchor
func (o *Object) AbsPos() (p Vec3) {
	p = o.lastStatus.pos
	o.forEachAnchor(func(a *Object) {
		p.Add(a.lastStatus.pos)
	})
	return
}

func (o *Object) Velocity() Vec3 {
	return o.velocity
}

func (o *Object) SetVelocity(velocity Vec3) {
	o.velocity = velocity
}

func (o *Object) AbsVelocity() (v Vec3) {
	v = o.lastStatus.velocity
	o.forEachAnchor(func(a *Object) {
		v.Add(a.lastStatus.velocity)
	})
	return
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

func (o *Object) tick(dt float64) {
	vel := o.velocity
	vl := vel.Len()
	if o.e.cfg.MaxSpeed > 0 {
		if vl > o.e.cfg.MaxSpeed {
			vl = o.e.cfg.MaxSpeed
		}
	} else if vl >= C {
		vl = 0
	}
	lf := o.e.LorentzFactor(vl)
	pt := o.e.ProperTime(dt, vl)

	// reset the state
	o.tickForce = ZeroVec
	o.gcenter = ZeroVec

	mass := 0.0
	for i, b := range o.blocks {
		b.Tick(pt, o)
		l := b.Outline()
		m := b.Mass()
		mass += m
		c := l.Center()
		if i == 0 {
			o.gcenter = c
		} else {
			o.gcenter.Add(c.Subbed(o.gcenter).ScaledN(m / mass))
		}
	}
	{ // apply the gravity
		factor := lf * mass
		var (
			largestL float64
			largestO *Object
		)
		for a, f := range o.lastStatus.passedGravity {
			f.ScaleN(factor)
			if l := f.SqLen(); largestL < l {
				largestL = l
				largestO = a
			}
			o.tickForce.Add(f)
		}
		if largestO != nil {
			o.AttachTo(largestO)
		}
	}
	if mass > 0 {
		o.velocity.Add(o.e.AccFromForce(mass, vl, o.tickForce).ScaledN(dt))
	} else if mass < 0 {
		mass = 0
	}

	// (vi + vf) / 2 * ∆t
	vel.Add(o.velocity)
	vel.ScaleN(dt / 2)
	if vel.SqLen() > o.e.cfg.MinSpeed*o.e.cfg.MinSpeed {
		o.pos.Add(vel)
	}

	// queue gravity change event
	gcenter := o.AbsPos()
	gcenter.Add(o.gcenter)
	if lastNil := o.lastStatus.gfield == nil; o.gfield != nil {
		if o.typ == ManMadeObj {
			if lastNil || mass != o.gfield.Mass() {
				o.gfield.SetMass(mass)
				var cb func(r *Object)
				if mass == 0 {
					cb = func(r *Object) {
						delete(r.passedGravity, o)
					}
				}else{
					g := *o.gfield
					cb = func(r *Object) {
						r.passedGravity[o] = g.FieldAt(r.AbsPos().Subbed(gcenter))
					}
				}
				o.e.queueEvent(newEventWave(o, gcenter, -1, cb))
			}
		}else if lastNil {
			g := *o.gfield
			o.e.queueEvent(newEventWave(o, gcenter, -1, func(r *Object) {
				r.passedGravity[o] = g.FieldAt(r.AbsPos().Subbed(gcenter))
			}))
		}
	}else if !lastNil {
		o.e.queueEvent(newEventWave(o, gcenter, -1, func(r *Object) {
			delete(r.passedGravity, o)
		}))
	}
}

func (o *Object) saveStatus() {
	o.lastStatus.from(&o.objStatus)
}
