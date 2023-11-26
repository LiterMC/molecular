package molecular

type Cube struct {
	P Vec // Pos
	S Vec // Size
}

func NewCube(pos, size Vec) (b *Cube) {
	if size.X < 0 {
		pos.X += size.X
		size.X = -size.X
	}
	if size.Y < 0 {
		pos.Y += size.Y
		size.Y = -size.Y
	}
	if size.Z < 0 {
		pos.Z += size.Z
		size.Z = -size.Z
	}
	return &Cube{
		P: pos,
		S: size,
	}
}

func (b *Cube) String() string {
	return "Cube(pos=" + b.P.String() + ", size=" + b.S.String() + ")"
}

func (b *Cube) Equals(x *Cube) bool {
	return b.P == x.P && b.S == x.S
}

func (b *Cube) Pos() Vec {
	return b.P
}

func (b *Cube) Size() Vec {
	return b.S
}

func (b *Cube) EndPos() Vec {
	return b.P.Added(b.S)
}

// Overlap will return if the two Cube overlapped or not
func (b *Cube) Overlap(x *Cube) bool {
	p1, p2 := b.Pos(), b.EndPos()
	q1, q2 := x.Pos(), x.EndPos()
	a1, a2 := q1.Subbed(p1), p2.Subbed(q1)
	b1, b2 := q2.Subbed(p1), p2.Subbed(q2)
	return (a1.X >= 0 && a2.X >= 0 || b1.X >= 0 && b2.X >= 0) &&
		(a1.Y >= 0 && a2.Y >= 0 || b1.Y >= 0 && b2.Y >= 0) &&
		(a1.Z >= 0 && a2.Z >= 0 || b1.Z >= 0 && b2.Z >= 0)
}

// OverlapBox will calcuate the overlapped area.
// If overlapped, the method will save the overlapped area into the second argument `area`,
// relative to the Cube `b`, and returns true.
// Note: `area` maybe changed even the box is not overlapped
func (b *Cube) OverlapBox(x *Cube, area *Cube) bool {
	p1, p2 := b.Pos(), b.EndPos()
	q1, q2 := x.Pos(), x.EndPos()
	a1, a2 := q1.Subbed(p1), p2.Subbed(q1)
	b1, b2 := q2.Subbed(p1), p2.Subbed(q2)
	if a1.X >= 0 && a2.X >= 0 { // p1-q1-[q2]-p2-[q2]
		area.P.X = a1.X
		if b2.X >= 0 { // p1-q1-q2-p2
			area.S.X = x.S.X
		} else { // p1-q1-p2-q2
			area.S.X = a2.X
		}
	} else if b1.X >= 0 && b2.X >= 0 { // q1-p1-q2-p2
		area.P.X = 0
		area.S.X = b1.X
	} else {
		return false
	}
	if a1.Y >= 0 && a2.Y >= 0 {
		area.P.Y = a1.Y
		if b2.Y >= 0 {
			area.S.Y = x.S.Y
		} else {
			area.S.Y = a2.Y
		}
	} else if b1.Y >= 0 && b2.Y >= 0 {
		area.P.Y = 0
		area.S.Y = b1.Y
	} else {
		return false
	}
	if a1.Z >= 0 && a2.Z >= 0 {
		area.P.Z = a1.Z
		if b2.Z >= 0 {
			area.S.Z = x.S.Z
		} else {
			area.S.Z = a2.Z
		}
	} else if b1.Z >= 0 && b2.Z >= 0 {
		area.P.Z = 0
		area.S.Z = b1.Z
	} else {
		return false
	}
	return true
}
