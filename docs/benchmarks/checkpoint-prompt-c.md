The directory `discovered/` holds noodle discovery fragments, one per Argo CD
Application, generated statically from the rendered charts (`noodle discover k8s`).
Each lists elements, connections and references with their source, and an
`unresolved` list of what the static pass could not decide. Start from them, and read
the repositories only for what they do not answer.
