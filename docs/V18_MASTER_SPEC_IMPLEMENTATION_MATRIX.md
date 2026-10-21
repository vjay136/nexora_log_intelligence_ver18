# V18 master-spec implementation matrix

This matrix maps the supplied NHLE, NAPI, and Nexora Assurance Lab engineering prompts to repository evidence. It prevents a broad design prompt from being mistaken for a completed or universally proven production capability.

## Implemented and verified in V18

| Area | V18 implementation | Verification evidence |
|---|---|---|
| Concurrent source ingestion | A bounded fixed worker pool scans up to 256 configured sources; automatic sizing is capped at 32 workers and explicit sizing at 64. Bounded per-source batches prevent one scan from monopolizing a worker. | `internal/agent/source_workers_test.go`, 101-source append/restart tests, uncached Go suite. |
| Linux live-file wakeup | Linux inotify watches deduplicated parent directories, coalesces wake signals, and retains periodic reconciliation as the correctness backstop. Watcher failure falls back to reconciliation. | `internal/agent/watcher_linux.go`, append/burst/failure tests. |
| Shutdown durability | Buffered multiline events are flushed before checkpoints; failed flushes leave offsets replayable. | `finishPendingSources` tests in `internal/agent/source_workers_test.go`. |
| Bounded parsing | Physical input, multiline records, candidate evidence, structural samples, AI queues, and controller request bodies have explicit bounds. | Agent/engine/controller tests and `docs/V18_ENGINEERING_AUDIT.md`. |
| Adaptive parser evidence | Unknown and ambiguous records receive bounded structural fingerprints, normalized format/delimiter/key evidence, and ranked candidate metadata. Candidate generation never activates a parser. | `internal/engine/engine.go` and focused fingerprint/ranking tests. |
| Fail-closed routing | Truncated JSON and ambiguous candidates retain extracted evidence but route to quarantine with explicit errors and capped confidence. | `internal/engine/engine_test.go`. |
| Parser assurance | Production parsing, corpus execution, parser creation helper, editable declarative drafts, validation, second-person review, publish, and fleet-sync paths are separated. | Controller tests and browser tests. |
| Versioned test plans | Assurance exposes bounded declarative test plans with immutable revisions, historical reads, RBAC, audit, WebApp creation/revision/download/delete controls, and no automatic production execution. | `/api/v1/assurance/plans`, controller API tests, browser test. |
| Durable Assurance runs | A bounded two-worker scheduler uses the metadata store as its durable queue, binds every run to an exact test-plan revision/hash and encrypted retained fixture, atomically serializes claim/cancel/finalize transitions within the active controller, re-queues interrupted `running` records after restart, records metadata-only evidence, and supports list/download/cancel controls. | `/api/v1/assurance/runs`, `internal/controlplane/assurance_run_scheduler.go`, restart/cancellation/controller tests, mocked and packaged browser tests. |
| Bounded NAPI synthesis | An authenticated off-hot-path endpoint generates multiple declarative candidates from existing JSON, key-value, anchored-regex, and registered parser primitives. It separates discovery from held-out samples, checks exact field output and negative false positives, statically bounds RE2 pattern size/captures/instructions, persists hashes and explainable evidence without raw samples, and creates a draft only after every gate passes. Promotion still uses independent review and publish. | `/api/v1/assurance/napi/synthesize`, `internal/controlplane/napi_synthesis.go`, focused acceptance/rejection tests. |
| Executable local Assurance scenarios | Exact plan revisions can run parser regression, controlled failure/recovery, deterministic replay, local destination JSON-contract, bounded performance, and fixture privacy/integrity scenarios. Every scenario is finite, metadata-only, and leaves production parsers and destinations unchanged. | `internal/controlplane/assurance_run_scenarios.go`, scenario API tests, Assurance Runs WebApp controls. |
| Input-path rollout | Published profiles target explicit clients or all eligible online clients through an approval proposal while preserving output and secret settings. | Dedicated controller contract and browser tests. |
| Destination isolation | Each output destination has an independent bounded worker queue. Flush is a delivery barrier that waits for all earlier writes and propagates worker failures before source checkpoints advance. | `internal/output/router_queue_test.go`, routing and agent tests. |
| Observability | Dashboard, live streams, per-file throughput, parser latency, quarantine counts, queue depth, API latency, client health, and desired/applied parser state are exposed without customer event payloads. | Dashboard/API/UI tests. |
| Hosted WebApp | Vite build plus same-origin Vercel/Cloudflare proxy adapters require a public HTTPS controller DNS origin and reject private/local origins. | 20 Web/proxy tests and production Vite build. |
| Corpus evidence | The supplied reviewed corpus accounts for 728/728 inputs, 723 exact-parity parsed records, five expected quarantines, 13,101 source/generated fields, and 278/278 known keys. | Generated corpus evidence referenced by `docs/V18_PROGRESS.md`. |

