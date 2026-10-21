# Nexora Log Intelligence V18 master engineering prompt

## Purpose

This document is the consolidated engineering contract for NLI V18. It merges the supplied NHLE, NAPI, NPA, NURSF, Monitoring, NAL, and extra parsing-intelligence proposals into one implementation program. The source prompts are design input; they are not evidence that a capability exists.

V18 must improve the existing product through measured, reviewable changes. It must preserve working behavior and must not represent a screen, API stub, document, or generated JSON as proof of an operational feature.

## Product objective

Build a coherent enterprise log-processing platform in which:

```text
SOURCE
  -> READ / FRAME / CHECKPOINT
  -> CLASSIFY / PARSE / NORMALIZE / VALIDATE / ENRICH
  -> ROUTE / PERSIST / DELIVER / ACKNOWLEDGE
  -> TELEMETRY / OPERATIONS / ASSURANCE
```

operates as one bounded, observable and recoverable system.

The product should parse approved known formats deterministically, preserve unfamiliar or ambiguous records, help operators develop better parsers, and promote parser changes only through evidence-backed policy gates. “Any log,” “hypersonic,” and “one-second live” are engineering goals whose achieved scope must be stated with workload and environment evidence. They are not universal guarantees.

## Non-negotiable rules

1. Inspect the repository, baseline tests, active configuration, persisted data models, packaging, and release workflow before changing behavior.
2. Reuse existing components and extend them compatibly. Do not create a second registry, telemetry pipeline, lifecycle controller, or dashboard data source for the same responsibility.
3. Never silently lose, corrupt, skip, or falsely acknowledge an accepted record.
4. Never report missing telemetry as zero or healthy. Use explicit live, delayed, stale, disconnected, unavailable, not-instrumented, and collection-failed states.
5. Bound record size, multiline state, recursion, parser candidates, queues, caches, retries, concurrency, browser history, retained samples, and expensive jobs.
6. Treat records, parser definitions, regular expressions, imports, archives, enrichments, and model output as untrusted input.
7. Production parsing must be deterministic, reproducible, testable, and independent of a per-record external AI request.
8. Parser proposals cannot approve or activate themselves. Approval is bound to the exact version, configuration, evidence and artifact checksum.
9. Preserve authentication, authorization, tenant boundaries, TLS verification, secret isolation, audit history, package ownership, service permissions, and migration safety.
10. Do not claim exactly-once delivery across unrelated systems. Define acceptance, buffering, sending, remote acknowledgement and durable commitment separately.
11. Do not add native code, SIMD, GPU processing, a message broker, database, paid API, or commercial dependency without profiling evidence and a documented operational and licensing review.
12. Every completion claim must name the test, measurement, artifact, or observable end-to-end path that supports it.

## Baseline discipline

Classify every requested capability as one of:

- **Implemented and verified** — operational code is connected end to end and current evidence passed.
- **Implemented; release verification pending** — operational code and tests exist, but the current release gate has not completed.
- **Partial** — a useful bounded subset exists and its boundary is documented.
- **Planned** — no complete operational implementation is evidenced.
- **Blocked** — an external prerequisite or architectural decision is documented.
- **Not applicable** — the capability does not apply to the supported product surface.

Update `V18_CAPABILITY_AND_ACCEPTANCE_MATRIX.md` as implementation evidence changes. A prompt or earlier-version report does not upgrade a status.

## Unified subsystem requirements

### 1. Nexora Hypersonic Log Engine (NHLE)

NHLE owns ingestion, record framing, ordered source progress, bounded concurrent processing, checkpoints, backpressure and dispatch to independent outputs.

Required behavior:

- Support all configured sources rather than assuming a fixed filename list.
- Use event-driven wakeups where supported and periodic reconciliation as the correctness backstop.
- Handle initial discovery, incremental reads, LF/CRLF, partial lines, bounded multiline records, rename rotation, replacement, truncation, copy-truncate and watcher failure.
- Maintain stable source/file identity, persisted checkpoints and restart-safe incomplete-record behavior.
- Schedule sources fairly through bounded workers and per-source work budgets. One busy file, expensive parser, or slow output must not starve unrelated work.
- Expose source state, file position, last activity, EOF wait/resume, throughput, queue pressure, parser outcome, error and recovery state.
- Shut down and reload configuration without acknowledging work that has not met the configured durability contract.

Acceptance requires fresh tests for append, partial line, multiline limit, rotation, truncation, watcher recovery, restart, source fairness, saturated queues, destination slowdown and clean shutdown. Performance reports must include bytes, record sizes, parser mix, outputs, hardware, worker count, correctness accounting, CPU and memory.

