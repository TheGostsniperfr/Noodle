package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

func loadSequence(t *testing.T, dir, view string) *diagram.Sequence {
	t.Helper()
	s, err := model.LoadSystem(dir)
	require.NoError(t, err)
	require.Empty(t, model.Check(s))
	seq, err := resolve.Sequence(s, view)
	require.NoError(t, err)
	return seq
}

func TestSequence_NumbersMessagesAndReplies_NotNotes(t *testing.T) {
	t.Parallel()

	seq := loadSequence(t, "../../examples/sequence-basics", "request")

	var labels []string
	for _, m := range seq.Messages {
		labels = append(labels, m.Label)
	}
	assert.Equal(t, []string{"[1] GET /", "[2] relay GET / ◀ over the tunnel", "[3] forward · HTTP", "[4] 200 · HTML", "[5]", "[6] 200"}, labels)
	assert.Equal(t, []string{"render page"}, []string{seq.Notes[0].Text})
}

func TestSequence_ReplyGoesBackOverItsRequest_AndIsDashed(t *testing.T) {
	t.Parallel()

	seq := loadSequence(t, "../../examples/sequence-basics", "request")

	relay, reply := seq.Messages[1], seq.Messages[4]
	assert.False(t, relay.Reply, "a request against its connection is still a request")
	assert.Equal(t, "tunnel", relay.Kind)
	assert.True(t, reply.Reply)
	assert.Equal(t, [2]float64{relay.ToX, relay.FromX}, [2]float64{reply.FromX, reply.ToX})
}

func TestSequence_OrdersColumnsByParticipantsThenFirstAppearance(t *testing.T) {
	t.Parallel()

	seq := loadSequence(t, "../../examples/cnp-runtime", "oidc-login")

	var ids []string
	for _, n := range seq.Frame.Nodes {
		ids = append(ids, n.ID)
	}
	assert.Equal(t, []string{"user", "cf-edge", "cloudflared-app", "cloudflared-platform", "envoy-shared-gateway", "keycloak"}, ids)
}

func TestSequence_EveryLabelFitsTheSpanItCrosses(t *testing.T) {
	t.Parallel()

	seq := loadSequence(t, "../../examples/cnp-runtime", "oidc-login")

	for _, m := range seq.Messages {
		span := m.ToX - m.FromX
		if span < 0 {
			span = -span
		}
		need := house.TextWidth(house.PlainText(m.Label), house.EdgeFontSize) + 2*house.LabelPadX
		assert.GreaterOrEqual(t, span, need, m.ID)
	}
}

func TestSequence_CarriesGapsOnParticipantsAndMessages(t *testing.T) {
	t.Parallel()

	seq := loadSequence(t, "../../examples/cnp-runtime", "oidc-login")

	badges := map[string]string{}
	for _, n := range seq.Frame.Nodes {
		badges[n.ID] = n.Badge
	}
	assert.Equal(t, "G4 G6", badges["envoy-shared-gateway"])
	assert.Equal(t, "[18] POST auth.3istor.com /token · !!⚠ G6!!", seq.Messages[17].Label)
}
