// molecular is a 3D physics engine written in Go
// Copyright (C) 2023  Kevin Z <zyxkad@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package molecular

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

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
	system    *System
	anchor    *Object
	children  []*Object
	blocks    []Block // TODO: sort or index blocks
	gcenter   Vec3    // the gravity center
	mass      float64 // the cached mass
	pos       Vec3    // the position relative to the anchor
	tickForce Vec3
	velocity  Vec3
	angle     Vec3
	headVel   Vec3
}

func makeObjStatus() objStatus {
	return objStatus{}
}

func (s *objStatus) from(a *objStatus) {
	s.system = a.system
	s.anchor = a.anchor
	s.children = append(s.children[:0], a.children...)
	s.blocks = append(s.blocks[:0], a.blocks...)
	s.gcenter = a.gcenter
	s.mass = a.mass
	s.angle = a.angle
	s.pos = a.pos
	s.tickForce = a.tickForce
	s.velocity = a.velocity
}

func (s *objStatus) clone() (a objStatus) {
	a = *s
	a.children = append(([]*Object)(nil), s.children...)
	a.blocks = append(([]Block)(nil), s.blocks...)
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
	gfield           *GravityField
	historyGFields   []*GravityField
	gfieldUpdateMask Bitset
	gfieldUpdateCd   time.Duration

	nextMux    sync.RWMutex
	nextStatus objStatus
	nextCalls  []func()
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

		gfield: NewGravityField(ZeroVec, 0, 0),
		historyGFields: make([]*GravityField, 16), // TODO: maybe dynamically set a suitable history cache?
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
	angle=%s,
	type=%s,
}`, o.id, anchorId, o.pos, o.angle, o.typ)
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
	o.RLock()
	defer o.RUnlock()
	return o.anchor
}

func (o *Object) AnchorLocked() *Object {
	return o.anchor
}

// Pos returns the position relative to the anchor
func (o *Object) Pos() Vec3 {
	o.RLock()
	defer o.RUnlock()
	return o.pos
}

// PosLocked returns the position relative to the anchor
func (o *Object) PosLocked() Vec3 {
	return o.pos
}

// SetPos sets the position relative to the anchor
func (o *Object) SetPos(pos Vec3) {
	o.nextMux.Lock()
	defer o.nextMux.Unlock()
	o.nextStatus.pos = pos
}

// Angle returns the rotate angles
func (o *Object) Angle() Vec3 {
	return o.angle
}

// SetAngle sets the rotate angles
func (o *Object) SetAngle(angle Vec3) {
	o.nextMux.Lock()
	defer o.nextMux.Unlock()
	o.nextStatus.angle = angle
}

func (o *Object) Velocity() Vec3 {
	o.RLock()
	defer o.RUnlock()
	return o.velocity
}

func (o *Object) VelocityLocked() Vec3 {
	return o.velocity
}

func (o *Object) SetVelocity(velocity Vec3) {
	o.nextMux.Lock()
	defer o.nextMux.Unlock()
	o.nextStatus.velocity = velocity
}

// HeadingVel returns the angle velocity vector
func (o *Object) HeadingVel() Vec3 {
	return o.headVel
}

// SetHeadingVel sets the angle velocity vector
func (o *Object) SetHeadingVel(v Vec3) {
	o.nextMux.Lock()
	defer o.nextMux.Unlock()
	o.nextStatus.headVel = v
}

func (o *Object) addChild(a *Object) {
	o.nextStatus.children = append(o.nextStatus.children, a)
}

func (o *Object) removeChild(a *Object) {
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
func (o *Object) AttachTo(anchor *Object) {
	anchor.RLock()
	defer anchor.RUnlock()
	o.RLock()
	defer o.RUnlock()
	o.nextMux.Lock()
	defer o.nextMux.Unlock()

	o.AttachToLocked(anchor)
}

// AttachToLocked is same as AttachTo, but used under locked condition
// e.g. inside the object's tick
func (o *Object) AttachToLocked(anchor *Object) {
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

	return o.MainAnchorLocked()
}

// MainAnchorLocked is same as MainAnchor, but used under locked condition
func (o *Object) MainAnchorLocked() (m *Object) {
	m = o
	n := m.anchor
	for n != nil {
		m = n
		m.RLock()
		n = m.anchor
		m.RUnlock()
	}
	return
}

// AbsPos returns the position relative to the main anchor
func (o *Object) AbsPos() (p Vec3) {
	p, _ = o.AbsPosAndAnchor()
	return
}

// AbsPosLocked is same as AbsPosAndAnchor, but used under locked condition
func (o *Object) AbsPosLocked() (p Vec3) {
	p, _ = o.AbsPosAndAnchorLocked()
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
			return
		}
		m = n
		m.RLock()
		p.Add(m.pos)
	}
}

// AbsPosAndAnchorLocked is same as AbsPosAndAnchor, but used under locked condition
func (o *Object) AbsPosAndAnchorLocked() (p Vec3, m *Object) {
	p, m = o.pos, o
	n := m.anchor
	for n != nil {
		m = n
		m.RLock()
		n = m.anchor
		p.Add(m.pos)
		m.RUnlock()
	}
	return
}

var objSetPool = newObjPool[set[*Object]]()

// RelPos returns the relative position of the passed object about this object
// To be clear, return the displacement from o to a (a.pos - o.pos)
func (o *Object) RelPos(a *Object) Vec3 {
	p, m := o.AbsPosAndAnchor()
	q, n := a.AbsPosAndAnchor()
	if m == n {
		q.Sub(p)
		return q
	}
	flags := objSetPool.Get()
	if *flags == nil {
		*flags = make(set[*Object], 8)
	}
	defer objSetPool.Put(flags)
	defer clear(*flags)
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
		RotateXYZ(o.angle).
		Add(o.gcenter)
	return p
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

func (o *Object) SetRadius(radius float64) {
	o.gfield.SetRadius(radius)
}

func (o *Object) FillGfields() {
	for i := 0; i < len(o.historyGFields); i++ {
		o.historyGFields[i] = o.gfield.Clone()
	}
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

func (o *Object) GravityField() *GravityField {
	o.RLock()
	defer o.RUnlock()
	return o.gfield
}

// GravityFieldAt will returns the correct history gravity field by position.
// argument pos is the position relative to the zero position of this object
func (o *Object)GravityFieldAt(pos Vec3) Vec3 {
	if o.gfield == nil {
		return ZeroVec
	}
	radius := o.gfield.Radius()
	if pos.SqLen() < radius * radius * 4 {
		return o.gfield.FieldAt(pos)
	}
	i := math.Ilogb(pos.Subbed(o.gfield.Pos()).SqLen() / cSq) / 2
	if i < 0 {
		return o.gfield.FieldAt(pos)
	}
	if i >= len(o.historyGFields) {
		return ZeroVec
	}
	g := o.historyGFields[i]
	return g.FieldAt(pos)
}

func (o *Object) reLorentzFactor() float64 {
	if o.anchor == nil {
		return 1
	}
	return o.e.ReLorentzFactorSq(o.velocity.SqLen()) * o.anchor.reLorentzFactor()
}

func (o *Object) ProperTime(dt time.Duration) float64 {
	return dt.Seconds() / o.reLorentzFactor()
}

func (o *Object) tick(dt time.Duration) {
	o.RLock()
	defer o.RUnlock()
	o.nextMux.Lock()
	defer o.nextMux.Unlock()

	rlf := o.reLorentzFactor()
	apt := o.anchor.ProperTime(dt)
	pt := dt.Seconds() * rlf
	if pt <= 0 {
		pt = math.SmallestNonzeroFloat64
	}

	// reset the state
	o.tickForce = ZeroVec

	// tick blocks
	gcenter := ZeroVec
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
	if mass < 0 {
		mass = 0
	}
	o.nextStatus.mass = mass
	o.nextStatus.gcenter = gcenter
	if mass > 0 { // apply gravities
		pos := o.pos
		var (
			vel Vec3
			smallestL float64
			smallestO *Object
		)
		if o.anchor != nil {
			vel = o.anchor.GravityFieldAt(pos)
			vel.ScaleN(dt.Seconds())
			o.forEachSibling(func(s *Object){
				f := s.GravityFieldAt(pos.Subbed(s.pos))
				f.ScaleN(dt.Seconds())
				vel.Add(f)
			})
		}
		if smallestO != nil {
			println(smallestL)
			o.AttachToLocked(smallestO)
		}
		o.nextStatus.velocity.Add(vel)
	}
	if mass > 0 {
		o.nextStatus.velocity.Add(o.tickForce.ScaledN(pt / mass))
	}

	{ // calculate the new position and angle
		// d = (vi + vf) / 2 * ∆t
		vel := o.velocity
		vel.Add(o.nextStatus.velocity)
		vel.ScaleN(apt / 2)
		if vel.SqLen() > o.e.minSpeedSq {
			o.nextStatus.pos.Add(vel)
		}
		av := o.headVel
		av.Add(o.nextStatus.headVel)
		av.ScaleN(apt / 2)
		o.nextStatus.angle.Add(av).ModN(math.Pi)
	}
}

func (o *Object) saveStatus(dt time.Duration) {
	o.Lock()
	defer o.Unlock()
	o.nextMux.RLock()
	defer o.nextMux.RUnlock()

	posdiff := o.nextStatus.pos.Subbed(o.objStatus.pos)

	for _, cb := range o.nextCalls {
		cb()
	}
	o.nextCalls = o.nextCalls[:0]
	o.objStatus.from(&o.nextStatus)

	o.gfield.SetPos(o.gcenter)
	o.gfield.SetMass(o.mass)
	if o.gfieldUpdateCd -= dt; o.gfieldUpdateCd < 0 {
		o.gfieldUpdateCd = time.Second

		for _, g := range o.historyGFields {
			if g != nil {
				g.pos.Add(posdiff)
			}
		}

		// update history gravity fields
		last := o.historyGFields[0]
		o.historyGFields[0] = o.gfield.Clone()
		for i := 1; i < len(o.historyGFields); i++ {
			if o.gfieldUpdateMask.Flip(i) {
				if last != nil {
					gravityFieldPool.Put(last)
				}
				break
			}
			last, o.historyGFields[i] = o.historyGFields[i], last
		}
	}
}
