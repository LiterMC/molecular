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

type objStatus struct {
	anchor        *Object
	gcenter       Vec3 // the gravity center
	gfield        *GravityField
	pos           Vec3 // the position relative to the anchor
	tickForce     Vec3
	velocity      Vec3
	heading       Vec3 // X=pitch, Y=yaw, Z=roll; facing Z+
	headVel       Vec3
	passedGravity map[*Object]*gravityStatus
}

func makeObjStatus() objStatus {
	return objStatus{
		passedGravity: make(map[*Object]*gravityStatus, 5),
	}
}

func (s *objStatus) from(a *objStatus) {
	s.anchor = a.anchor
	s.gcenter = a.gcenter
	s.heading = a.heading
	s.pos = a.pos
	s.tickForce = a.tickForce
	s.velocity = a.velocity
	for k, v := range a.passedGravity {
		g := s.passedGravity[k]
		if g != nil {
			*g = *v
			g.c = 1
		} else {
			s.passedGravity[k] = v.clone()
		}
	}
	for k, g := range s.passedGravity {
		if g.gone {
			g.release()
			delete(s.passedGravity, k)
		} else {
			g.gone = true
		}
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
	a.passedGravity = make(map[*Object]*gravityStatus, len(s.passedGravity))
	for k, v := range s.passedGravity {
		a.passedGravity[k] = v.clone()
	}
	return
}

// Object represents an object in the physics engine.
type Object struct {
	sync.RWMutex
	e      *Engine
	id     uuid.UUID // a v7 UUID
	typ    ObjType
	blocks []Block
	objStatus

	// lastStatus should only be read during a tick
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
}`, o.id, o.anchor.id, o.pos, o.heading.X, o.heading.Y, o.heading.Z, o.typ)
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

// Heading returns the rotate vector
// X == pitch
// Y == yaw
// Z == roll
func (o *Object) Heading() Vec3 {
	return o.heading
}

// SetHeading sets the heading angles
func (o *Object) SetHeading(heading Vec3) {
	o.heading = heading
}

// HeadingPYR returns pitch, yaw, and roll
func (o *Object) HeadingPYR() (pitch, yaw, roll float64) {
	return o.heading.XYZ()
}

// HeadingVel returns the heading velocity vector
func (o *Object) HeadingVel() Vec3 {
	return o.headVel
}

// SetHeadingVel sets the heading velocity vector
func (o *Object) SetHeadingVel(v Vec3) {
	o.headVel = v
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
		p  = o.lastStatus.pos
		v  = o.lastStatus.velocity
		v2 = anchor.lastStatus.velocity
	)
	p.Sub(anchor.lastStatus.pos)
	o.forEachAnchor(func(a *Object) {
		p.Add(a.lastStatus.pos)
		v.ScaleN(o.e.ReLorentzFactorSq(a.lastStatus.velocity.SqLen()))
		v.Add(a.lastStatus.velocity)
	})
	anchor.forEachAnchor(func(a *Object) {
		p.Sub(a.lastStatus.pos)
		v2.ScaleN(o.e.ReLorentzFactorSq(a.lastStatus.velocity.SqLen()))
		v2.Add(a.lastStatus.velocity)
	})
	v.Sub(v2)
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
		a.RotatedPos(p)
		p.Add(a.lastStatus.pos)
	})
	return
}

func (o *Object) RotatedPos(p Vec3) Vec3 {
	p.Sub(o.lastStatus.gcenter)
	p.RotateXYZ(o.lastStatus.heading)
	p.Add(o.lastStatus.gcenter)
	return p
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
		v.ScaleN(o.e.ReLorentzFactorSq(a.lastStatus.velocity.SqLen()))
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

func (o *Object) reLorentzFactor() float64 {
	if o.lastStatus.anchor == nil {
		return 1
	}
	return o.e.ReLorentzFactorSq(o.lastStatus.velocity.Len()) * o.lastStatus.anchor.reLorentzFactor()
}

func (o *Object) ProperTime(dt float64) float64 {
	return dt / o.reLorentzFactor()
}

func (o *Object) tick(dt float64) {
	o.Lock()
	defer o.Unlock()

	rlf := o.reLorentzFactor()
	apt := o.anchor.ProperTime(dt)
	pt := dt * rlf
	if pt <= 0 {
		pt = math.SmallestNonzeroFloat64
	}

	// reset the state
	o.tickForce = ZeroVec
	o.gcenter = ZeroVec

	// tick blocks
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
	pos := o.AbsPos()
	{ // apply the gravity
		factor := mass / rlf
		var (
			smallestL float64
			smallestO *Object
		)
		v := o.AbsVelocity()
		for a, g := range o.lastStatus.passedGravity {
			l := v.Subbed(a.AbsVelocity()).SqLen()
			if smallestL > l {
				smallestL = l
				smallestO = a
			}
			f := g.FieldAt(pos)
			if f.SqLen() >= o.e.minAccelSq {
				f.ScaleN(factor)
				o.tickForce.Add(f)
			} else {
				g.release()
				delete(o.lastStatus.passedGravity, a)
			}
		}
		if smallestO != nil {
			o.AttachTo(smallestO)
		}
	}
	if mass > 0 {
		o.velocity.Add(o.tickForce.ScaledN(pt / mass))
	} else if mass < 0 {
		mass = 0
	}

	moved := false
	{ // calculate the new position and angle
		// d = (vi + vf) / 2 * ∆t
		vel := o.lastStatus.velocity
		vel.Add(o.velocity)
		vel.ScaleN(apt / 2)
		if vel.SqLen() > o.e.minSpeedSq {
			o.pos.Add(vel)
			moved = true
		}
		av := o.lastStatus.headVel
		av.Add(o.headVel)
		av.ScaleN(apt / 2)
		o.heading.Add(av).ModN(math.Pi)
	}

	// queue gravity change event
	gcenter := o.AbsPos()
	gcenter.Add(o.gcenter)
	if o.gfield != nil {
		lastNil := o.lastStatus.gfield == nil
		if o.typ == ManMadeObj {
			if lastNil || mass != o.gfield.Mass() {
				o.gfield.SetMass(mass)
				if mass > 0 {
					o.e.queueEvent(o.e.newGravityWave(o, gcenter, o.gfield))
				}
			}
		} else if lastNil || moved {
			o.e.queueEvent(o.e.newGravityWave(o, gcenter, o.gfield))
		}
	}
}

func (o *Object) saveStatus() {
	o.RLock()
	defer o.RUnlock()

	o.lastStatus.from(&o.objStatus)
}
