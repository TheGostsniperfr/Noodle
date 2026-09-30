package k8s

import (
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

var clusterScoped = set("Namespace", "ClusterRole", "ClusterRoleBinding", "CustomResourceDefinition",
	"StorageClass", "PersistentVolume", "PriorityClass", "IngressClass", "GatewayClass", "ClusterIssuer",
	"ClusterSecretStore", "MutatingWebhookConfiguration", "ValidatingWebhookConfiguration", "APIService",
	"Node", "RuntimeClass", "CSIDriver", "VolumeSnapshotClass", "ClusterPolicy")

// podSpecPath is where each workload kind keeps its pod spec.
var podSpecPath = map[string][]string{
	"Pod":         {"spec"},
	"Deployment":  {"spec", "template", "spec"},
	"StatefulSet": {"spec", "template", "spec"},
	"DaemonSet":   {"spec", "template", "spec"},
	"ReplicaSet":  {"spec", "template", "spec"},
	"Job":         {"spec", "template", "spec"},
	"CronJob":     {"spec", "jobTemplate", "spec", "template", "spec"},
}

// knownKinds are read by a rule, now or in a later task of spec 003, or cannot add an
// element or a connection to a diagram. Anything else is reported as unknown-kind:
// a CRD such as a CloudNativePG Cluster often stands for a real component.
var knownKinds = set(
	"Namespace", "Service", "ConfigMap", "Secret",
	"Ingress", "IngressClass", "Gateway", "GatewayClass", "HTTPRoute", "GRPCRoute", "TLSRoute", "TCPRoute", "ReferenceGrant",
	"NetworkPolicy", "ExternalSecret", "SecretStore", "ClusterSecretStore", "Application", "AppProject",
	"ServiceAccount", "Role", "RoleBinding", "ClusterRole", "ClusterRoleBinding",
	"PodDisruptionBudget", "HorizontalPodAutoscaler", "VerticalPodAutoscaler", "LimitRange", "ResourceQuota", "PriorityClass",
	"PersistentVolumeClaim", "PersistentVolume", "StorageClass", "VolumeSnapshotClass", "CSIDriver", "RuntimeClass",
	"CustomResourceDefinition", "MutatingWebhookConfiguration", "ValidatingWebhookConfiguration", "APIService",
	"ServiceMonitor", "PodMonitor", "PrometheusRule", "Probe",
	"Certificate", "Issuer", "ClusterIssuer",
	"ClusterPolicy", "Policy", "PolicyException", "ValidatingPolicy", "MutatingPolicy", "ImageValidatingPolicy",
	"ValidatingAdmissionPolicy", "ValidatingAdmissionPolicyBinding", "MutatingAdmissionPolicy", "MutatingAdmissionPolicyBinding",
	"ClientTrafficPolicy", "BackendTrafficPolicy", "EnvoyPatchPolicy", "EnvoyExtensionPolicy", "EnvoyProxy",
)

func isWorkload(kind string) bool { return podSpecPath[kind] != nil }

// unknownKinds is one entry per unknown kind with a count, never one per object, so a
// chart full of CRD instances costs one line (ADR-0017).
func unknownKinds(idx *index) []model.Unresolved {
	byKind := map[string]*model.Unresolved{}
	var order []string
	for _, o := range idx.objects {
		if knownKinds[o.Kind] || isWorkload(o.Kind) {
			continue
		}
		key := group(o.APIVersion) + "/" + o.Kind
		if u, ok := byKind[key]; ok {
			u.Count++
			continue
		}
		byKind[key] = &model.Unresolved{Kind: "unknown-kind", Value: key, Count: 1, Src: o.Src}
		order = append(order, key)
	}
	out := make([]model.Unresolved, 0, len(order))
	for _, k := range order {
		u := *byKind[k]
		if u.Count == 1 {
			u.Count = 0
		}
		out = append(out, u)
	}
	return out
}

// group is the API group of an apiVersion; the core group is "core".
func group(apiVersion string) string {
	g, _, ok := strings.Cut(apiVersion, "/")
	if !ok {
		return "core"
	}
	return g
}
