# Third-Party Notices

This program includes vendored Go dependencies. Their source revisions are recorded in `trash.conf`, and exact license or patent texts are copied without modification under `LICENSES/` for source and container distributions.

The dependency set includes Gorilla WebSocket, mapstructure, pkg/errors, Logrus, urfave/cli, Go system packages, gocheck, and historical control-plane event, API, and logging clients. Their licenses include Apache-2.0, MIT, BSD-family terms, and Go patent grants as identified by the corresponding copied files.

The historical `go-rancher-metadata` snapshot did not contain an explicit license file at the pinned revision. Its code is therefore not included in the maintained tree. `internal/metadata` is a separately implemented client for the documented metadata HTTP contract.

The root [LICENSE](LICENSE) governs inherited project code and PastureStack modifications offered under the same terms. It does not replace third-party license texts.
