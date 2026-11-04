# How to draw a proposal as a diff

A proposal is a copy of a view with the changes marked, over the same model. Reviewers see
what is added and what is removed against the platform they already know.

1. In `model.yaml`, mark what the proposal adds with `status: planned` and what it removes
   with `status: deprecated`. Give each a `target`, the phase or date it lands:

   ```yaml
   - {id: loki, kind: database, title: Loki, status: planned, target: SP3, ...}
   - {id: elastic, kind: database, title: Elasticsearch, status: deprecated, ...}
   ```

   A connection takes its status from its ends. Set `status` on a connection only to mark
   a connection that is added or removed between two elements that stay.

2. Copy the reference view and its layout to `views/proposal.yaml` and
   `layouts/proposal.yaml`, then add `state: diff` to the view.

3. Render it:

   ```bash
   noodle render examples/topology-diff -view proposal -theme dark -o proposal.dark.drawio
   ```

Added items are drawn with a change-coloured frame, a `+ target` pill and a `+` label.
Removed items have a red dotted frame and a struck title. Everything else is dimmed.

To spotlight one area of a large proposal, add `highlight: [g-obs/**]` to the view: those
elements and every edge touching them stay in full colour.

When the plan ships, drop the statuses. The reference and the proposal then both show the
new platform. The worked example is [`examples/topology-diff`](../../examples/topology-diff).
