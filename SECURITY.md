# Security policy

## Reporting a vulnerability

Report it privately at <https://github.com/TheGostsniperfr/Noodle/security/advisories/new>.
Do not open a public issue or pull request for it. Noodle has one maintainer, so please
allow some days for a first answer.

## Why an architecture map is sensitive

Noodle models real systems. A model of a platform lists its services, ports, trust
boundaries, secrets references and known gaps, which is what an attacker maps first.
Treat these as sensitive when you report:

- a model, view or discovered fragment taken from a real system: redact names, hosts and
  ports before attaching it, or reproduce with an invented system;
- a bug that makes `noodle discover` read or print a secret value. Discovery is
  read-only and must never output a secret value (C09).

## What is in scope

- The `noodle` binary and `bin/noodle`, including the download of a release binary and
  its checksum verification.
- `noodle setup`, which writes files into a repository or a home directory.
- The skills and agents that ship with the plugin.

## Supported versions

Only the latest release and `main` receive fixes.
