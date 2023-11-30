package molecular

func (e *Engine) appendObjsInsideRange(objs []*Object, pos Vec3, radius float64) []*Object {
	radius2 := radius * radius
	for _, o := range e.objects {
		if o.AbsPos().Subbed(pos).SqLen() <= radius2 {
			objs = append(objs, o)
		}
	}
	return objs
}

func (e *Engine) appendObjsInsideRing(objs []*Object, pos Vec3, minR, maxR float64) []*Object {
	minR2 := minR * minR
	maxR2 := maxR * maxR
	for _, o := range e.objects {
		l := o.AbsPos().Subbed(pos).SqLen()
		if minR2 <= l && l <= maxR2 {
			objs = append(objs, o)
		}
	}
	return objs
}
