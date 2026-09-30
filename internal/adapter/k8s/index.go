package k8s

// index keeps the objects in read order, which is path order, and finds them by kind.
type index struct {
	objects []*object
	byKind  map[string][]*object
}

func newIndex(objects []*object) *index {
	idx := &index{objects: objects, byKind: map[string][]*object{}}
	for _, o := range objects {
		idx.byKind[o.Kind] = append(idx.byKind[o.Kind], o)
	}
	return idx
}
