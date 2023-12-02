package molecular

type set[T comparable] map[T]struct{}

func (s set[T]) Put(v T) {
	s[v] = struct{}{}
}

func (s set[T]) Has(v T) (ok bool) {
	_, ok = s[v]
	return
}

func (s set[T]) Remove(v T) (ok bool) {
	_, ok = s[v]
	delete(s, v)
	return
}

func (s set[T]) Clear() {
	clear(s)
}

func (s set[T]) AsSlice() (a []T) {
	a = make([]T, 0, len(s))
	for i := range s {
		a = append(a, i)
	}
	return
}

func (s set[T]) Clone() (o set[T]) {
	o = make(set[T], len(s))
	for a := range s {
		o.Put(a)
	}
	return
}
