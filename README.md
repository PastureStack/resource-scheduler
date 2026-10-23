# PastureStack Resource Scheduler

Resource Scheduler is the control-plane daemon that tracks host capacity, labels, and host-port reservations. It consumes the dated metadata API, handles scheduling events, ranks eligible hosts, reserves and releases resources, and exposes the health endpoint used by the infrastructure catalog.

PastureStack is an independent community effort to preserve, audit, and modernize the Rancher 1.6 ecosystem. It is not affiliated with or endorsed by Rancher Labs or SUSE.

**Upstream:** [`rancher/scheduler`](https://github.com/rancher/scheduler). This GitHub fork retains the upstream Git history, authorship, dates, and license notices. PastureStack maintenance is consolidated into one commit after the preserved upstream boundary.

The current published release is [`v0.8.20`](https://github.com/PastureStack/resource-scheduler/releases/tag/v0.8.20).
The corresponding Catalog version `5` is recorded in the
[Catalog update](https://github.com/PastureStack/catalog-templates/pull/18).

Release `v0.8.20` refreshes the official Go security release and Ubuntu packages.
Its actual runtime scan has 0 High/Critical findings, 0 vendor-fixed findings,
and 0 secrets. The 4 Low and 47 Medium vendor-unfixed findings remain
`under_investigation` in the release risk register and OpenVEX; they are not
resolved. Release checksums and the SBOM and image build-provenance attestations
were verified against the exact signed tag, source commit, and release workflow.
Publication does not establish a managed-system upgrade or rollback result.

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

The [native evidence run](https://github.com/PastureStack/resource-scheduler/actions/runs/39357570483)
retains the raw builder scan: 8 Critical and 192 High header source-package
matches are separately classified by 200 exact `not_affected` statements.
Another 2,600 vendor-unfixed package findings remain recorded, including the
unrated `GO-2026-5932` module advisory. The actual Buildx compilation graph
contains no `golang.org/x/crypto/openpgp` packages; that fact does not remove
the module finding or assess any host kernel.

## Runtime image

The maintained image is published from this repository as:

```text
ghcr.io/pasturestack/resource-scheduler:<version>
```

This image is a system component, not a standalone application. The PastureStack infrastructure catalog supplies the control-plane URL, scoped access credentials, metadata endpoint, health-check configuration, and an explicit numeric version tag. The container runs as UID/GID `10001:10001`; only the scheduler binary receives `cap_net_bind_service` so the established port `80` health contract can remain available without a root process.

`/healthcheck` is a process-liveness endpoint. `/readiness` separately verifies the metadata and control-plane dependencies. A temporary control-plane outage therefore does not cause the scheduler container to be replaced. The metadata watcher and event subscriber reconnect after a clean disconnect, an error, or a recovered watcher panic.

The native command is `resource-scheduler`. A `scheduler` executable symlink remains temporarily available for catalog compatibility. The default metadata service name is `metadata`; deployments can set `PASTURESTACK_METADATA_ADDRESS` or pass `--metadata-address` explicitly.

## Build and test

The maintained build uses Go 1.27.3 to compile Docker CLI 29.7.2 and Docker Buildx 0.36.1 from checksum-pinned source. A small checksum-locked Buildx patch removes its sole compiled dependency on the legacy Docker module; the security-sensitive Buildx modules are then pinned to `github.com/moby/go-archive` 0.3.0 and `golang.org/x/mod` 0.41.0, with the final binary metadata checked in CI. The Ubuntu base image is digest-pinned; all inherited base packages are updated from the fixed `20261020T000000Z` Canonical snapshot before the direct packages pinned in `ubuntu-apt.lock` are installed, and each built image records the complete resolved `dpkg` inventory. BuildKit receives the source commit time through `SOURCE_DATE_EPOCH`, and CI compares binary hashes, runtime configuration, and resolved package inventories across clean rebuilds.

The dependency graph is declared in `go.mod`, checksum-bound by `go.sum`, and committed in the standard module-aware `vendor` tree for reproducible offline builds. Security CI produces short-lived source and runtime CycloneDX SBOMs, runs binary reachability analysis, and blocks detected secrets, every High/Critical runtime finding, and any finding for which Ubuntu has published a fixed package. Low/Medium findings without an available Ubuntu fix remain visible as `under_investigation` in the release risk register and OpenVEX artifact; they are not represented as resolved and are re-evaluated when the vendor, base digest, or vulnerability data changes.

```bash
make test
make validate
bash scripts/check-build-downloads
go list -mod=vendor ./...
bash scripts/check-migration-policy
VERSION_OVERRIDE=v0.8.20 IMAGE_NAMESPACE=pasturestack make package
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

## October 2026 security maintenance

Maintained Go builds select the patched Go 1.27.3 compiler without raising the existing Go language requirement. Ubuntu build/runtime stages apply the checksum-pinned `ca-certificates 20260816ubuntu0.26.04.1` security package after their frozen APT baseline; the final trust store and installed version are verified. Release artifacts must pass the repository's existing checks before publication.
