# Security

Resource Scheduler receives scoped control-plane credentials, consumes host and container metadata, and decides placement and resource reservations. Treat it as a trusted infrastructure component.

## Deployment requirements

- Pull only the explicit numeric version tag approved by the reviewed PastureStack infrastructure catalog. Published tags are immutable.
- Deploy exactly one scheduler instance in the system environment.
- Supply `CATTLE_URL`, `CATTLE_ACCESS_KEY`, and `CATTLE_SECRET_KEY` through the control plane; never store them in images, catalog files, logs, or repository content.
- Use the link-local or service-discovery metadata endpoint provided by the control plane.
- Keep the container non-root. Only `/usr/bin/resource-scheduler` may retain `cap_net_bind_service` for the established health-check port.
- When `CATTLE_CA_CHECKSUM` is set, the downloaded CA must match that SHA-256 value before it is trusted.
- Treat debug logs as operationally sensitive.
- Keep liveness independent of transient external dependency failures. Use
  `/readiness` when dependency availability must be checked without replacing
  a healthy scheduler process.

## Build requirements

- Build from the public source commit named by `org.opencontainers.image.revision`.
- Verify Go and Docker CLI downloads and the Buildx source archive before extraction or execution.
- Apply the checksum-recorded Buildx patch and reject the build if the resulting binary still records the legacy Docker module.
- Keep the runtime base image digest-pinned. Resolve exact direct-package versions from the dated Canonical snapshot in `ubuntu-apt.lock`, and retain the complete resolved `dpkg` inventory in each build and runtime image.
- Verify that `go.mod`, `go.sum`, and `vendor/modules.txt` agree by compiling and testing with `-mod=vendor`.
- Run unit tests, race tests, `go vet`, formatting checks, build-policy checks, migration-policy checks, secret scanning, an SBOM inventory, and all-severity vulnerability scanning before publishing.
- Block every runtime High/Critical finding and every finding with a vendor-published fixed version. Preserve vendor-unfixed Low/Medium findings in the release scan, risk register, and OpenVEX as `under_investigation`; never label them fixed or not affected.
- A builder High or Critical finding may be classified as not affected only when its CVE, package version, architecture, official Canonical archive SHA-256, and complete installed owned-file identity match the committed per-CVE review in `toolchain/linux-libc-dev-review.json`. CI must prove that the Linux CNA kernel implementation and private-header paths are not shipped, the package contains no ELF/kernel payload, and the runtime does not contain the package. Emit exact-package OpenVEX with `vulnerable_code_not_present`, while preserving the raw scan. No date-based exception is supported. Unknown CVEs, package or content drift, any vendor-fixed builder finding, and unrelated High/Critical findings remain blocking. This classification makes no statement about the host kernel.
- Publish a new immutable version when source or dependencies change; do not replace an existing release digest.

Report vulnerabilities privately to the PastureStack organization maintainers. Do not include credentials, internal addresses, customer data, or exploit details in public issues.
