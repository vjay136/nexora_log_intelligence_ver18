# NLI V18 capability and acceptance matrix

## Status basis

This matrix records the Ver18 source baseline inspected on 2026-10-21. The
baseline was derived from Ver17 commit
`f896cdebc04a5790330a38e0be1dd5f4bea23c44`. Statuses below use source, test
and existing evidence-file inspection. Local Ver18 verification subsequently
passed the complete Go package suite, 41 WebApp browser cases, and 20
Vercel/Cloudflare/Vite contract tests; packaging and immutable publication
remain release gates until their exact tagged artifacts pass.

Status meanings:

- **Implemented** — operational code and focused repository evidence exist in the current baseline.
- **Partial** — a bounded subset exists; the stated remainder is not claimed.
- **Planned** — no complete operational implementation was found.
- **Release pending** — implementation exists, but the Ver18 packaging/publication gate remains incomplete.

Existing evidence anchors include `V18_PROGRESS.md`, `V18_ENGINEERING_AUDIT.md`, `V18_MASTER_SPEC_IMPLEMENTATION_MATRIX.md`, `V18_NURSF_ARCHITECTURE_AND_VERIFICATION.md`, `V18_STREAMING_AND_PERFORMANCE_SLO.md`, source tests under `internal/` and browser/packaging tests under `web/` and `tests/`. A release report must replace evidence descriptions with actual current-run results.

## Capability matrix

