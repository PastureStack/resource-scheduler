# Release Notes

## 0.8.19 candidate

- Rebuild with the official Go 1.27.2 security release and checksum-verified archives.
- Refresh the dated Canonical snapshot to `20261020T000000Z`, including the
  official curl, OpenSSL, Perl, Git, and Linux API header package updates.
- Preserve the runtime security gate and replace the expired builder-header
  exception with exact-CVE, package, archive, and installed-content applicability
  evidence. Unknown or changed High/Critical findings remain blocking; the host
  kernel is outside that package-only assessment.
- Verify the packaged scheduler's non-root port-80 health endpoint and the
  liveness/readiness behavior with isolated CI fixtures.

## 0.8.18

- Update every runtime package available from the fixed Ubuntu snapshot.
- Preserve the full all-severity runtime scan and explicitly register Low/Medium
  findings for which Ubuntu has not published a fixed package; keep them marked
  `under_investigation` instead of weakening the scan or claiming resolution.
- Block release whenever a fixed package is available or a High/Critical
  runtime finding remains.

## 0.8.17

- Update the checksum-pinned Go, Docker CLI, Buildx, gRPC, and Ubuntu package
  inputs used by the maintained build and runtime images.
- Produce reproducible runtime images and complete source, builder, and runtime
  dependency evidence from the reviewed release commit.
- Enforce pure numeric SemVer release tags and publish the image, SBOMs,
  checksums, and build provenance from the same annotated tag.

## 0.8.16

- Reconnect the control-plane event subscriber after a clean disconnect or an
  error instead of exiting the container.
- Restart the metadata watcher in place after an error or a recovered panic.
- Keep `/healthcheck` as process liveness and expose dependency checks through
  `/readiness` so a planned control-plane restart does not replace the running
  scheduler container.
- Add regression tests for reconnection, panic recovery, and the separation of
  liveness from readiness.
