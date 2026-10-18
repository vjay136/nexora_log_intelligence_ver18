package controlplane

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nexora.local/nli-v16/internal/middleware"
	metastore "nexora.local/nli-v16/internal/store"
)

// Run records contain metadata only. The source sample stays in the encrypted
// fixture store and is never copied into a run revision.
type assuranceRunRecord struct {
	SchemaVersion int            `json:"schema_version"`
	ID            string         `json:"id"`
	PlanID        string         `json:"plan_id"`
	PlanRevision  int            `json:"plan_revision"`
	PlanSHA256    string         `json:"plan_sha256"`
	PlanVersion   string         `json:"plan_version"`
	Scenario      string         `json:"scenario"`
	CaseID        string         `json:"case_id"`
	State         string         `json:"state"`
	QueuedAt      time.Time      `json:"queued_at"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	FinishedAt    *time.Time     `json:"finished_at,omitempty"`
	Evidence      map[string]any `json:"evidence,omitempty"`
	Error         string         `json:"error,omitempty"`
}

func (s *Server) persistAssuranceRun(r *http.Request, record assuranceRunRecord, actor string) error {
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = s.metadata.Put(ctx, metastore.Document{Collection: "assurance-runs", ID: record.ID, State: record.State, Payload: payload, CreatedBy: actor})
	return err
}

func (s *Server) handleAssuranceRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if s.authorize(w, r, middleware.PermissionReadParsers) == nil {
			return
		}
		limit := 50
		if raw := r.URL.Query().Get("limit"); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value < 1 || value > 100 {
				jsonError(w, "limit must be 1..100", 400)
				return
			}
			limit = value
		}
		docs, err := s.metadata.List(r.Context(), "assurance-runs", metastore.Query{Limit: limit})
		if err != nil {
			jsonError(w, "Assurance runs unavailable", 503)
			return
		}
		runs := make([]any, 0, len(docs))
		for _, doc := range docs {
			runs = append(runs, documentView(doc))
		}
		jsonOK(w, map[string]any{"runs": runs, "count": len(runs)})
		return
	}
	if r.Method != http.MethodPost {
		jsonError(w, "GET or POST required", 405)
		return
	}
	principal := s.authorize(w, r, middleware.PermissionAssuranceSave)
	if principal == nil {
		return
	}
	var input struct {
		PlanID       string `json:"plan_id"`
		PlanRevision int    `json:"plan_revision"`
		CaseID       string `json:"case_id"`
		Scenario     string `json:"scenario"`
	}
	if decodeJSON(r, &input, 4096) != nil || input.PlanID == "" || safeID(input.PlanID) != input.PlanID || input.CaseID == "" || safeID(input.CaseID) != input.CaseID || input.PlanRevision < 1 {
		jsonError(w, "plan_id, positive plan_revision and case_id are required", 400)
		return
	}
	planDoc, err := s.metadata.Get(r.Context(), "assurance-plans", input.PlanID, input.PlanRevision)
	if err != nil {
		jsonError(w, "Assurance plan revision not found", 404)
		return
	}
	var plan struct {
		Version string `json:"version"`
		Plan    struct {
			Scenarios []string `json:"scenarios"`
		} `json:"plan"`
	}
	if json.Unmarshal(planDoc.Payload, &plan) != nil {
		jsonError(w, "Assurance plan is invalid", 503)
		return
	}
	scenario := input.Scenario
	if scenario == "" {
		scenario = "parser-regression"
	}
	allowed := false
	for _, item := range plan.Plan.Scenarios {
		if item == scenario {
			allowed = true
		}
	}
	if !allowed || (scenario != "parser-regression" && scenario != "parse") {
		jsonError(w, "scenario is not supported by the plan runner", 422)
		return
	}
	_, secret, err := s.decryptAssurance(r, input.CaseID)
	if err != nil {
		jsonError(w, "retained Assurance case not found", 404)
		return
	}
	if len(secret.Raw) == 0 || len(secret.Raw) > 1<<20 {
		jsonError(w, "case size is outside run bounds", 422)
		return
	}
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		jsonError(w, "run identifier unavailable", 500)
		return
	}
	now := time.Now().UTC()
	record := assuranceRunRecord{SchemaVersion: 1, ID: "run-" + hex.EncodeToString(idBytes), PlanID: input.PlanID, PlanRevision: planDoc.Revision, PlanSHA256: planDoc.Hash, PlanVersion: plan.Version, Scenario: scenario, CaseID: input.CaseID, State: "queued", QueuedAt: now}
	if err := s.persistAssuranceRun(r, record, principal.Username); err != nil {
		jsonError(w, "run persistence failed", 503)
		return
	}
	record.State = "running"
	record.StartedAt = &now
	if err := s.persistAssuranceRun(r, record, principal.Username); err != nil {
		jsonError(w, "run start persistence failed", 503)
		return
	}
	result := s.runAssurance(assuranceRequest{Raw: secret.Raw, Desired: secret.Desired, Service: secret.Service, Source: secret.Source, FieldMapping: secret.FieldMapping})
	finished := time.Now().UTC()
	record.FinishedAt = &finished
	if result["passed"] == true {
		record.State = "passed"
	} else {
		record.State = "failed"
	}
	// Only bounded, non-content evidence crosses the metadata boundary.
	record.Evidence = map[string]any{"raw_sha256": result["raw_sha256"], "passed": result["passed"], "score": result["score"], "counts": result["counts"], "route": result["route"], "latency_us": result["latency_us"]}
	if err := s.persistAssuranceRun(r, record, principal.Username); err != nil {
		jsonError(w, "run result persistence failed", 503)
		return
	}
	outcome := "success"
	if record.State != "passed" {
		outcome = "failure"
	}
	s.auditAction(principal, "assurance.run", "assurance-runs/"+record.ID, outcome, map[string]string{"state": record.State, "plan_revision": fmt.Sprint(record.PlanRevision)})
	jsonCreated(w, record)
}

func (s *Server) handleAssuranceRunRecord(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/assurance/runs/"), "/")
	parts := strings.Split(path, "/")
	id := parts[0]
	if id == "" || safeID(id) != id || len(parts) > 2 {
		jsonError(w, "invalid run id", 400)
		return
	}
	if r.Method == http.MethodGet && len(parts) == 1 {
		if s.authorize(w, r, middleware.PermissionReadParsers) == nil {
			return
		}
		doc, err := s.metadata.Get(r.Context(), "assurance-runs", id, 0)
		if err != nil {
			jsonError(w, "Assurance run not found", 404)
			return
		}
		jsonOK(w, documentView(doc))
		return
	}
	if r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "cancel" {
		if s.authorize(w, r, middleware.PermissionAssuranceSave) == nil {
			return
		}
		// Runs execute synchronously in the creating request. Advertising a
		// cancellable queued window would introduce a state race between the
		// request and a concurrent cancel call. A future durable scheduler may
		// add an atomic compare-and-swap cancellation transition.
		jsonError(w, "synchronous Assurance runs cannot be cancelled", 409)
		return
	}
	jsonError(w, "unsupported run operation", 405)
}