| Subsystem | Capability | Baseline status | Current evidence or boundary | Acceptance gate |
|---|---|---|---|---|
| NHLE | Concurrent multi-source readers | Implemented | Fixed bounded pool, up to 256 configured sources, worker caps and per-source batches in agent source-worker code/tests. | Append to 100+ sources concurrently; prove every accepted line accounted, per-source order, bounded memory and no starvation. |
| NHLE | Linux event wakeup and reconciliation | Implemented | Inotify watcher with coalesced wakeups and periodic reconciliation fallback. | Test append/burst/watcher failure/overflow; reconciliation resumes without full-file rereads or loss. |
| NHLE | Checkpoints and restart recovery | Implemented | Source checkpointing and pending multiline flush behavior have focused tests. | Crash/restart at read, parse and output barriers; verify documented replay/duplicate behavior and byte offsets. |
| NHLE | Rotation, truncation and partial records | Implemented | Existing tailing and source-worker tests cover supported local-file behavior. | Test rename, replacement, copy-truncate, partial LF/CRLF and restart; declare network-filesystem limits. |
| NHLE | Bounded physical/multiline input | Implemented | 1 MiB physical segments; 8 MiB/10,000-line multiline ceilings; invalid input uses visible unknown route. | Adversarial oversized/invalid-encoding tests prove no crash, silent truncation or false semantic success. |
| NHLE | Fair scheduling/backpressure | Implemented | Bounded source batches, queues, worker limits and backpressure telemetry exist. | Saturate one source/parser/output; unrelated sources must progress and queue limits remain enforced. |
| NHLE | General adaptive worker tuning | Partial | Static/automatic bounded worker selection exists; broad runtime workload adaptation is not evidenced. | Compare bounded adaptive policy with baseline under burst, CPU/memory pressure and mixed-cost parsers; preserve order/correctness. |
| NAPI | Unknown/ambiguous structural evidence | Implemented | Engine emits bounded fingerprints, delimiters, keys and ranked candidate metadata; fail-closed ambiguity path exists. | Golden tests for known, variant, unknown, corrupted and ambiguous samples; no weak forced classification. |
| NAPI | Multi-sample JSON/KV/regex synthesis | Implemented | Authenticated off-hot-path synthesis uses discovery/held-out/negative samples and bounded RE2 checks. | Candidate exact-field, false-positive, safety and reproducibility gates pass; candidate remains a draft. |
| NAPI | Multi-strategy general grammar induction | Partial | JSON, KV, anchored-regex and registered primitives exist; arbitrary grammar induction is not complete. | Add quoted/escaped delimiter, optional/reordered, nested and multiline grammar corpus with held-out correctness evidence. |
| NAPI | Confidence calibration | Partial | Confidence metadata exists; broad empirical calibration against labeled correctness buckets is not evidenced. | Publish reliability curve/Brier or equivalent calibration measure on versioned labeled datasets. |
| NAPI | Drift-triggered learning | Partial | Drift and parser-state views exist; automatic evidence-driven candidate lifecycle is incomplete. | Inject schema/template/distribution drift; create evidence and candidate without changing active production parser. |
| Parser compiler | Declarative parser safety | Implemented | Parser packs are declarative; controller does not execute imported Go/Alloy/VRL/shell/model code; RE2 bounds exist. | Malformed, pathological and oversized definitions fail closed with useful diagnostics and no production impact. |
| Parser compiler | Typed universal IR and optimized plans | Planned | No complete typed IR/compiler with reference equivalence, CSE and plan artifacts is claimed. | Versioned schema, static checks, resource budgets, reproducible artifact and differential equivalence suite. |
| Parser compiler | Version-aware bounded caches | Partial | Registry/cache behavior exists in current runtime; complete compiler-plan cache metrics/invalidation are not evidenced. | Version activation/rollback invalidates exactly affected entries; memory cap and hit/miss telemetry verified. |
| Schema/semantics | Known-key, layout, projection and normalization catalogs | Implemented | Catalogs and reviewed parser packs are exposed, versioned and exercised by corpus tests. | Supplied reviewed corpus meets golden field/type expectations and field accounting. |
| Schema/semantics | Canonical provenance-bearing event | Partial | Normalized envelopes, parser metadata and source accounting exist; universal field-level derivation lineage is incomplete. | Every output field reports copied/normalized/derived/enriched origin and parser/config version in an explain trace. |
| Schema/semantics | Target Format Studio and mapping graph | Partial | Output projection/configuration exists; complete schema-first interactive graph and loss simulator are not evidenced. | Preview exact mapping/types/nulls/loss before activation; version and validate mappings against each implemented sink. |
| Schema/semantics | Additional Avro/Protobuf/CEF adapters | Planned | Core supported outputs are JSONL, Loki and ClickHouse. | Implement serializer/schema registry contract, round-trip tests, sink acceptance and documentation per adapter. |
| NPA | Draft, validation, independent review, publication and rollout | Implemented | Assurance and parser fleet-sync paths separate draft, validation, review, publish and client acknowledgement. | A requester cannot self-approve where policy forbids it; changed hash invalidates approval; client reports applied hash. |
| NPA | Full durable lifecycle state machine | Partial | Several lifecycle stages exist; observe-through-drift durable orchestration is not complete. | Persist every transition/job/evidence/error; restart and idempotent retry cannot duplicate actions. |
| NPA | Shadow execution and traffic canary | Planned | No complete traffic-scoped shadow/canary orchestrator is claimed. | Compare candidate/active fields and cost on configured scope; enforce halt thresholds and audited rollback. |
| NPA | Emergency freeze and policy-as-code | Partial | Approval/RBAC controls exist; a complete unified freeze/promotion policy engine is not evidenced. | Freeze blocks new promotions without interrupting ingestion; policies are versioned, validated and audited. |
| Durability | Independent bounded destination queues | Implemented | Output router has per-destination workers, barriers and failure propagation. | One blocked sink cannot stall healthy sinks outside declared durability policy; bounds and errors remain visible. |
| Durability | Retry, fallback/spool and quarantine | Implemented | Output transports, drain/quarantine and client-control APIs/tests exist. | Exhaust retries under outage; records remain searchable/replayable and service restart resumes documented state. |
| Durability | Cross-destination exactly once | Planned / not claimed | V18 correctly does not claim a distributed transaction across file, Loki and ClickHouse. | If pursued, define sink idempotency keys and transaction boundaries; otherwise retain at-least-once disclosure. |
| Replay | Search, diagnosis and bounded reprocessing | Implemented | Quarantine APIs, history, diagnostics, raw access controls and replay flows exist. | Failure→parser correction→authorized replay→per-destination outcome completes with original record reference and audit. |
| Replay | Parse-at-query and selective historical backfill | Planned | General raw-first/query-time parsing and cost preview are not evidenced. | Durable bounded jobs, range/source filters, preview, budgets, parser-version comparison and duplicate-safe output. |
| Outputs | Local JSONL | Implemented | Hourly family output, local durability and output tests exist. | Validate JSON, permissions, partial write, disk-full, rotation, flush barrier and restart. |
| Outputs | Loki | Implemented with limits | Transport supports configured endpoint/batch/retry paths; live endpoint performance is environment-specific. | Test auth/TLS/rate-limit/transient/permanent responses and remote acknowledgement with a controlled endpoint. |
| Outputs | ClickHouse | Implemented with limits | Typed HTTP delivery/config paths exist; target schema is operator-specific. | Verify real table types, batch insert, auth/TLS, schema rejection, retry and acknowledgement. |
| NURSF | Canonical telemetry snapshot and incremental stream | Implemented | `nli.sync.v1`, client batches, controller canonical store, snapshot/diagnostics/SSE and shared WebApp store exist. | Concurrent-client tests cover snapshot, delta, sequence, generation, duplicate, out-of-order, gap and reconnect. |
| NURSF | One-second heartbeat and background UI refresh | Implemented | In-memory heartbeat and semantic render coalescing preserve page state. | Browser test proves no whole-page replacement, focus/input/scroll/subtab loss, or duplicate history on heartbeat. |
| NURSF | Explicit freshness/missing semantics | Implemented | Live/delayed/stale/disconnected/unavailable/not-instrumented states and actionable issues are represented. | Missing measurement never becomes green zero; disconnect retains last value and timestamp while status becomes stale. |
| NURSF | 1,000-node measured capacity | Planned | Architecture is bounded, but 1,000-node production capacity is not proven. | Record p50/p95/p99 propagation, missed deadlines, controller CPU/memory/network and subscriber behavior at declared scale. |
| Monitoring | Eleven canonical dashboards | Implemented | WebApp exposes Overview, Fleet, Pipeline, Sources, Parsers, Destinations, Queues, Infrastructure, Sync, Incidents and Release views. | All KPIs reconcile with one snapshot; filters/drilldowns persist; unavailable data is explicit; no per-panel fleet polling. |
| Monitoring | Professional chart interactions | Implemented with limits | Shared filters, legends, exact hover values and freshness exist; usability remains subject to browser acceptance. | Desktop/browser accessibility and high-frequency rendering tests pass at representative series volume. |
| Monitoring | Confirmed automated root cause | Partial | Structured diagnostic reasons/actions exist; some causes are inferred. | Label hypotheses; mark root cause confirmed only when evidence identifies the failed stage/dependency. |
| NAL | Analysis, corpus, drafts, review, history, fixtures and helper | Implemented | Assurance WebApp/API provides these workflows with RBAC and audit. | Browser→API→persistence flow passes; raw/sensitive access obeys authorization and retention. |
| NAL | Versioned plans and durable bounded runs | Implemented | Plan revisions, two-worker durable queue, restart requeue, cancel/finalize and evidence metadata exist. | Exact plan hash binds run; interrupted run is requeued/failed, never falsely passed; concurrent claims are safe for one controller. |
| NAL | Executable local scenarios | Implemented | Parser regression, failure/recovery, replay, destination JSON contract, bounded performance and privacy/integrity scenarios exist. | Each scenario produces a durable truthful result and evidence without changing production. |
| NAL | Distributed/external lab orchestration | Planned | Multi-controller workers, external sink tests, resumable dependency graphs and customer environments are not complete. | Tenant-isolated environments, leases, stage resume, cancellation, external endpoint safety and evidence integrity tests. |
| NAL | Evidence-backed promotion | Partial | Validation/approval/publish/rollout exists; fully automated preproduction and canary orchestration does not. | Exact artifact/evidence/checksum and separation of duties gate promotion; canary and rollback are verified. |
| AI | Optional provider configuration and bounded proposals | Implemented | Off/manual/automatic-proposal modes, encrypted server credentials, bounded/redacted samples and audited proposals exist. | Provider failure never blocks ingestion; AI cannot approve/publish; secret and payload leakage tests pass. |
| AI | Offline/local model mode | Partial | AI can be disabled; a supported packaged local inference runtime is not evidenced. | Define provider interface, resource isolation, model provenance and deterministic post-validation before claiming local inference. |
| Security | Authentication, RBAC, TLS, audit and secret isolation | Implemented | Controller middleware, enrollment/TLS, audit log and server-only secret conventions have tests/docs. | Negative authorization, TLS failure, secret scanning, log redaction and audit-integrity gates pass. |
| Security | Tenant isolation | Partial | Role and metadata boundaries exist; full multi-tenant isolation across all samples/caches/jobs is not proven. | Cross-tenant negative tests cover APIs, storage, stream subscriptions, caches, exports, replay and AI. |
| Packaging | Rocky/RHEL-family x86_64 RPM/systemd | Release pending | Specs/installers/lifecycle tests target Rocky 9/10; V18 acceptance checklist packaging gates remain open. | Build reproducibly; fresh install, V15→V18 upgrade, downgrade/re-upgrade, two-node enrollment, uninstall and service checks pass. |
| Hosted WebApp | Vercel/Cloudflare proxy packages | Release pending | Same-origin proxy contracts and Vite export exist; target deployment requires public HTTPS controller origin. | Production build plus authenticated browser→proxy→controller login/API/SSE passes without exposing server secrets. |
| Provenance | Checksums, source revision, SBOM and provenance | Release pending | Release scripts generate these assets; exact Ver18 release artifacts are not yet published. | Two clean builds match where defined; manifest/SBOM/provenance/checksums bind the exact tagged artifacts. |

