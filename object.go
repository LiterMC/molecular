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
	system        *System
	anchor        *Object
	children      []*Object
	blocks        []Block       // TODO: sort or index blocks
	gcenter       Vec3          // the gravity center
	mass          float64       // the cached mass
	gfield        *GravityField // the gravity field
	pos           Vec3          // the position relative to the anchor
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
	s.system = a.system
	s.anchor = a.anchor
	s.children = append(s.children[:0], a.children...)
	s.blocks = append(s.blocks[:0], a.blocks...)
	s.gcenter = a.gcenter
	s.mass = a.mass
	s.heading = a.heading
	s.pos = a.pos
	s.tickForce = a.tickForce
	s.velocity = a.velocity
	for k, v := range a.passedGravity {
		g := s.passedGravity[k]
		if g != nil {
			g.f = v.f
			g.pos = v.pos
			g.life = v.life
			g.c.Store(1)
		} else {
			s.passedGravity[k] = v.clone()
		}
	}
	for k, g := range s.passedGravity {
		if g.life == 0 {
			g.release()
			delete(s.passedGravity, k)
		} else {
			g.life--
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
	a.children = append(make([]*Object, 0, len(s.children)), s.children...)
	a.blocks = append(make([]Block, 0, len(s.blocks)), s.blocks...)
	a.passedGravity = make(map[*Object]*gravityStatus, len(s.passedGravity))
	for k, v := range s.passedGravity {
		a.passedGravity[k] = v.clone()
	}
	return
}

// Object represents an object in the physics engine.
type Object struct {
	sync.RWMutex
	ready atomic.Bool
	e     *Engine
	id    uuid.UUID // a v7 UUID
	typ   ObjType
	objStatus

	nextMux    sync.RWMutex
	nextStatus objStatus
	nextCalls  []func()

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

	for _, b := range stat.blocks {
		b.SetObject(o)
	}
	return
}

func (o *Object) String() string {
	return fmt.Sprintf("Object[%s]", o.id)
}

func (o *Object) GoString() string {
	anchorId := "nil"
	if o.anchor != nil {
		anchorId = o.anchor.id.String()
	}
	return fmt.Sprintf(`Object[%s]{
	anchor=%s,
	pos=%v,
	facing=(pitch=%v, yaw=%v, roll=%v),
	type=%s,
}`, o.id, anchorId, o.pos, o.heading.X, o.heading.Y, o.heading.Z, o.typ)
}

// An object's id will never be changed
func (o *Object) Id() uuid.UUID {
	return o.id
}

// Engine returns the engine of the object
func (o *Object) Engine() *Engine {
	return o.e
}

func (o *Object) Type() ObjType {
	return o.typ
}

func (o *Object) SetType(t ObjType) {
	o.typ = t
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
	o.nextStatus.pos = pos
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
	o.nextStatus.heading = heading
}

// HeadingVel returns the heading velocity vector
func (o *Object) HeadingVel() Vec3 {
	return o.headVel
}

// SetHeadingVel sets the heading velocity vector
func (o *Object) SetHeadingVel(v Vec3) {
	o.nextStatus.headVel = v
}

func (o *Object) addChild(a *Object) {
	o.nextMux.Lock()
	defer o.nextMux.Unlock()

	o.nextStatus.children = append(o.nextStatus.children, a)
}

func (o *Object) removeChild(a *Object) {
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
// the object has to be read locked before invoke
func (o *Object) forEachAnchor(cb func(*Object)) {
	if o.anchor == nil {
		return
	}
	a := o.anchor

	cb(a)

	a.RLock()
	defer a.RUnlock()

	a.forEachAnchor(cb)
}

// forEachSibling will invoke the callback function on each sibling object
// the object has to be read locked before invoke
func (o *Object) forEachSibling(cb func(*Object)) {
	if o.anchor == nil {
		return
	}
	a := o.anchor
	for _, s := range a.children {
		if s != o {
			cb(s)
		}
	}
}

// MainAnchor returns the main anchor object
func (o *Object) MainAnchor() (m *Object) {
	o.RLock()
	defer o.RUnlock()

	m = o
	for m.anchor != nil {
		m = m.anchor
	}
	return
}

// AbsPos returns the position relative to the main anchor
func (o *Object) AbsPos() (p Vec3) {
	p, _ = o.AbsPosAndAnchor()
	return
}

// AbsPosAndAnchor combine the results of AbsPos and MainAnchor
func (o *Object) AbsPosAndAnchor() (p Vec3, m *Object) {
	o.RLock()

	p, m = o.pos, o
	for {
		n := m.anchor
		m.RUnlock()
		if n == nil {
			break
		}
		m = n
		m.RLock()
		p.Add(m.pos)
	}
	return
}

var objSetPool = sync.Pool{
	New: func() any {
		return make(set[*Object], 8)
	},
}

// RelPos returns the relative position of the passed object about this object
// To be clear, return the displacement from o to a (a.pos - o.pos)
func (o *Object) RelPos(a *Object) Vec3 {
	p, m := o.AbsPosAndAnchor()
	q, n := a.AbsPosAndAnchor()
	if m == n {
		q.Sub(p)
		return q
	}
	flags := objSetPool.Get().(*set[*Object])
	defer objSetPool.Put(flags)
	clear(*flags)
	pos, ok := m.findRelPos(n, *flags)
	if !ok {
		panic("molecular.Object: calling RelPos() on two unrelative objects")
	}
	pos.Add(q).Sub(p)
	return pos
}

// findRelPos should only be called on main anchor
func (o *Object) findRelPos(target *Object, flags set[*Object]) (pos Vec3, ok bool) {
	o.RLock()
	defer o.RUnlock()

	if o.anchor != nil {
		panic("molecular.Object: findRelPos() should only be called on main anchor")
	}

	if flags.Has(o) {
		return
	}
	flags.Put(o)
	if pos, ok = o.system.anchorPos[target]; ok {
		return
	}
	for a, p := range o.system.anchorPos {
		if pos, ok = a.findRelPos(target, flags); ok {
			pos.Add(p)
			return
		}
	}
	return
}

func (o *Object) RotatePos(p *Vec3) *Vec3 {
	o.RLock()
	defer o.RUnlock()

	p.
		Sub(o.gcenter).
		RotateXYZ(o.heading).
		Add(o.gcenter)
	return p
}

func (o *Object) Velocity() Vec3 {
	return o.velocity
}

func (o *Object) SetVelocity(velocity Vec3) {
	o.nextStatus.velocity = velocity
}

func (o *Object) AbsVelocity() (v Vec3) {
	v, m := o.velocity, o
	for m.anchor != nil {
		m = m.anchor
		v.
			ScaleN(o.e.ReLorentzFactorSq(m.velocity.SqLen())).
			Add(m.velocity)
	}
	return
}

func (o *Object) Blocks() []Block {
	return o.nextStatus.blocks
}

func (o *Object) SetBlocks(blocks []Block) {
	o.nextCalls = append(o.nextCalls, func() {
		for _, b := range blocks {
			b.SetObject(o)
		}
	})
	o.nextStatus.blocks = blocks
}

func (o *Object) AddBlock(blocks ...Block) {
	o.nextCalls = append(o.nextCalls, func() {
		for _, b := range blocks {
			b.SetObject(o)
		}
	})
	o.nextStatus.blocks = append(o.nextStatus.blocks, blocks...)
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

// SetGField sets the gravitational field
func (o *Object) SetGField(field *GravityField) {
	o.nextStatus.gfield = field
}

func (o *Object) Mass() (mass float64) {
	o.RLock()
	defer o.RUnlock()

	mass = o.mass
	for _, a := range o.children {
		mass += a.Mass()
	}
	return
}

func (o *Object) GravityCenterAndMass() (center Vec3, mass float64) {
	o.RLock()
	defer o.RUnlock()

	center, mass = o.gcenter, o.mass
	for _, a := range o.children {
		g, m := a.GravityCenterAndMass()
		mass += m
		if mass == 0 {
			center = g
		} else {
			center.Add(g.Subbed(center).ScaledN(m / mass))
		}
	}
	return
}

func (o *Object) GravityCenter() (center Vec3) {
	center, _ = o.GravityCenterAndMass()
	return
}

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
		b.Tick(pt)
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
	o.nextStatus.mass = mass
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
			println(smallestL)
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
		vel := o.velocity
		vel.Add(o.nextStatus.velocity)
		vel.ScaleN(apt / 2)
		if vel.SqLen() > o.e.minSpeedSq {
			o.nextStatus.pos.Add(vel)
			moved = true
		}
		av := o.headVel
		av.Add(o.nextStatus.headVel)
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

	for _, cb := range o.nextCalls {
		cb()
	}
	o.nextCalls = o.nextCalls[:0]
	o.objStatus.from(&o.nextStatus)
	clear(o.nextStatus.passedGravity)
}
