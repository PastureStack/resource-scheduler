# PastureStack Resource Scheduler

Resource Scheduler is the control-plane daemon that tracks host capacity, labels, and host-port reservations. It consumes the dated metadata API, handles scheduling events, ranks eligible hosts, reserves and releases resources, and exposes the health endpoint used by the infrastructure catalog.

PastureStack is an independent community effort to preserve, audit, and modernize the Rancher 1.6 ecosystem. It is not affiliated with or endorsed by Rancher Labs or SUSE.

**Upstream:** [`rancher/scheduler`](https://github.com/rancher/scheduler). This GitHub fork retains the upstream Git history, authorship, dates, and license notices. PastureStack maintenance is consolidated into one commit after the preserved upstream boundary.

## Runtime image

The maintained image is published from this repository as:

```text
ghcr.io/pasturestack/resource-scheduler:<version>
```

This image is a system component, not a standalone application. The PastureStack infrastructure catalog supplies the control-plane URL, scoped access credentials, metadata endpoint, health-check configuration, and immutable image digest. The container runs as UID/GID `10001:10001`; only the scheduler binary receives `cap_net_bind_service` so the established port `80` health contract can remain available without a root process.

The native command is `resource-scheduler`. A `scheduler` executable symlink remains temporarily available for catalog compatibility. The default metadata service name is `metadata`; deployments can set `PASTURESTACK_METADATA_ADDRESS` or pass `--metadata-address` explicitly.

## Build and test

The reviewed build uses Go 1.26.5, Docker CLI 29.6.2, and Docker Buildx 0.34.1. Downloaded tools are checked against fixed SHA-256 values. The Ubuntu base image is digest-pinned, package installation uses the fixed `20260722T164940Z` Ubuntu archive snapshot, and the image exporter normalizes file timestamps to the source commit time.

```bash
make test
make validate
bash scripts/check-build-downloads
bash scripts/check-migration-policy
VERSION_OVERRIDE=v0.8.14 IMAGE_NAMESPACE=pasturestack make package
```

No CI/CD workflow is enabled during the migration and system-integration phase.

The scheduling test suite includes repeated allocation of the same host-port
reservation. Retries for one resource UUID are idempotent, while a different
resource UUID requesting the same host port remains rejected.

## Compatibility and security

Some event names, environment variables, API client type names, dependency namespaces, and scheduling labels are protocol or data contracts. They are isolated and documented in [COMPATIBILITY.md](COMPATIBILITY.md), rather than exposed as PastureStack branding.

Review [SECURITY.md](SECURITY.md) before changing credential handling, metadata trust, custom CA verification, capabilities, or event subscriptions.

## License and attribution

The inherited project remains licensed under [Apache License 2.0](LICENSE). PastureStack does not claim authorship of inherited work. See [ORIGIN.md](ORIGIN.md) and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for provenance and bundled dependency notices.
