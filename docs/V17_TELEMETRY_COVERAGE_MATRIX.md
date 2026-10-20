# V17 canonical telemetry coverage matrix

| Area | Authoritative owner | V17 live state | Publication/subscribers | Freshness | Validation |
|---|---|---|---|---|---|
| Node identity and connectivity | controller fleet registry | node ID, generation, sequence, last seen, reconnect state | `nli.sync.v1` → NURSF store → Fleet/Monitoring | 1 s target; stale/disconnected thresholds | registration, duplicate, gap, reconnect tests |
| Source ingestion | client source workers | records/bytes, rates, phase, source errors | compact client batch → NHLE/Source views | per-source timestamp and reason | multi-source and canonical aggregation tests |
| Parsing/normalization | client parser workers | successes, failures, normalized/quarantine totals and rates | same batch → Command Center/NHLE/Parser views | same canonical revision | corpus, parity, and sync contract tests |
| Queues/backpressure | client bounded queues | depth, capacity, backpressure, diagnostic issue | immediate critical delta plus periodic batch | critical transition immediate | saturation/coalescing tests |
| Destination delivery | output adapters | metadata already present where adapter reports it | controller telemetry/history → Destination view | explicit last success; missing stays unavailable | adapter and delivery tests |
| Replay/quarantine | controller rollout/quarantine state | job and record lifecycle from existing APIs | controller state → Monitoring/NAL/Quarantine | state transition driven | replay and quarantine API tests |
| NHLE | NURSF canonical aggregate | source/node counts, input/output/quarantine, queue, rates | snapshot + SSE → NHLE Live | controller classified | shared-store UI tests |
| NAPI | controller NAPI/AI state | requests, failures, pending drafts, last proposal | canonical subsystem state + safe events → NAPI Live | controller classified | synthesis and UI tests |
| NAL | assurance scheduler/store | total, queued, running, passed, failed, cancelled | canonical subsystem state + run events → NAL Live | controller classified | durable scheduler and UI tests |
| Synchronization | NURSF controller | accepted/duplicate/out-of-order/gap/reconnect/coalesced/rejected, p50/p95/p99, subscribers | `/sync/snapshot`, `/sync/diagnostics`, `/sync/stream` | measured and displayed | concurrency and 1,000-node bounded-state tests |
| System resources | existing heartbeat/system APIs | only metrics actually reported by the installed node | canonical snapshot or existing bounded API | missing is unavailable | API/UI empty-state tests |
| Errors/incidents | NURSF issue index and existing events | severity, stage, component, reason, evidence, first/last seen, occurrences, action | immediate critical or incremental delta | transition driven | critical-delivery and deduplication tests |
| Browser state | one WebApp telemetry store | one schema/revision and bounded history | initial snapshot + ordered SSE deltas | browser apply delay shown | reconnect/schema/gap/UI consistency tests |

## Current instrumentation limits

Host disk I/O, network throughput, file descriptor use, destination-specific
acknowledgement latency, and some pipeline-stage latency fields depend on the
installed platform and adapter. When no real measurement is present, V17 shows
`Not instrumented` or `Unavailable`. It does not synthesize those measurements.

Historical charts continue to use bounded, aggregated controller queries. The
live stream is not treated as an unlimited historical database.
