# Account-service review handoff

`nextendo-account-uch.patch` is a reviewable patch against `NextendoNetwork/nextendo-account` source commit `bd8d4a3d8acb44197af1da6664a8e09c9f921127`. It is supplied for its maintainer to review/apply; no production account code or service was modified.

The diff only appends the optional `DASH_UCH_URL` presence source and adds tests for preserving existing game sources and detecting a live UCH peer. It excludes the laboratory listener override and the separate Classics proposal. The source-linked tests passed locally, and `git apply --check` passed against the recorded base file.

After review, from a checkout of the account repository:

```sh
git apply --check /path/to/nextendo-account-uch.patch
git apply /path/to/nextendo-account-uch.patch
go test ./...
```

The maintainer privately configures `DASH_UCH_URL` to the UCH presence listener's internal base URL. `/api/stats` is appended by the existing account code. The existing private `DASH_TOKEN` must match UCH's private `UCH_STATS_KEY`; these values are not included in the patch. Leaving `DASH_UCH_URL` unset preserves the existing source list.

UCH currently restricts `-stats-addr` to literal loopback. A co-located deployment can use `http://127.0.0.1:8089`; separate container network namespaces cannot share this loopback address without an explicitly reviewed deployment arrangement. Do not expose the presence listener publicly or put its keyed URL into logs. Review that topology before rollout.

The existing account service's presence polling is fail-open on network errors; this patch does not change that upstream policy. A configured but unreachable UCH listener therefore cannot establish global one-location exclusion. Confirm the live presence integration and exclusion/release behavior during acceptance.
