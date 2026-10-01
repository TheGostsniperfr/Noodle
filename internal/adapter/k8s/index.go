package k8s

// index keeps the objects in read order, which is path order, and finds them by kind
// or by identity.
type index struct {
	objects []*object
	byKind  map[string][]*object
	byID    map[string]*object
}

func newIndex(objects []*object) *index {
	idx := &index{objects: objects, byKind: map[string][]*object{}, byID: map[string]*object{}}
	for _, o := range objects {
		idx.byKind[o.Kind] = append(idx.byKind[o.Kind], o)
		if _, dup := idx.byID[o.id()]; !dup {
			idx.byID[o.id()] = o
		}
	}
	return idx
}

func (idx *index) get(ns, kind, name string) *object { return idx.byID[id(ns, kind, name)] }

// workloadsIn returns the workloads of a namespace in read order.
func (idx *index) workloadsIn(ns string) []*object {
	var out []*object
	for _, o := range idx.objects {
		if isWorkload(o.Kind) && o.Namespace == ns {
			out = append(out, o)
		}
	}
	return out
}

// podLabels are the labels a workload gives its pods, which Services select.
func podLabels(o *object) map[string]string {
	if o.Kind == "Pod" {
		return o.Labels
	}
	path := podSpecPath[o.Kind]
	return strMap(o.Body, append(append([]string{}, path[:len(path)-1]...), "metadata", "labels")...)
}