### 2. Nexora Adaptive Parser Intelligence (NAPI)

NAPI owns unknown-format evidence, classification, structural inference, candidate generation and measured parser quality. Expensive discovery runs outside the ingestion hot path.

Required behavior:

- Build bounded fingerprints from lexical, delimiter, timestamp, quoting, encoding, token, key, length, source and repeated-structure signals.
- Distinguish known formats, known variants, probable new formats, corrupted input and unresolved ambiguity.
- Support source-free classification and mixed-format sources.
- Select among direct structured decoding, deterministic prefix/state parsing, delimiter-aware tokenization, key-value extraction, named-capture regular expressions, templates, grammars, multiline framing and bounded fallback.
- Rank compatible approved parsers using evidence and measured cost; record why a parser was selected.
- Preserve competing interpretations when material ambiguity remains.
- Select structurally diverse samples, rare examples, counterexamples and held-out evaluation data.
- Convert operator corrections into versioned regression evidence.
- Detect parser, template, schema and distribution drift without changing production automatically.

Candidate confidence must be calibrated against labeled results when available. A successful match or valid JSON object does not prove semantic correctness.

### 3. Parser IR, compiler and execution plans

Introduce a typed, versioned parser intermediate representation only where the existing declarative parser model cannot express the required behavior.

The representation may describe framing, encoding, fingerprints, tokenization, extraction, fields and types, mappings, normalization, validation, nested payloads, multiline behavior, enrichment, fallbacks, confidence evidence, provenance, resource budgets and compatibility.

The compiler or plan builder must:

- Validate schemas and types.
- Reject arbitrary code and unsafe or unbounded operations.
- Bound expression count, regex size and captures, recursion, input, memory and execution work.
- Precompile reusable expressions and invalidate bounded caches by version.
- Apply constant folding or redundant-operation removal only when equivalence is tested.
- Produce reproducible versioned artifacts and useful diagnostics.
- Keep the prior approved parser runnable while a candidate compiles or fails.
- Compare an optimized plan with a reference path on representative and adversarial records.

### 4. Schema, semantics and target formats

Parse once into a canonical, provenance-bearing event, then derive destination-specific representations without rereading the source.

Provide:

- Source-to-canonical-to-destination field mappings.
- Strict type, null, overflow, timestamp, precision and ambiguous-time policies.
- Nested and flattened output mappings.
- Compatibility classification for schema changes.
- Explicit loss analysis for dropped, truncated, merged or precision-reduced fields.
- Destination validation for the adapters actually implemented.
- Field provenance that distinguishes copied, normalized, derived and externally enriched values.
- Inferred semantic meanings clearly separated from operator-verified meanings.

JSONL, Loki and ClickHouse are core destinations. Additional formats such as CSV, Avro, Protobuf or CEF remain planned until their schema, serializer, compatibility and round-trip tests exist.

### 5. Nexora Parsing Autopilot (NPA)

NPA is a controlled parser lifecycle, not an unrestricted autonomous agent.

Use a durable lifecycle equivalent to:

```text
OBSERVE -> DISCOVER -> CLASSIFY -> COLLECT EVIDENCE
-> GENERATE CANDIDATES -> COMPILE -> VALIDATE
-> DIFFERENTIAL TEST -> BENCHMARK -> REVIEW
-> SHADOW -> CANARY -> APPROVE -> PROMOTE
-> MONITOR -> DETECT DRIFT -> PROPOSE IMPROVEMENT
```

Support rejected, failed, cancelled, suspended and rolled-back terminal or recovery states. Every transition carries stable identifiers, versions, tenant/source scope, actor, timestamps, correlation ID, evidence, timeout/cancellation state and recoverable error details.

Automatic discovery, candidate creation and testing may run when policy allows. Production promotion requires configured authorization, separation of duties where required, the exact approved artifact checksum, compatibility evidence, canary gates and a rollback target. Emergency freeze stops new promotions while current approved parsing continues.

### 6. Durability, quarantine and replay

Each material queue must define capacity, admission behavior, high/low watermarks, oldest-item age, concurrency, persistence, retry, cancellation, overflow and cleanup.

Use explicit state transitions for parsing and each destination. Apply bounded exponential backoff with jitter, retry budgets, duplicate-job prevention and circuit breaking where measurements justify it. Records that exhaust automatic recovery stay searchable with their original or reversibly encoded content, reason, parser version, eligibility and history.

Replay must be authorized, bounded, idempotent by job identity, restart-safe, cancellable and observable. Replayed work must not be counted as new source ingestion unless the source was actually reread.

### 7. Independent output delivery

Each destination owns isolated capacity, batching, timeouts, connection reuse, retries and acknowledgements.

