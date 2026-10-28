# Tempo 3.0.3 block bounds fixes

Branch `fix/3.0.3-block-time-bounds` preserves the Tempo 3.0.3 runtime and applies
the block metadata and metrics-sharding fixes. Base source:
`1900ed7bb5cad1a3edc285783d7d4ac4278337dc`; fix commit:
`2ebbd8f5a97607bc45aac2550155db9bbacef8c0`.

The fork image workflow runs the seven ingestion and five frontend regression
cases plus adjacent timestamp/exemplar tests with the race detector. It builds
Linux amd64 using the original pinned Dockerfile, checks `--version`, and
publishes `ghcr.io/emmett08/tempo:git-<full-source-sha>` with source/revision labels.
Consumers should additionally pin the published manifest digest. The source
revision and fork version are also embedded in the Tempo binary.

The writer clamps both endpoints inside the ingestion window and preserves
ordered bounds for entirely old/future traces. The frontend rejects invalid
metadata and nonpositive query intervals before constructing a shard request;
valid zero-duration block metadata remains searchable. The warning diagnostic
is retained when a bad block is skipped.

An ingestion allowance matching the intended historical/replay window is still
needed to preserve original timestamps. The patch does not rewrite previously
stored invalid block metadata. Repair those blocks separately from backups,
using stored span timestamps, and verify data coverage rather than relying only
on a successful HTTP status.
