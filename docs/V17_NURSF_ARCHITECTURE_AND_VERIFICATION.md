# Nexora Log Intelligence V17 — NURSF architecture and verification

## Scope

V17 adds the Nexora Unified Real-Time Synchronization Fabric (NURSF) to the
existing client, controller, and WebApp. NURSF is the authoritative live path
for operational telemetry. It does not carry customer log bodies and cannot
approve parsers, change production configuration, start replay, or promote a
release.

The default target is one compact publication per client per second. This is a
target, not a latency guarantee. The controller records the actual timing,
marks delayed or stale data, and exposes why the target was missed when the
available evidence identifies a stage.

## Data flow

```text
input readers / parser / output workers
        │ inexpensive local counters and gauges
        ▼
bounded client collector (default 1 second)
        │ authenticated nli.sync.v1 batch
        ▼
controller ordering + validation
        │ generation, sequence, duplicate and gap handling
        ▼
canonical in-memory fleet state
        ├── GET /api/v1/sync/snapshot
        ├── GET /api/v1/sync/diagnostics
        └── GET /api/v1/sync/stream (SSE snapshot + ordered deltas)
                    │
                    ▼
one bounded browser telemetry store
        ├── Command Center
        ├── Monitoring workspace
        ├── NHLE Live
        ├── NAPI Live
        ├── NAL Live
        └── Streams and synchronization diagnostics
```

An additional browser does not create another fleet collection cycle. Each
browser receives the current controller snapshot and subsequent incremental
revisions. A schema mismatch, sequence gap, or invalid delta triggers an
authoritative snapshot resynchronization instead of local counter invention.

## Canonical contract

The transport schema is `nli.sync.v1`. A client batch contains a trusted node
identity, process generation, monotonic sequence, collection and publication
timestamps, configured intervals, a full-snapshot marker, and a bounded list
of source measurements. The controller rejects malformed or oversized input,
deduplicates equal sequences, rejects older sequences, detects gaps, and treats
a generation change as a reconnect/reset boundary.

Canonical counters include source records and bytes read, parser successes and
failures, normalized and quarantined records. Canonical gauges include queue
depth, backpressure sources, live rates, and freshness. Missing values remain
unavailable; they are never converted to healthy zero.

The shared controller snapshot also contains:

- NURSF diagnostics and latency percentiles;
- NHLE ingestion, parsing, queue, and backpressure state;
- NAPI synthesis request, failure, and pending-draft state;
- NAL queued, running, passed, failed, and cancelled run state;
- bounded actionable issues with stage, component, node/source, reason,
  evidence, severity, correlation ID, occurrence count, and suggested action.

## Freshness and latency

The default collection and publication target is 1,000 ms. Controller state
uses explicit `live`, `delayed`, `stale`, `disconnected`, `unavailable`, and
`not_instrumented` semantics. The last known measurement is retained when a
node stops publishing, but its freshness changes and the last update remains
visible.

Controller processing durations use a local monotonic clock. Cross-host
propagation is reported only as an observed wall-clock estimate. If clock skew
makes that estimate untrustworthy, the value is unavailable and a clock-skew
diagnostic is recorded. The UI additionally measures event-to-store browser
application time.

Diagnostics separate collection delay, client queue/backpressure, transport
or reconnect failure, controller validation/processing, subscriber coalescing,
schema mismatch, authentication/TLS rejection, and browser application delay.
Suspected bottlenecks are labelled as likely causes unless evidence proves the
cause.

## Bounded operation and security

- Client publications are batched; there is no request per parsed record.
- Source state, issue history, delay samples, subscribers, and browser history
  have fixed limits.
- Slow subscribers receive the newest canonical revision and a coalesced count;
  an unbounded subscriber queue is never created.
- Critical state transitions bypass the ordinary visualization cadence.
- Node authentication and controller RBAC protect ingestion and subscription.
- Raw log content, email addresses, session identifiers, and arbitrary event
  values are excluded from this telemetry contract.
- Telemetry failure does not block parsing or destination delivery.

## Monitoring information architecture

The Monitoring workspace uses one filter context and one shared store for its
Overview, Fleet & Nodes, Pipeline, Sources, Parser Performance, Destinations,
Queues/Replay/Quarantine, Infrastructure, Latency & Synchronization,
Errors & Incidents, and Assurance & Release Health dashboards. Panels show
units, legends, exact hover values, the applicable update time, and freshness.
Metrics that the installed node does not publish are labelled `Unavailable` or
`Not instrumented` rather than populated with demo data.

## Verification commands

```bash
go test -count=1 ./...
go vet ./...
npm run web:check
make release-static
make release-package-test
make release-lifecycle-test
make release-build OUTPUT_DIR="$NLI_BUILD_SCRATCH/nli-v17-release"
make release-rocky \
  OUTPUT_DIR="$NLI_BUILD_SCRATCH/nli-v17-release" \
  PREVIOUS_DIR=/path/to/genuine/v16-rpms
```

The release report must record actual pass/fail results and measured benchmark
values. Repository design alone does not prove one-second delivery for every
network or 1,000-node production deployment.

## Rollback

Stop V17 services, install the verified V16 controller and client RPMs from the
previous release, restore the version-matched configuration backup if schema
changes were applied, then restart services. Do not reuse a V17 controller
state backup with older binaries unless the release notes explicitly declare
it compatible. Existing parser approval, authentication, and artifact
verification rules remain in force during rollback.

