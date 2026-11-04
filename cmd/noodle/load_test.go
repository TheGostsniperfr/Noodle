package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func TestLoad_MergesFragments(t *testing.T) {
	tests := []struct {
		name, dir string
		want      []model.Port
		wantTech  string
	}{
		{"fills the ports the curated element leaves empty", "../../examples/discovery-k8s", []model.Port{{Name: "http", Protocol: "TCP", Number: 8080}}, "[Deployment]"},
		{"keeps the ports the curated element sets", "../../internal/model/testdata/minimal", []model.Port{{Name: "https", Protocol: "TCP", Number: 443}}, "[Deployment]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, g, err := loadSystem(tt.dir)

			require.NoError(t, err)
			node, ok := g.Lookup("api")
			require.True(t, ok)
			assert.Equal(t, tt.want, node.Element.Ports)
			assert.Equal(t, tt.wantTech, node.Element.Tech)
			var api model.Element
			for _, e := range s.Model.Elements {
				if e.ID == "api" {
					api = e
				}
			}
			assert.Equal(t, tt.want, api.Ports, "the system's own model is the merged one")
			assert.NotEmpty(t, g.Sources["api"])
		})
	}
}

func TestLoad_FailsOnACuratedModelThatDoesNotCheck(t *testing.T) {
	_, _, err := loadSystem("../../internal/model/testdata/broken")

	assert.ErrorContains(t, err, "check finding")
}
