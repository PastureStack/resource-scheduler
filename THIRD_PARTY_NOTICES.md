# Third-Party Notices

This program includes vendored Go dependencies declared in `go.mod`, checksum-bound by `go.sum`, and materialized in the standard module-aware `vendor` tree. Exact license or patent texts are copied without modification under `LICENSES/` for source and container distributions.

The dependency set includes Gorilla WebSocket, mapstructure v2, Logrus, urfave/cli v3, Go system packages, gocheck, and the bounded historical control-plane event/API compatibility sources under `third_party`. The obsolete `pkg/errors` and `rancher/log` dependencies are no longer part of the build graph.

The historical `go-rancher-metadata` snapshot did not contain an explicit license file at the pinned revision. Its code is therefore not included in the maintained tree. `internal/metadata` is a separately implemented client for the documented metadata HTTP contract.

The root [LICENSE](LICENSE) governs inherited project code and PastureStack modifications offered under the same terms. It does not replace third-party license texts.
