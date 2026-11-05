package k8s

import (
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

var addressWords = []string{"HOST", "URL", "URI", "ENDPOINT", "DSN", "JDBC", "ADDR", "SERVER"}

// credentialWords win over addressWords: SMTP_SERVER_PASSWORD names a credential.
var credentialWords = []string{"PASSWORD", "PASSWD", "TOKEN", "SECRET"}

func isAddressKey(key string) bool {
	k := strings.ToUpper(key)
	for _, w := range credentialWords {
		if strings.Contains(k, w) {
			return false
		}
	}
	for _, w := range addressWords {
		if strings.Contains(k, w) {
			return true
		}
	}
	return false
}

// secretEndpoints emits one secret-endpoint entry per Secret whose address-like keys a
// workload reads through env or envFrom (FR-017). The address itself lives in the secret
// manager; naming the keys and readers tells the agent where to look, so it does not
// read the repository for it. Only key names are read, never data values.
func secretEndpoints(idx *index, uses []configUse, u *unresolvedSet) {
	type found struct {
		name    string
		keys    map[string]bool
		readers []string
		first   configUse
	}
	bySecret := map[string]*found{}
	var order []string
	note := func(use configUse, name string, keys []string) {
		for _, k := range keys {
			if !isAddressKey(k) {
				continue
			}
			f, ok := bySecret[use.target]
			if !ok {
				f = &found{name: name, keys: map[string]bool{}, first: use}
				bySecret[use.target] = f
				order = append(order, use.target)
			}
			f.keys[k] = true
			if !slices.Contains(f.readers, use.workload.id()) {
				f.readers = append(f.readers, use.workload.id())
			}
		}
	}
	for _, use := range uses {
		if !strings.Contains(use.target, "/secret/") {
			continue
		}
		name := use.target[strings.LastIndex(use.target, "/")+1:]
		switch use.kind {
		case "secret-key":
			note(use, name, []string{use.key})
		case "env-from":
			note(use, name, secretKeys(idx, use.workload.Namespace, name))
		}
	}
	for _, target := range order {
		f := bySecret[target]
		keys := sortedKeys(f.keys)
		value := f.name + "/" + strings.Join(keys, ",")
		for _, r := range f.readers {
			u.addKeyed("secret-endpoint "+target, "secret-endpoint", value, "secret-manager", r, f.first.workload.srcAt(f.first.line))
		}
	}
}

// secretKeys are the key names a Secret is known to hold: from the Secret itself when it
// is in the input, and from every ExternalSecret that writes it.
func secretKeys(idx *index, ns, name string) []string {
	keys := map[string]bool{}
	if s := idx.get(ns, "Secret", name); s != nil {
		for _, field := range []string{"data", "stringData"} {
			addKeys(keys, mapValue(s.Node, field))
		}
	}
	for _, es := range idx.byKind["ExternalSecret"] {
		if es.Namespace != ns || or(text(at(es.Node, "spec", "target", "name")), es.Name) != name {
			continue
		}
		for _, d := range seq(at(es.Node, "spec", "data")) {
			if k := text(mapValue(d, "secretKey")); k != "" {
				keys[k] = true
			}
		}
		addKeys(keys, at(es.Node, "spec", "target", "template", "data"))
	}
	return sortedKeys(keys)
}

// addKeys reads the keys of a mapping, never its values.
func addKeys(keys map[string]bool, m *yaml.Node) {
	if m == nil || m.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		keys[m.Content[i].Value] = true
	}
}
