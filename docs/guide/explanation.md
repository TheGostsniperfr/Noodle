# How noodle thinks about a system

noodle draws architecture as code. You describe a system once, and each diagram is a
question asked of that description. Four ideas carry the rest.

## Model

The model, `model.yaml`, says what exists: the elements (components and zones), the
connections between them and the references. It is written once and never repeated. An
element has a stable id that is never renamed for looks, a kind, a title and, where they
matter, ports. Everything else refers to elements by id.

## View

A view says what one diagram shows: which elements, in which order the connections are
numbered, what the title asks and which cards sit under the drawing. One model has many
views: the runtime topology, a login sequence, a tech stack, a service catalogue, a table
of who may act on what. A view never defines an element: it picks from the model.

## Layout

A layout says where a topology view puts its boxes and how its lines run. It is a separate
file because placing boxes is craft that changes for reasons unrelated to what exists.
Only topology views have one. Sequence, landscape, catalog and matrix views are computed.
noodle checks the result and refuses to write a diagram whose lines cross boxes or whose
labels collide.

## Fragment

A fragment is what a tool found, as opposed to what a person wrote. `noodle discover` reads
manifests and writes a fragment with the file and line of every fact. It never edits the
model. A curated element links to discovered objects with `matches`, and the gap between
the two is something to report, not to hide. Fragments are generated and not committed.

## Connection semantics

An arrow is a connection. It points from the side that opens it, the client, to the side
that listens, the server, and the listening port is on the server. This holds even when
the data flows the other way: a tunnel opened from inside the cluster to a proxy outside
is drawn from the cluster to the proxy, and the label says what is relayed back.

Draw requests, not responses. A response is a note, unless it changes the path.

A reference is not traffic. A pod that mounts a secret or a route that names a gateway is
configuration, drawn dotted with an open arrow, never numbered.

Topology and sequence are different views. A topology has one numbered step per
connection. A flow that reuses a connection or goes back and forth, such as a login with
redirects, is a sequence view.

The full rules are in [ADR-0003](../adr/0003-connection-semantics.md) and
[ADR-0004](../adr/0004-topology-vs-sequence.md), and the file formats are in the
[skill reference](../../skills/noodle-diagram/reference.md).
