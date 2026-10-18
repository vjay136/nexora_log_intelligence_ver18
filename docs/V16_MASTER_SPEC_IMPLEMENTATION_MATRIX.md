# V16 master-spec implementation matrix

This matrix maps the supplied NHLE, NAPI, and Nexora Assurance Lab engineering prompts to repository evidence. It prevents a broad design prompt from being mistaken for a completed or universally proven production capability.

## Implemented and verified in V16

| Area | V16 implementation | Verification evidence |
|---|---|---|
| Concurrent source ingestion | A bounded fixed worker pool scans up to 256 configured sources; automatic sizing is capped at 32 workers and explicit sizing at 64. Bounded per-source batches prevent one scan from monopolizing a worker. | `internal/agent/source_workers_test.go`, 101-source append/restart tests, uncached Go suite. |
| Linux live-file wakeup | Linux inotify watches deduplicated parent directories, coalesces wake signals, and retains periodic reconciliation as the correctness backstop. Watcher failure falls back to reconciliation. | `internal/agent/watcher_linux.go`, append/burst/failure tests. |
| Shutdown durability | Buffered multiline events are flushed before checkpoints; failed flushes leave offsets replayable. | `finishPendingSources` tests in `internal/agent/source_workers_test.go`. |
| Bounded parsing | Physical input, multiline records, candidate evidence, structural samples, AI queues, and controller request bodies have explicit bounds. | Agent/engine/controller tests and `docs/V16_ENGINEERING_AUDIT.md`. |
| Adaptive parser evidence | Unknown and ambiguous records receive bounded structural fingerprints, normalized format/delimiter/key evidence, and ranked candidate metadata. Candidate generation never activates a parser. | `internal/engine/engine.go` and focused fingerprint/ranking tests. |
| Fail-closed routing | Truncated JSON and ambiguous candidates retain extracted evidence but route to quarantine with explicit errors and capped confidence. | `internal/engine/engine_test.go`. |
| Parser assurance | Production parsing, corpus execution, parser creation helper, editable declarative drafts, validation, second-person review, publish, and fleet-sync paths are separated. | Controller tests and browser tests. |
| Versioned test plans | Assurance exposes bounded declarative test plans with immutable revisions, historical reads, RBAC, audit, WebApp creation/revision/download/delete controls, and no automatic production execution. | `/api/v1/assurance/plans`, controller API tests, browser test. |
| Persisted Assurance runs | A bounded synchronous run binds an exact test-plan revision/hash to an encrypted retained fixture, records metadata-only evidence and terminal state, supports listing/download, rejects cancellation, and emits an audit event. It is not a background scheduler. | `/api/v1/assurance/runs`, `internal/controlplane/assurance_runs_test.go`, mocked and packaged browser tests. |
| Input-path rollout | Published profiles target explicit clients or all eligible online clients through an approval proposal while preserving output and secret settings. | Dedicated controller contract and browser tests. |
| Destination isolation | Each output destination has an independent bounded worker queue. Flush is a delivery barrier that waits for all earlier writes and propagates worker failures before source checkpoints advance. | `internal/output/router_queue_test.go`, routing and agent tests. |
| Observability | Dashboard, live streams, per-file throughput, parser latency, quarantine counts, queue depth, API latency, client health, and desired/applied parser state are exposed without customer event payloads. | Dashboard/API/UI tests. |
| Hosted WebApp | Vite build plus same-origin Vercel/Cloudflare proxy adapters require a public HTTPS controller DNS origin and reject private/local origins. | 20 Web/proxy tests and production Vite build. |
| Corpus evidence | The supplied reviewed corpus accounts for 728/728 inputs, 723 exact-parity parsed records, five expected quarantines, 13,101 source/generated fields, and 278/278 known keys. | Generated corpus evidence referenced by `docs/V16_PROGRESS.md`. |

## Implemented with explicit limits

| Requirement | Current boundary |
|---|---|
| Hypersonic or universal throughput | The repository records a 5 GiB benchmark on one declared six-logical-CPU container host. Results are host/workload specific and are not a guarantee for every machine, parser, output, or 100-source production workload. |
| Exact parsing of every possible log | Semantic guarantees apply to reviewed fixtures. Unknown, malformed, oversized, and ambiguous records are preserved and quarantined instead of being silently called correct. |
| Event-driven file handling | Linux inotify reduces detection latency; periodic reconciliation remains authoritative. Rotation behavior is tested through the existing tail/checkpoint implementation, but no claim is made for every network filesystem or platform watcher. |
| Remote output durability | Loki/ClickHouse/controller outputs have bounded batches, retry/backoff, and local spool/fallback behavior. V16 does not claim a distributed exactly-once transaction across remote destinations. |
| AI remediation | AI can propose bounded declarative drafts after deterministic failure. It cannot approve, publish, activate, or execute provider output. |
| Supabase | Supabase stores controller metadata and allowlisted encrypted Assurance/rollout artifacts. Production raw and normalized event streams remain outside Supabase. |
| Platform support | Verified packaging is RPM/systemd on Rocky/RHEL-family x86_64 targets listed in the support matrix. Other distributions and architectures are unverified. |

## Design requirements not claimed as completed

- A resumable multi-stage Assurance scheduler for long-running external jobs, automatic dependency orchestration, crash recovery, and an evidence-vault lifecycle beyond the current synchronous bounded run records and retained fixtures.
- Executable Assurance scenarios beyond the current bounded `parse` and `parser-regression` runner, including replay/recovery, remote-output integration, performance, security, customer evaluation, and interrupted-run recovery.
- The full NAPI parser-factory lifecycle: multi-sample grammar/regex synthesis, regex complexity analysis, held-out and adversarial candidate scoring, optimizer/compiler stages, drift-triggered learning, and measured safe promotion. V16 currently fingerprints bounded records, ranks existing candidates, and can create review-gated declarative drafts with optional AI assistance.
- Customer-isolated demonstration workspaces, customer-specific certification reports, and automatic preproduction/canary orchestration. Existing production activation remains validation, independent approval, signed rollout, client acknowledgement, and manual operational verification.
- Proven 1,000-plus-node control-plane scale or guaranteed 10-million-records-per-second throughput.
- A cross-destination exactly-once ledger spanning files, Loki, ClickHouse, and controller delivery.
- Complete automated production promotion from a test plan. Production activation intentionally remains validation, review, approval, publish, rollout, and client-acknowledgement gated.
- Universal parser correctness for future formats that are absent from the reviewed corpus.

These remaining items require their own architecture, failure model, implementation, and measured acceptance evidence before release claims can include them.
