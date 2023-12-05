package molecular

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"

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
	children      []*Object
	blocks        []Block // TODO: sort or index blocks
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
	s.children = append(s.children[:0], a.children...)
	s.blocks = append(s.blocks[:0], a.blocks...)
	s.gcenter = a.gcenter
	s.heading = a.heading
	s.pos = a.pos
	s.tickForce = a.tickForce
	s.velocity = a.velocity
	for k, v := range a.passedGravity {
		g := s.passedGravity[k]
		if g != nil {
			g.f = v.f
			g.pos = v.pos
			g.gone = false
			g.c.Store(1)
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
	ready  atomic.Bool
	e      *Engine
	id     uuid.UUID // a v7 UUID
	typ    ObjType
	objStatus

	nextMux    sync.RWMutex
	nextStatus objStatus

	gtick uint16
}

func (e *Engine) newAndPutObject(id uuid.UUID, stat objStatus) (o *Object) {
	if stat.anchor == nil {
		stat.anchor = e.mainAnchor
	}
	o = &Object{
		e:          e,
		id:         id,
		objStatus:  stat,
		nextStatus: stat.clone(),
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

// HeadingVel returns the heading velocity vector
func (o *Object) HeadingVel() Vec3 {
	return o.headVel
}

// SetHeadingVel sets the heading velocity vector
func (o *Object) SetHeadingVel(v Vec3) {
	o.headVel = v
}

func (o *Object)addChild(a *Object){
	o.nextMux.Lock()
	defer o.nextMux.Unlock()

	o.nextStatus.children = append(o.nextStatus.children, a)
}

func (o *Object)removeChild(a *Object){
	o.nextMux.Lock()
	defer o.nextMux.Unlock()

	last := len(o.nextStatus.children) - 1
	for i, b := range o.nextStatus.children {
		if b == a {
			o.nextStatus.children[i] = o.nextStatus.children[last]
			o.nextStatus.children = o.nextStatus.children[:last]
			return
		}
	}
}

// AttachTo will change the object's anchor to another.
// The new position will be calculated at the same time.
// AttachTo must be called inside the object's tick
func (o *Object) AttachTo(anchor *Object) {
	o.RLock()
	defer o.RUnlock()
	o.nextMux.Lock()
	defer o.nextMux.Unlock()

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
		p  = o.pos
		v  = o.velocity
		v2 = anchor.velocity
	)
	p.Sub(anchor.pos)
	o.forEachAnchor(func(a *Object) {
		p.Add(a.pos)
		v.
			ScaleN(o.e.ReLorentzFactorSq(a.velocity.SqLen())).
			Add(a.velocity)
	})
	anchor.forEachAnchor(func(a *Object) {
		p.Sub(a.pos)
		v2.
			ScaleN(o.e.ReLorentzFactorSq(a.velocity.SqLen())).
			Add(a.velocity)
	})
	v.Sub(v2)
	o.nextStatus.anchor = anchor
	o.nextStatus.pos = p
	o.nextStatus.velocity = v
	o.anchor.removeChild(o)
	anchor.addChild(o)
}

// forEachAnchor will invoke the callback function on each anchor object
func (o *Object) forEachAnchor(cb func(*Object)) {
	if o.anchor == nil {
		return
	}
	a := o.anchor

	a.RLock()
	defer a.RUnlock()

	cb(a)
	a.forEachAnchor(cb)
}

// AbsPos returns the position relative to the main anchor
func (o *Object) AbsPos() (p Vec3) {
	o.RLock()
	defer o.RUnlock()

	p = o.pos
	o.forEachAnchor(func(a *Object) {
		p.Add(a.pos)
	})
	return
}

func (o *Object) RotatePos(p *Vec3) *Vec3 {
	o.RLock()
	defer o.RUnlock()

	if o.anchor != nil {
		p.
			Sub(o.gcenter).
			RotateXYZ(o.heading).
			Add(o.gcenter)
	}
	return p
}

func (o *Object) Velocity() Vec3 {
	return o.velocity
}

func (o *Object) SetVelocity(velocity Vec3) {
	o.velocity = velocity
}

func (o *Object) AbsVelocity() (v Vec3) {
	o.RLock()
	defer o.RUnlock()

	v = o.velocity
	o.forEachAnchor(func(a *Object) {
		v.
			ScaleN(o.e.ReLorentzFactorSq(a.velocity.SqLen())).
			Add(a.velocity)
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
	return o.nextStatus.blocks
}

func (o *Object) SetBlocks(blocks []Block) {
	o.nextStatus.blocks = blocks
}

func (o *Object) AddBlock(b ...Block) {
	o.nextStatus.blocks = append(o.nextStatus.blocks, b...)
}

func (o *Object) RemoveBlock(target Block) {
	blocks := o.nextStatus.blocks
	last := len(blocks) - 1
	for i, b := range blocks {
		if b == target {
			blocks[i] = blocks[last]
			o.nextStatus.blocks = blocks[:last]
			return
		}
	}
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

// // SetGField sets the gravitational field
// func (o *Object) SetGField(field *GravityField) {
// 	o.gfield = field
// }

func (o *Object) reLorentzFactor() float64 {
	if o.anchor == nil {
		return 1
	}
	return o.e.ReLorentzFactorSq(o.velocity.Len()) * o.anchor.reLorentzFactor()
}

func (o *Object) ProperTime(dt float64) float64 {
	return dt / o.reLorentzFactor()
}

func (o *Object) tick(dt float64) {
	o.nextMux.Lock()
	defer o.nextMux.Unlock()

	rlf := o.reLorentzFactor()
	apt := o.anchor.ProperTime(dt)
	pt := dt * rlf
	if pt <= 0 {
		pt = math.SmallestNonzeroFloat64
	}

	// reset the state
	o.tickForce = ZeroVec
	gcenter := ZeroVec

	// tick blocks
	mass := 0.0
	for _, b := range o.blocks {
		b.Tick(pt, o)
		l := b.Outline()
		m := b.Mass()
		mass += m
		c := l.Center()
		if mass == 0 {
			gcenter = c
		} else {
			gcenter.Add(c.Subbed(gcenter).ScaledN(m / mass))
		}
	}
	o.nextStatus.gcenter = gcenter

	pos := o.AbsPos()
	{ // apply the gravity
		factor := mass / rlf
		var (
			smallestL float64
			smallestO *Object
		)
		v := o.AbsVelocity()
		for a, g := range o.passedGravity {
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
				delete(o.passedGravity, a)
			}
		}
		if smallestO != nil {
			o.AttachTo(smallestO)
		}
	}
	if mass > 0 {
		o.nextStatus.velocity.Add(o.tickForce.ScaledN(pt / mass))
	} else if mass < 0 {
		mass = 0
	}

	moved := false
	{ // calculate the new position and angle
		// d = (vi + vf) / 2 * ∆t
		vel := o.nextStatus.velocity
		vel.Add(o.nextStatus.velocity)
		vel.ScaleN(apt / 2)
		if vel.SqLen() > o.e.minSpeedSq {
			o.pos.Add(vel)
			moved = true
		}
		av := o.headVel
		av.Add(o.headVel)
		av.ScaleN(apt / 2)
		o.nextStatus.heading.Add(av).ModN(math.Pi)
	}

	// queue gravity change event
	center := o.AbsPos()
	center.Add(gcenter)
	if gfield := o.nextStatus.gfield; gfield != nil {
		lastNil := o.gfield == nil
		if lastNil {
			o.gtick = 0
		} else {
			o.gtick++
		}
		if o.typ == ManMadeObj {
			if lastNil || mass != gfield.Mass() {
				gfield.SetMass(mass)
				if mass > 0 {
					o.e.queueEvent(o.e.newGravityWave(o, center, gfield, o.gtick))
				}
			}
		} else if lastNil || moved {
			o.e.queueEvent(o.e.newGravityWave(o, center, gfield, o.gtick))
		}
	}
}

func (o *Object) saveStatus() {
	o.Lock()
	defer o.Unlock()
	o.nextMux.RLock()
	defer o.nextMux.RUnlock()

	o.objStatus.from(&o.nextStatus)
}