## Approved WebApp information architecture

This layout consolidates overlapping proposals and prevents a 25-tab Parsing Intelligence section from duplicating NPA, NAL and Monitoring.

| Top-level workspace | Subtabs | Data authority |
|---|---|---|
| Command Center | Overview; Incidents; Capacity | Canonical controller snapshot and incident store |
| Monitoring | Overview; Fleet & Nodes; Pipeline; Sources; Parser Performance; Destinations; Queues, Replay & Quarantine; Infrastructure; Latency & Synchronization; Errors & Incidents; Assurance & Release Health | NURSF shared state plus bounded historical queries |
| Ingestion & Execution (NHLE) | Sources; Framing & Checkpoints; Workers & Backpressure; Performance Profiles | Client source/runtime telemetry |
| Parsing Intelligence (NAPI) | Discovery; Grammar & Structure; Candidate Brain; Schema & Semantic Studio; Compiler; Explainability; Drift; Registry & Lineage | Parser registry, candidate/evidence store and canonical telemetry |
| Parsing Autopilot (NPA) | Lifecycle; Validation & Differential; Shadow & Canary; Deployments; Historical Replay; Policies & Freeze | Durable parser lifecycle and rollout state |
| Delivery & Recovery | Outputs; Quarantine; Replay; Durability | Per-destination delivery and recovery stores |
| Real-Time Sync (NURSF) | Fabric Overview; Nodes & Sequences; Freshness & Latency; Contract Coverage; Diagnostics | Canonical telemetry store |
| Assurance Lab (NAL) | Overview; Feature Explorer; Test Workbench; Parser Assurance; Performance & Reliability; Integration; Security; Evidence Vault; Release & Promotion; Customer Demonstration; Readiness Reports | Durable Assurance records and immutable evidence |
| Administration | Identity & RBAC; Data Policy; AI Providers; Configuration; Audit; Licensing & Provenance | Authenticated controller configuration/audit services |

