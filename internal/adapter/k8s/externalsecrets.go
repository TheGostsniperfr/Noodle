package k8s

import (
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// providers are the secret manager backends whose element is worth drawing: where a
// secret really lives (ADR-0018).
var providers = map[string]struct{ slug, tech string }{
	"onepassword":    {"onepassword", "1Password"},
	"onepasswordSDK": {"onepassword", "1Password"},
	"vault":          {"vault", "Vault"},
	"aws":            {"aws", "AWS"},
}

func providerID(slug string) string { return "k8s:_external/" + slug }

// externalSecrets links each ExternalSecret to the Secret it writes and to its store, and
// each store to the provider element behind it (FR-002).
func externalSecrets(idx *index, g *graph, u *unresolvedSet) []model.DiscoveredElement {
	for _, es := range idx.byKind["ExternalSecret"] {
		target := or(text(at(es.Node, "spec", "target", "name")), es.Name)
		g.refer(es.id(), id(es.Namespace, "Secret", target), "target", es,
			es.srcAt(firstLine(at(es.Node, "spec", "target", "name"), at(es.Node, "spec", "target"))))
		storeRef := at(es.Node, "spec", "secretStoreRef")
		if name := text(mapValue(storeRef, "name")); name != "" {
			store := id(es.Namespace, "SecretStore", name)
			if text(mapValue(storeRef, "kind")) == "ClusterSecretStore" {
				store = id("", "ClusterSecretStore", name)
			}
			g.refer(es.id(), store, "store", es, es.srcAt(firstLine(mapValue(storeRef, "name"), storeRef)))
		}
	}
	var out []model.DiscoveredElement
	made := map[string]bool{}
	for _, kind := range []string{"SecretStore", "ClusterSecretStore"} {
		for _, store := range idx.byKind[kind] {
			provider := at(store.Node, "spec", "provider")
			if provider == nil || len(provider.Content) < 2 {
				continue
			}
			key, value := provider.Content[0].Value, provider.Content[1]
			p, known := providers[key]
			if !known {
				u.add("unknown-provider", key, "secretstore", store.id(), store.srcAt(provider.Content[0].Line))
				continue
			}
			g.refer(store.id(), providerID(p.slug), "provider", store, store.srcAt(firstLine(provider.Content[0], value)))
			if !made[p.slug] {
				made[p.slug] = true
				out = append(out, model.DiscoveredElement{Element: model.Element{ID: providerID(p.slug), Kind: "external", Tech: p.tech},
					Src: store.srcAt(provider.Content[0].Line)})
			}
		}
	}
	return out
}
