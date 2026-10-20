# PastureStack Resource Scheduler

Resource Scheduler is the control-plane daemon that tracks host capacity, labels, and host-port reservations. It consumes the dated metadata API, handles scheduling events, ranks eligible hosts, reserves and releases resources, and exposes the health endpoint used by the infrastructure catalog.

PastureStack is an independent community effort to preserve, audit, and modernize the Rancher 1.6 ecosystem. It is not affiliated with or endorsed by Rancher Labs or SUSE.

**Upstream:** [`rancher/scheduler`](https://github.com/rancher/scheduler). This GitHub fork retains the upstream Git history, authorship, dates, and license notices. PastureStack maintenance is consolidated into one commit after the preserved upstream boundary.

The current published release and Catalog image are `v0.8.18`.

The `v0.8.19` candidate refreshes the official Go security release and Ubuntu packages. It is not a published release or a Catalog deployment until its required checks and release workflow complete.

Its disposable builder also selects the official `golang.org/x/net` 0.60.0,
OpenTelemetry-Go 1.45.0, and containerd 2.3.6 security updates for Docker CLI
and Buildx. These tools are not copied into the runtime image.

Builder kernel-header matches require an exact per-CVE applicability review,
official package archive hash, and matching installed-file inventory before CI
emits `not_affected` OpenVEX. Raw scan findings remain available. New CVEs,
package or content drift, unrelated High/Critical findings, and every
vendor-fixed builder finding remain blocking. This does not assess the host
kernel; vendor-unfixed Low/Medium and unrated builder findings retain their
original severity and `under_investigation` status.

## Runtime image

The maintained image is published from this repository as:

```text
ghcr.io/pasturestack/resource-scheduler:<version>
```

This image is a system component, not a standalone application. The PastureStack infrastructure catalog supplies the control-plane URL, scoped access credentials, metadata endpoint, health-check configuration, and an explicit numeric version tag. The container runs as UID/GID `10001:10001`; only the scheduler binary receives `cap_net_bind_service` so the established port `80` health contract can remain available without a root process.

`/healthcheck` is a process-liveness endpoint. `/readiness` separately verifies the metadata and control-plane dependencies. A temporary control-plane outage therefore does not cause the scheduler container to be replaced. The metadata watcher and event subscriber reconnect after a clean disconnect, an error, or a recovered watcher panic.

The native command is `resource-scheduler`. A `scheduler` executable symlink remains temporarily available for catalog compatibility. The default metadata service name is `metadata`; deployments can set `PASTURESTACK_METADATA_ADDRESS` or pass `--metadata-address` explicitly.

## Build and test

The candidate build uses Go 1.27.2 to compile Docker CLI 29.7.2 and Docker Buildx 0.36.1 from checksum-pinned source. A small checksum-locked Buildx patch removes its sole compiled dependency on the legacy Docker module; the security-sensitive Buildx modules are then pinned to `github.com/moby/go-archive` 0.3.0 and `golang.org/x/mod` 0.41.0, with the final binary metadata checked in CI. The Ubuntu base image is digest-pinned; all inherited base packages are updated from the fixed `20261020T000000Z` Canonical snapshot before the direct packages pinned in `ubuntu-apt.lock` are installed, and each built image records the complete resolved `dpkg` inventory. BuildKit receives the source commit time through `SOURCE_DATE_EPOCH`, and CI compares binary hashes, runtime configuration, and resolved package inventories across clean rebuilds.

The dependency graph is declared in `go.mod`, checksum-bound by `go.sum`, and committed in the standard module-aware `vendor` tree for reproducible offline builds. Security CI produces short-lived source and runtime CycloneDX SBOMs, runs binary reachability analysis, and blocks detected secrets, every High/Critical runtime finding, and any finding for which Ubuntu has published a fixed package. Low/Medium findings without an available Ubuntu fix remain visible as `under_investigation` in the release risk register and OpenVEX artifact; they are not represented as resolved and are re-evaluated when the vendor, base digest, or vulnerability data changes.

```bash
make test
make validate
bash scripts/check-build-downloads
go list -mod=vendor ./...
bash scripts/check-migration-policy
VERSION_OVERRIDE=v0.8.19 IMAGE_NAMESPACE=pasturestack make package
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