- **JSONL:** validate serialization, append/flush behavior, partial-write recovery, permissions, disk exhaustion and rotation.
- **Loki:** validate labels, authentication, TLS, batch limits, rate limiting, retry classes and remote acknowledgement.
- **ClickHouse:** validate table/schema compatibility, field types, batching, authentication, TLS, transient retry and permanent schema failures.

A failed destination cannot block unrelated healthy destinations beyond the explicitly configured source durability policy. Local queue admission is never displayed as remote delivery success.

### 8. Nexora Unified Real-Time Synchronization Fabric (NURSF)

Use one canonical telemetry path:

```text
LOCAL COUNTERS/GAUGES/EVENTS
-> BOUNDED CLIENT AGGREGATION
-> AUTHENTICATED COMPACT PUBLICATION
-> CONTROLLER CANONICAL STATE
-> SNAPSHOT + INCREMENTAL STREAM
-> SHARED BROWSER STORE
-> SUBSCRIBED VIEWS
```

For every metric define owner, type, unit, collection, reset, aggregation, labels, retention, freshness and missing-data semantics. Handle process generations, counter resets, sequence gaps, duplicates, out-of-order updates, partial updates, schema versions, reconnect and snapshot resynchronization.

Critical health and security transitions bypass ordinary chart coalescing. Slow subscribers are isolated and resynchronized instead of accumulating unbounded queues. More open dashboards must not multiply fleet collection work.

One-second publication/freshness is a configurable target. Measure collection, publication, controller processing, transport, browser application and end-to-end delay independently where clocks permit. Report missed targets and clock-skew limitations.

### 9. Command Center and Monitoring

Every panel consumes canonical telemetry or an authoritative bounded historical query. No production panel uses fabricated demo values.

Provide shared time range, pause/live state, refresh policy, and node/source/parser/component/destination/severity/environment filters. Preserve filter and drill-down context. Charts expose units, controller timestamps, exact hover values, legends, thresholds, annotations and freshness.

An incident view should show severity, first/last occurrence, affected entities, evidence, current state, recovery, and a likely cause only when labeled as an inference. Confirmed root causes require confirming evidence.

### 10. Nexora Assurance Lab (NAL)

NAL owns executable proof, not production telemetry or parser activation.

Each test run records the NLI version, parser versions, test and dataset versions, environment, configuration fingerprint, start/end time, status, measurements, errors and evidence references. Use explicit queued, running, passed, failed, blocked, cancelled and incomplete states. An interrupted run cannot become passed.

Cover parser correctness, held-out regression, performance, soak, failure injection, restart/recovery, output outage, telemetry consistency, authorization, installation, upgrade and browser-to-API-to-storage flows. Evidence bundles include checksums and enough context to reproduce the result without retaining unbounded raw data.

Customer demonstrations use synthetic or explicitly authorized data inside isolated workspaces. Saving a test never changes production. Release promotion remains a separate authorized workflow.

### 11. Optional AI assistance

Support disabled, local/offline and explicitly configured provider modes. AI may propose parser candidates, mappings, tests and counterexamples asynchronously. Validate every response against a strict declarative schema and deterministic test gates.

Logs are untrusted data, never model instructions. Provider, model, network, retention, redaction and allowed data classes are policy-controlled. Sensitive samples cannot leave the approved boundary without explicit authorization. Record requests, failures, latency and cost without exposing prompts, credentials or customer payloads. Approved production parsers execute without AI access.

### 12. Enterprise security and maintainability

Enforce least privilege, RBAC, tenant isolation, TLS validation, secret redaction, parser/regex denial-of-service protection, artifact integrity, retention, deletion, immutable audit and safe configuration activation. Preserve supported RPM/systemd installation, upgrade and rollback behavior.

Record dependency and parser-asset provenance, checksums, license review and SBOM evidence. Keep assets with incompatible or unclear commercial rights out of promoted releases.

## WebApp information architecture

Use a small number of coherent top-level workspaces. Reuse a shared view rather than duplicating the same capability under several names.

1. **Command Center**
   - Overview
   - Incidents
   - Capacity
2. **Monitoring**
   - Overview
   - Fleet & Nodes
   - Pipeline
   - Sources
   - Parser Performance
   - Destinations
   - Queues, Replay & Quarantine
   - Infrastructure
   - Latency & Synchronization
   - Errors & Incidents
   - Assurance & Release Health
3. **Ingestion & Execution — NHLE**
   - Sources
   - Framing & Checkpoints
   - Workers & Backpressure
   - Performance Profiles
4. **Parsing Intelligence — NAPI**
   - Discovery
   - Grammar & Structure
   - Candidate Brain
   - Schema & Semantic Studio
   - Compiler
   - Explainability
   - Drift
   - Registry & Lineage
