# Security

Resource Scheduler receives scoped control-plane credentials, consumes host and container metadata, and decides placement and resource reservations. Treat it as a trusted infrastructure component.

## Deployment requirements

- Pull only the immutable digest pinned by the reviewed PastureStack infrastructure catalog.
- Deploy exactly one scheduler instance in the system environment.
- Supply `CATTLE_URL`, `CATTLE_ACCESS_KEY`, and `CATTLE_SECRET_KEY` through the control plane; never store them in images, catalog files, logs, or repository content.
- Use the link-local or service-discovery metadata endpoint provided by the control plane.
- Keep the container non-root. Only `/usr/bin/resource-scheduler` may retain `cap_net_bind_service` for the established health-check port.
- When `CATTLE_CA_CHECKSUM` is set, the downloaded CA must match that SHA-256 value before it is trusted.
- Treat debug logs as operationally sensitive.

## Build requirements

- Build from the public source commit named by `org.opencontainers.image.revision`.
- Verify Go, Docker CLI, and Buildx downloads before extraction or execution.
- Keep the runtime base image digest-pinned and install packages from a fixed Ubuntu archive snapshot.
- Run unit tests, race tests, `go vet`, formatting checks, build-policy checks, migration-policy checks, secret scanning, an SBOM inventory, and High/Critical vulnerability scanning before publishing.
- Publish a new immutable version when source or dependencies change; do not replace an existing release digest.

Report vulnerabilities privately to the PastureStack organization maintainers. Do not include credentials, internal addresses, customer data, or exploit details in public issues.
