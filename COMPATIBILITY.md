# Compatibility Contract

Resource Scheduler is branded, packaged, and deployed as a PastureStack component. A small set of inherited literals must remain while the historical control-plane protocol is supported.

## Retained protocol and data identifiers

- Metadata API date: `/2016-07-29`
- Event names: `scheduler.prioritize`, `scheduler.reserve`, and `scheduler.release`
- Event resource type: `schedulerRequest`
- Scheduling labels under `io.rancher.scheduler.*`
- Control-plane environment variables under `CATTLE_*`
- Debug alias: `RANCHER_DEBUG`
- Vendored dependency import paths and generated API type names under `github.com/rancher/*`
- Compatibility command: `/usr/bin/scheduler`

These strings are compatibility identifiers, not product names, image names, public service names, or claims of affiliation. Removing or renaming one without a coordinated Server, Agent, Catalog, metadata, and upgrade test can silently break placement or resource accounting.

## PastureStack-native interfaces

- Executable: `/usr/bin/resource-scheduler`
- Image: `ghcr.io/pasturestack/resource-scheduler`
- Source repository: `https://github.com/PastureStack/resource-scheduler`
- Debug variable: `PASTURESTACK_DEBUG`
- Metadata address variable: `PASTURESTACK_METADATA_ADDRESS`
- Default metadata service name: `metadata`
- Repeated placement checks may present the same resource UUID after a
  successful host-port reservation; that retry must remain idempotent.
- `/healthcheck` is the compatibility liveness endpoint used by the system
  service. `/readiness` reports metadata and control-plane dependency state.
- The event subscriber and metadata watcher reconnect in place after a
  temporary control-plane interruption so the existing container identity is
  retained.

Production catalog templates must use the PastureStack image name and an explicit numeric version tag. The compatibility command and variables remain accepted only so the running control plane can be upgraded without an all-at-once protocol change.
