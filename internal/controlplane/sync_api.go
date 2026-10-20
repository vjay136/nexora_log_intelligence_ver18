package controlplane

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	metastore "nexora.local/nli-v17/internal/store"
	"nexora.local/nli-v17/telemetry"
)

func (s *Server) handleSyncSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "GET required", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	jsonOK(w, s.canonicalSyncSnapshot(r.Context()))
}

func (s *Server) canonicalSyncSnapshot(parent context.Context) map[string]any {
	snapshot := s.streamTelemetry.CanonicalSnapshot()
	ctx, cancel := context.WithTimeout(parent, 1500*time.Millisecond)
	defer cancel()
	return map[string]any{
		"schema_version":     snapshot.SchemaVersion,
		"revision":           snapshot.Revision,
		"generated_at":       snapshot.GeneratedAt,
		"target_interval_ms": snapshot.TargetIntervalMS,
		"aggregate":          snapshot.Aggregate,
		"sources":            snapshot.Sources,
		"diagnostics":        snapshot.Diagnostics,
		"subsystems":         s.subsystemSyncSnapshot(ctx, snapshot),
	}
}

func (s *Server) subsystemSyncSnapshot(ctx context.Context, snapshot telemetry.SyncSnapshot) map[string]any {
	nhleStatus, nhleUpdatedAt := subsystemFreshness(snapshot.Sources, snapshot.GeneratedAt)
	activeNodes := map[string]struct{}{}
	for _, source := range snapshot.Sources {
		if source.Freshness == telemetry.FreshnessLive || source.Freshness == telemetry.FreshnessDelayed {
			activeNodes[source.NodeID] = struct{}{}
		}
	}
	nhle := map[string]any{
		"status": nhleStatus, "updated_at": nhleUpdatedAt,
		"source_count": len(snapshot.Sources), "active_nodes": len(activeNodes),
		"records_read_total":    snapshot.Aggregate.SourceRecordsReadTotal,
		"normalized_total":      snapshot.Aggregate.NormalizedRecordsTotal,
		"quarantined_total":     snapshot.Aggregate.QuarantinedRecordsTotal,
		"queue_depth":           snapshot.Aggregate.QueueDepth,
		"backpressure_sources":  snapshot.Aggregate.BackpressureSourceCount,
		"read_lines_per_second": snapshot.Aggregate.ReadLinesPerSecond,
		"normalized_per_second": snapshot.Aggregate.NormalizedPerSecond,
	}
	requests, _, failures, _ := s.aiMetrics.snapshot()
	napi := map[string]any{"status": telemetry.FreshnessUnavailable, "requests_total": requests, "failures_total": failures, "pending_drafts": 0}
	if docs, err := s.metadata.List(ctx, "ai-proposals", metastore.Query{Limit: 1000}); err == nil {
		pending := 0
		var latest time.Time
		for _, doc := range docs {
			if doc.State == "draft" || doc.State == "review" {
				pending++
			}
			if doc.UpdatedAt.After(latest) {
				latest = doc.UpdatedAt
			}
		}
		napi["pending_drafts"] = pending
		napi["status"] = telemetry.FreshnessLive
		napi["updated_at"] = snapshot.GeneratedAt
		if !latest.IsZero() {
			napi["latest_proposal_at"] = latest
		}
	} else {
		napi["status"] = telemetry.FreshnessCollectionFailed
	}
	nal := map[string]any{"status": telemetry.FreshnessUnavailable, "total": 0, "queued": 0, "running": 0, "passed": 0, "failed": 0, "cancelled": 0}
	if docs, err := s.metadata.List(ctx, "assurance-runs", metastore.Query{Limit: 1000}); err == nil {
		counts := map[string]int{}
		var latest time.Time
		for _, doc := range docs {
			state := doc.State
			if state == "" {
				var record assuranceRunRecord
				if json.Unmarshal(doc.Payload, &record) == nil {
					state = record.State
				}
			}
			counts[state]++
			if doc.UpdatedAt.After(latest) {
				latest = doc.UpdatedAt
			}
		}
		nal["total"] = len(docs)
		for _, state := range []string{"queued", "running", "passed", "failed", "cancelled"} {
			nal[state] = counts[state]
		}
		nal["status"] = telemetry.FreshnessLive
		nal["updated_at"] = snapshot.GeneratedAt
		if !latest.IsZero() {
			nal["latest_run_at"] = latest
		}
	} else {
		nal["status"] = telemetry.FreshnessCollectionFailed
	}
	return map[string]any{"nhle": nhle, "napi": napi, "nal": nal}
}

func subsystemFreshness(sources []telemetry.StreamFileTelemetry, now time.Time) (string, any) {
	if len(sources) == 0 {
		return telemetry.FreshnessUnavailable, nil
	}
	status := telemetry.FreshnessDisconnected
	var latest time.Time
	rank := map[string]int{telemetry.FreshnessDisconnected: 0, telemetry.FreshnessStale: 1, telemetry.FreshnessDelayed: 2, telemetry.FreshnessLive: 3}
	for _, source := range sources {
		if rank[source.Freshness] > rank[status] {
			status = source.Freshness
		}
		if source.ControllerReceivedAt.After(latest) {
			latest = source.ControllerReceivedAt
		}
	}
	if latest.IsZero() {
		return freshnessFromAge(now.Sub(sources[0].UpdatedAt)), sources[0].UpdatedAt
	}
	return status, latest
}

func freshnessFromAge(age time.Duration) string {
	if age < 0 {
		age = 0
	}
	switch {
	case age <= 2*time.Second:
		return telemetry.FreshnessLive
	case age <= 5*time.Second:
		return telemetry.FreshnessDelayed
	case age <= 15*time.Second:
		return telemetry.FreshnessStale
	default:
		return telemetry.FreshnessDisconnected
	}
}
