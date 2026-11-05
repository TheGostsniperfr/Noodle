package k8s

import (
	"sort"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// unresolvedSet merges entries of the same kind and value, so a value met by twelve
// workloads costs one line with twelve abouts (ADR-0017).
type unresolvedSet struct {
	byKey map[string]*model.Unresolved
	order []string
}

func (u *unresolvedSet) add(kind, value, hint, about string, src model.Src) {
	u.addKeyed(kind+" "+value, kind, value, hint, about, src)
}

// addKeyed merges on key instead of kind and value, for entries whose value repeats
// across distinct objects, such as a Secret name in two namespaces.
func (u *unresolvedSet) addKeyed(key, kind, value, hint, about string, src model.Src) {
	if u.byKey == nil {
		u.byKey = map[string]*model.Unresolved{}
	}
	e, ok := u.byKey[key]
	if !ok {
		e = &model.Unresolved{Kind: kind, Value: value, Hint: hint, Src: src}
		u.byKey[key] = e
		u.order = append(u.order, key)
	}
	for _, a := range e.About {
		if a == about {
			return
		}
	}
	e.About = append(e.About, about)
}

func (u *unresolvedSet) list() []model.Unresolved {
	out := make([]model.Unresolved, 0, len(u.order))
	for _, k := range u.order {
		e := *u.byKey[k]
		sort.Strings(e.About)
		out = append(out, e)
	}
	return out
}
