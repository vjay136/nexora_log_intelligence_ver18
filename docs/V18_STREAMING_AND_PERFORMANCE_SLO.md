# NLI V18 streaming and performance SLO

This document defines the measurements used to accept V18. It does not turn a
hardware-dependent throughput target into a product guarantee.

## Operator story

An operator can follow each enrolled client's active input files from the
first read through parsing and output delivery. The controller shows bounded,
non-customer telemetry only: counters, rates, queue state, phase duration,
source path, node identity, timestamps, and health. Raw log bodies and parsed
customer fields remain in the configured data destinations.

## Latency and freshness budgets

| Signal | Target | Measurement |
|---|---:|---|
| Agent-to-controller stream telemetry | 2 seconds by default, configurable 1–60 seconds | Difference between adjacent samples |
| Controller stream publication | 250 ms p95 after receipt | Controller receive timestamp to SSE write |
| Live dashboard freshness | 2 seconds p95 on a healthy LAN | Agent sample timestamp to DOM update |
| Snapshot API | 100 ms p95 for 1,000 active sources | Handler benchmark with a populated bounded store |
| API monitor refresh | One bounded parallel batch | Wall time, p50, p95, failure count |
| Slow stream subscriber impact | Zero parser blocking | Full subscriber queue drops/replaces telemetry, never parser work |

The WebApp uses one authenticated server-sent-event connection for live
telemetry. It reconnects with bounded exponential backoff and uses the snapshot
endpoint as a polling fallback. It does not poll every controller API to update
the live counters.

## Throughput qualification

The requested 5 GiB in ten minutes requires at least 8.53 MiB/s of sustained
end-to-end throughput. Five GiB in one minute requires 85.33 MiB/s. V18 records
the measured host, storage, parser mix, worker count, output destination, input
size, records, elapsed time, bytes per second, records per second, normalized
count, quarantine count, and output acknowledgement count.

Qualification passes only when all of these are true:

1. `read = normalized + quarantined + deliberately_filtered` for the selected
   input policy.
2. Every normalized record has a successful output acknowledgement or remains
   in the durable output retry spool.
3. Every quarantine record retains its exact replay reference.
4. Resident memory remains bounded as input size grows.
5. Per-source output order matches input order.
6. The measured sustained rate is at least 8.53 MiB/s for the tested local-file
   destination.

Remote Loki and ClickHouse rates are reported independently because network
latency, endpoint throttling, and acknowledgements can limit delivery. The
parser applies bounded backpressure instead of discarding records.

## Parallel pipeline

Each source is a streaming reader and ordered parser lane with a monotonically
increasing checkpoint. A local client config may contain 1–256 input files.
Independent sources progress concurrently through a fixed worker pool and a
ready queue bounded by that input count; a single source remains ordered
without a reorder buffer. Worker count defaults to the smaller of logical
CPUs, 32, and the configured source count, and can be set from one through
64. Each source processes at most `source_batch_lines` before yielding; the
default is 1,000 and the accepted range is 1–10,000. This keeps active sources
fair when one file has a large backlog. No stage reads an entire file into
memory. Each configured destination owns a bounded worker queue so a slow
remote destination cannot serialize writes to healthy destinations. The agent
flushes every queue and waits for its durable barrier before advancing that
source's checkpoint. Queue admission is reported as `queued`; it is never
called a destination acknowledgement. A source that spends at least half of a
telemetry interval waiting on output is marked as backpressured. Per-source
telemetry includes byte and record rates, cumulative parse/output time, source
error count, and aggregate output queue depth/capacity.

At end of file, the reader reports `eof_wait`, retains its offset, and waits for
new data. When data appears it reports `resumed`. Rotation, truncation,
permission failure, parsing failure, output retry, and recovery are structured,
rate-limited journal events. Journal messages never include raw log bodies or
credentials.

## High-volume thresholds

Thresholds are configuration, not parser decisions. V18 evaluates configured
read-lines, bytes, quarantine-rate, and queue-depth thresholds against the
actual aggregate destination queue depth/capacity. The default queue threshold
is 80 percent when no explicit depth is configured. The UI also marks telemetry
stale after twice the configured sampling interval and shows the exact breached
condition and timestamp.

## Verified qualification result

On the recorded Linux/amd64 Docker host (six logical CPUs, four sources and four
source workers), V18 processed 5 GiB of generated JSONL through the production
agent, reviewed ParserPack attempt/fallback path, durable local JSONL output, and
durable checkpoints in 295.596 seconds (17.32 MiB/s). All 1,310,720 records and
5,368,709,120 bytes were accounted for, output JSON was valid, checkpoint bytes
and lines matched every source, and quarantine was zero. The bounded post-run
validation took a further 57.672 seconds and is outside the processing SLO.
The machine-readable, host-specific report is
[`docs/evidence/THROUGHPUT_5G.json`](evidence/THROUGHPUT_5G.json).

## Public edge behavior

NGINX and Apache are reverse proxies. They can terminate TLS and expose the
controller on an existing routable address, but cannot allocate a public IP or
register public DNS. Installer modes therefore verify the address, DNS records,
ports, and certificate identity before claiming that the controller is public.
A tunnel mode may create a URL; temporary quick-tunnel URLs are test-only and
can change after restart.