## Implemented with explicit limits

| Requirement | Current boundary |
|---|---|
| Hypersonic or universal throughput | The repository records a 5 GiB benchmark on one declared six-logical-CPU container host. Results are host/workload specific and are not a guarantee for every machine, parser, output, or 100-source production workload. |
| Exact parsing of every possible log | Semantic guarantees apply to reviewed fixtures. Unknown, malformed, oversized, and ambiguous records are preserved and quarantined instead of being silently called correct. |
| Event-driven file handling | Linux inotify reduces detection latency; periodic reconciliation remains authoritative. Rotation behavior is tested through the existing tail/checkpoint implementation, but no claim is made for every network filesystem or platform watcher. |
| Remote output durability | Loki/ClickHouse/controller outputs have bounded batches, retry/backoff, and local spool/fallback behavior. V18 does not claim a distributed exactly-once transaction across remote destinations. |
| AI remediation | AI can propose bounded declarative drafts after deterministic failure. It cannot approve, publish, activate, or execute provider output. |
| Supabase | Supabase stores controller metadata and allowlisted encrypted Assurance/rollout artifacts. Production raw and normalized event streams remain outside Supabase. |
| Platform support | Verified packaging is RPM/systemd on Rocky/RHEL-family x86_64 targets listed in the support matrix. Other distributions and architectures are unverified. |

## Design requirements not claimed as completed

- A distributed multi-controller scheduler for long-running external jobs, automatic dependency graphs, stage-level resume, and cross-region worker leases. V18.0.0-1 provides restart recovery for bounded local runs on one active controller; it does not claim distributed job execution.
- Remote-destination integration runs, customer evaluation, interrupted-run recovery, and distributed performance/security test orchestration. V18's corresponding Assurance scenarios are bounded local contract checks; they deliberately do not contact production Loki/ClickHouse/controller destinations or claim load-test capacity.
- The complete NAPI lifecycle remains broader than the bounded synthesis path: optimizer/compiler equivalence passes, automatic drift-triggered learning, canary traffic orchestration, measured rollback thresholds, and general grammar synthesis for arbitrary unseen formats are not implemented. V18 provides deterministic multi-sample JSON/KV/anchored-regex candidate generation, held-out and negative scoring, bounded RE2 analysis, evidence-backed drafts, and the existing independent approval workflow.
- Customer-isolated demonstration workspaces, customer-specific certification reports, and automatic preproduction/canary orchestration. Existing production activation remains validation, independent approval, signed rollout, client acknowledgement, and manual operational verification.
- Proven 1,000-plus-node control-plane scale or guaranteed 10-million-records-per-second throughput.
- A cross-destination exactly-once ledger spanning files, Loki, ClickHouse, and controller delivery.
- Complete automated production promotion from a test plan. Production activation intentionally remains validation, review, approval, publish, rollout, and client-acknowledgement gated.
- Universal parser correctness for future formats that are absent from the reviewed corpus.

These remaining items require their own architecture, failure model, implementation, and measured acceptance evidence before release claims can include them.
