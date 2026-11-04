package k8s

import (
	"gopkg.in/yaml.v3"
)

// configUse is one place a workload reads a Secret, a ConfigMap or a claim: the object
// it reads, how, and which key when it reads one key only.
type configUse struct {
	workload *object
	target   string
	kind     string
	key      string
	line     int
}

// configRefs emits one reference per workload, target and way of reading (FR-001). It
// never looks at a Secret's data: only the names that connect a workload to it matter.
func configRefs(idx *index, g *graph) []configUse {
	var uses []configUse
	for _, w := range idx.objects {
		if !isWorkload(w.Kind) {
			continue
		}
		for _, u := range workloadUses(w) {
			g.refer(w.id(), u.target, u.kind, w, w.srcAt(u.line))
			uses = append(uses, u)
		}
	}
	return uses
}

func workloadUses(w *object) []configUse {
	var uses []configUse
	use := func(kind, targetKind string, name *yaml.Node, key string, ref *yaml.Node) {
		if n := text(name); n != "" {
			uses = append(uses, configUse{workload: w, target: id(w.Namespace, targetKind, n), kind: kind, key: key, line: firstLine(name, ref)})
		}
	}
	for _, c := range w.containers() {
		for _, from := range seq(mapValue(c, "envFrom")) {
			if ref := mapValue(from, "configMapRef"); ref != nil {
				use("env-from", "ConfigMap", mapValue(ref, "name"), "", ref)
			}
			if ref := mapValue(from, "secretRef"); ref != nil {
				use("env-from", "Secret", mapValue(ref, "name"), "", ref)
			}
		}
		for _, env := range seq(mapValue(c, "env")) {
			from := mapValue(env, "valueFrom")
			if ref := mapValue(from, "secretKeyRef"); ref != nil {
				use("secret-key", "Secret", mapValue(ref, "name"), text(mapValue(ref, "key")), ref)
			}
			if ref := mapValue(from, "configMapKeyRef"); ref != nil {
				use("config-key", "ConfigMap", mapValue(ref, "name"), text(mapValue(ref, "key")), ref)
			}
		}
	}
	for _, vol := range seq(mapValue(w.podSpecNode(), "volumes")) {
		if ref := mapValue(vol, "secret"); ref != nil {
			use("mounts", "Secret", mapValue(ref, "secretName"), "", ref)
		}
		if ref := mapValue(vol, "configMap"); ref != nil {
			use("mounts", "ConfigMap", mapValue(ref, "name"), "", ref)
		}
		if ref := mapValue(vol, "persistentVolumeClaim"); ref != nil {
			use("mounts", "PersistentVolumeClaim", mapValue(ref, "claimName"), "", ref)
		}
		for _, source := range seq(at(vol, "projected", "sources")) {
			if ref := mapValue(source, "secret"); ref != nil {
				use("mounts", "Secret", mapValue(ref, "name"), "", ref)
			}
			if ref := mapValue(source, "configMap"); ref != nil {
				use("mounts", "ConfigMap", mapValue(ref, "name"), "", ref)
			}
		}
	}
	return uses
}
