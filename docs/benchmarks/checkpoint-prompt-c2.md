The file `digest.txt` is the output of `noodle digest` on discovery fragments generated
statically from the rendered charts, one line per fact. It is complete for the
connections between components inside the cluster, routes from the gateway included.
Each `secret-endpoint` line names the Secret keys that hold an external address and the
workloads that read them; the values live in the secret manager, never in the
repositories. Its `src` paths point to the rendered charts, which are not here. Start
from it, and read the repositories only for what it does not answer.
