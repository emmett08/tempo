# Tempo 3.0.3 block bounds fixes

Branch `release/3.0.3` preserves the Tempo 3.0.3 runtime and applies
the block metadata and metrics-sharding fixes. Base source:
`1900ed7bb5cad1a3edc285783d7d4ac4278337dc`; fix commit:
`2ebbd8f5a97607bc45aac2550155db9bbacef8c0`.

Pull requests into the release branch run verification without publishing. A
normal merge triggers the image workflow from the merged source SHA. The
separate `main` line tracks newer upstream development; fixes there are ported
through a main-based PR.

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

Release CI installs the locked Jsonnet dependencies when checking generated
examples. Updating them against upstream heads can remove the historical
Kubernetes library version required by this release; dependency updates remain
an explicit maintenance action.

The release also pins Grafana's e2e test helper at
`cfcdc6861c9cbf097547e905af8251a40f747350`, which replaces the removed MinIO
community image with the Silo test image and command. This restores the S3
integration fixtures without changing Tempo's runtime dependency versions.

Upstream fix `f97d7856253481b24b8c3ed207c17676a2fbe7c9` also pins the
compaction selector's clock in its one-second-window test. The production clock
still defaults to `time.Now`; the test no longer depends on staying inside the
same wall-clock second while the suite runs.

The scheduler recovery test also serializes its local-backend read with the
running scheduler's existing cache-flush mutex. The filesystem fixture truncates
and rewrites the cache file, so an overlapping recovery read could see incomplete
JSON. This test-only synchronization does not change production cache behavior.