## Truthful product claims

Use these formulations in release material:

- **Known formats:** correctness is scoped to reviewed, versioned fixtures and golden expectations.
- **Unknown formats:** accepted input is preserved or reversibly represented with explicit diagnostics; semantic correctness is not invented.
- **Throughput:** publish measured results with hardware, workload, parser/output mix and correctness accounting.
- **Freshness:** one second is a target; publish actual p50/p95/p99 and missed-deadline counts.
- **Delivery:** state at-least-once and sink acknowledgement semantics; do not claim universal exactly once.
- **AI:** assistance proposes bounded declarative changes; deterministic validation and human/policy approval control production.
- **Scale:** 1,000 nodes and 10 million records/s remain qualification goals until reproduced.

## Release evidence checklist

- [ ] Clean committed source revision recorded.
- [ ] Uncached Go unit/integration suite passed.
- [ ] Browser UI and proxy contract suites passed.
- [ ] Corpus golden/parity/unknown accounting passed.
- [ ] Parser synthesis safety and held-out/negative tests passed.
- [ ] NURSF reset/duplicate/gap/reconnect/staleness/no-shake tests passed.
- [ ] Queue saturation, sink outage, restart and replay recovery tests passed.
- [ ] Performance result records host, dataset checksum, configuration, outputs and correctness.
- [ ] RPM contents, permissions, fresh install and service health passed on each declared platform.
- [ ] Genuine supported-version upgrade, rollback/re-upgrade and configuration/state preservation passed.
- [ ] Two-node enrollment, TLS and parser/config synchronization passed.
- [ ] Vercel/Cloudflare production build and authenticated proxy/SSE contract passed.
- [ ] Secret scan and source/parser/dependency license/provenance review passed.
- [ ] SHA256SUMS, source revision, manifest, SBOM and provenance match released bytes.
- [ ] All partial, planned, blocked and environment-dependent capabilities are disclosed.

