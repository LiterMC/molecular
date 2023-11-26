package molecular

type AnchorObject struct {
	rpos map[*AnchorObject]Vec // the pos relative to other known anchors
	objs map[*Object]struct{}
}

func (a *AnchorObject) Tick(dt float64) {
	for o, _ := range a.objs {
		o.Tick(dt)
	}
}

type Object struct {
	anchor *AnchorObject
	pos    Vec // the position relative to the anchor
	facing Vec
	blocks []Block
}

func (o *Object) Tick(dt float64) {
	for _, b := range o.blocks {
		b.Tick(dt)
	}
}
