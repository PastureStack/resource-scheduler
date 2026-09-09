# PastureStack Resource Scheduler

Resource Scheduler is the control-plane daemon that tracks host capacity, labels, and host-port reservations. It consumes the dated metadata API, handles scheduling events, ranks eligible hosts, reserves and releases resources, and exposes the health endpoint used by the infrastructure catalog.

PastureStack is an independent community effort to preserve, audit, and modernize the Rancher 1.6 ecosystem. It is not affiliated with or endorsed by Rancher Labs or SUSE.

**Upstream:** [`rancher/scheduler`](https://github.com/rancher/scheduler). This GitHub fork retains the upstream Git history, authorship, dates, and license notices. PastureStack maintenance is consolidated into one commit after the preserved upstream boundary.

The current public release and Catalog image are `v0.8.17`.

## Runtime image

The maintained image is published from this repository as:

```text
ghcr.io/pasturestack/resource-scheduler:<version>
```

This image is a system component, not a standalone application. The PastureStack infrastructure catalog supplies the control-plane URL, scoped access credentials, metadata endpoint, health-check configuration, and an explicit numeric version tag. The container runs as UID/GID `10001:10001`; only the scheduler binary receives `cap_net_bind_service` so the established port `80` health contract can remain available without a root process.

`/healthcheck` is a process-liveness endpoint. `/readiness` separately verifies the metadata and control-plane dependencies. A temporary control-plane outage therefore does not cause the scheduler container to be replaced. The metadata watcher and event subscriber reconnect after a clean disconnect, an error, or a recovered watcher panic.

The native command is `resource-scheduler`. A `scheduler` executable symlink remains temporarily available for catalog compatibility. The default metadata service name is `metadata`; deployments can set `PASTURESTACK_METADATA_ADDRESS` or pass `--metadata-address` explicitly.

## Build and test

The reviewed build uses Go 1.27.0 to compile Docker CLI 29.7.2 and Docker Buildx 0.36.1 from checksum-pinned source. A small checksum-locked Buildx patch removes its sole compiled dependency on the legacy Docker module; the two security-sensitive Buildx modules are then pinned to `github.com/moby/go-archive` 0.3.0 and `golang.org/x/mod` 0.40.0, with the final binary metadata checked in CI. The Ubuntu base image is digest-pinned; direct packages are version-pinned in `ubuntu-apt.lock` against the fixed `20260909T000000Z` Canonical snapshot, and each built image records the complete resolved `dpkg` inventory. BuildKit receives the source commit time through `SOURCE_DATE_EPOCH`, and CI rejects differing binary hashes or image IDs across clean rebuilds.

The dependency graph is declared in `go.mod`, checksum-bound by `go.sum`, and committed in the standard module-aware `vendor` tree for reproducible offline builds. Security CI produces short-lived source and runtime CycloneDX SBOMs, runs binary reachability analysis, and blocks runtime Critical or High vulnerabilities and detected secrets.

```bash
make test
make validate
bash scripts/check-build-downloads
go list -mod=vendor ./...
bash scripts/check-migration-policy
VERSION_OVERRIDE=v0.8.17 IMAGE_NAMESPACE=pasturestack make package
```

CI validates source and dependency locks, tests and reproducible builds, and generates short-lived security evidence. Releases are published only from an annotated, pure numeric SemVer tag that resolves to the reviewed commit.

The scheduling test suite includes repeated allocation of the same host-port
reservation. Retries for one resource UUID are idempotent, while a different
resource UUID requesting the same host port remains rejected.

## Compatibility and security

Some event names, environment variables, API client type names, dependency namespaces, and scheduling labels are protocol or data contracts. They are isolated and documented in [COMPATIBILITY.md](COMPATIBILITY.md), rather than exposed as PastureStack branding.

Review [SECURITY.md](SECURITY.md) before changing credential handling, metadata trust, custom CA verification, capabilities, or event subscriptions.

## License and attribution

The inherited project remains licensed under [Apache License 2.0](LICENSE). PastureStack does not claim authorship of inherited work. See [ORIGIN.md](ORIGIN.md) and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for provenance and bundled dependency notices.
