# Release Notes

## 0.8.16

- Reconnect the control-plane event subscriber after a clean disconnect or an
  error instead of exiting the container.
- Restart the metadata watcher in place after an error or a recovered panic.
- Keep `/healthcheck` as process liveness and expose dependency checks through
  `/readiness` so a planned control-plane restart does not replace the running
  scheduler container.
- Add regression tests for reconnection, panic recovery, and the separation of
  liveness from readiness.