5. **Parsing Autopilot — NPA**
   - Lifecycle
   - Validation & Differential
   - Shadow & Canary
   - Deployments
   - Historical Replay
   - Policies & Freeze
6. **Delivery & Recovery**
   - Outputs
   - Quarantine
   - Replay
   - Durability
7. **Real-Time Sync — NURSF**
   - Fabric Overview
   - Nodes & Sequences
   - Freshness & Latency
   - Contract Coverage
   - Diagnostics
8. **Assurance Lab — NAL**
   - Overview
   - Feature Explorer
   - Test Workbench
   - Parser Assurance
   - Performance & Reliability
   - Integration
   - Security
   - Evidence Vault
   - Release & Promotion
   - Customer Demonstration
   - Readiness Reports
9. **Administration**
   - Identity & RBAC
   - Data Policy
   - AI Providers
   - Configuration
   - Audit
   - Licensing & Provenance

Every workspace requires loading, empty, stale, error and permission-denied states; bounded queries; real API wiring; authorization; audit for mutations; and keyboard-accessible controls. Background updates must preserve scroll, focus, input text and selected subtabs.

## Cross-subsystem integration gates

Verify these paths with stable identifiers and compatible schemas:

1. Reader -> framing -> parser -> normalized or quarantine output.
2. Parser failure -> evidence -> correction -> validation -> replay.
3. Normalized event -> independent JSONL/Loki/ClickHouse state and acknowledgement.
4. Client -> authenticated controller -> canonical state -> incremental stream -> every subscribed view.
5. Candidate -> evidence -> review -> shadow/canary -> approval -> active version -> rollback.
6. Test run -> evidence vault -> readiness -> approval workflow.
7. Operational failure -> critical event -> incident -> drill-down -> recovery verification.
8. Configuration -> validation -> approval where required -> activation -> client acknowledgement -> audit.
9. Source commit -> reproducible artifacts -> RPM install -> systemd health -> upgrade -> rollback.

## Performance and scale method

Measure before optimizing. Record throughput in records/s and bytes/s; p50/p95/p99 stage and end-to-end latency; CPU; resident memory; allocations where supported; queue depth/age; sink batch/ack latency; retry/quarantine rates; restart recovery; telemetry propagation; API latency; and browser rendering cost.

Test small, medium, large, multiline, malformed and adversarial records across common and rare parsers, bursts, sustained traffic, slow and failed destinations, restart with pending work, and concurrent operators. State exact hardware, storage, OS, runtime, configuration, warm-up, input distribution and correctness criteria.

The 5 GiB/10-minute figure is a qualification target for the declared local-file workload. Ten million records/s and 1,000-node control-plane scale remain unclaimed until reproducible tests demonstrate them.

## Implementation order

1. Audit and baseline.
2. Data-loss, corruption and security defects.
3. Bounded ingestion, checkpoints, durability and failure isolation.
4. Canonical telemetry correctness and stale/missing semantics.
5. Parser correctness, evidence and safe candidate generation.
6. Versioned compiler/plan improvements supported by profiling.
7. NPA lifecycle, canary, approval and rollback.
8. NAL evidence and release integration.
9. Consolidated UI and operator workflows.
10. Performance, scale, failure injection, packaging and release proof.

## Release acceptance

V18 is releasable only when:

- The capability matrix reflects actual repository evidence and remaining limits.
- Existing regression behavior remains intact or the versioned change is approved.
- New persisted state has migration, restart and recovery tests.
- Accepted records satisfy the documented lossless/durability contract.
- Parser changes are immutable, validated, auditable and reversible where technically supported.
- Canonical telemetry drives all live views and exposes stale/missing states accurately.
- Queues, caches, samples, retries, subscribers and browser state remain bounded.
- Authentication, authorization, TLS, secrets and tenant boundaries pass applicable tests.
- Performance claims have reproducible measurements with correctness accounting.
- RPM build, fresh install, upgrade, service startup, WebView, hosted WebApp proxy and rollback gates pass on the declared support matrix.
- Checksums, source revision, manifest, SBOM and provenance refer to the exact released bytes.
- Planned, blocked and partially implemented items remain explicitly disclosed.

## Required completion report

Report the inspected revision, baseline, reused features, confirmed gaps, changed files, API/schema/configuration migrations, completed end-to-end paths, tests actually run, benchmark conditions/results, failure-injection evidence, security/licensing review, installation/upgrade/rollback evidence and remaining limitations. Map each completed claim to an acceptance row in `V18_CAPABILITY_AND_ACCEPTANCE_MATRIX.md`.

