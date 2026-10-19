package controlplane

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func waitAssuranceRun(t *testing.T, f *apiFixture, id string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, body := f.request(http.MethodGet, "/api/v1/assurance/runs/"+id, nil)
		if response.StatusCode != http.StatusOK {
			t.Fatalf("get run: %d %s", response.StatusCode, body)
		}
		doc := decodeMap(t, body)
		payload, _ := doc["payload"].(map[string]any)
		if payload == nil {
			payload = doc
		}
		state, _ := payload["state"].(string)
		if state == "passed" || state == "failed" || state == "cancelled" {
			return payload
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("Assurance run %s did not reach a terminal state", id)
	return nil
}

func TestAssuranceRunUsesExactPlanRevisionAndRetainedCase(t *testing.T) {
	f := newAPIFixture(t)
	plan := map[string]any{"objective": "Qualify parser", "environment": "staging", "scenarios": []any{"parser-regression"}, "pass_criteria": []any{"all fixtures pass"}}
	response, body := f.request(http.MethodPost, "/api/v1/assurance/plans", map[string]any{"id": "qualification", "name": "Qualification", "version": "1.0.0", "plan": plan})
	if response.StatusCode != 201 {
		t.Fatalf("plan: %d %s", response.StatusCode, body)
	}
	response, body = f.request(http.MethodPost, "/api/v1/assurance/cases", map[string]any{"id": "run-case", "name": "Run case", "retain": true, "case": map[string]any{"raw": "status=OK", "desired": map[string]any{"status": "OK"}}})
	if response.StatusCode != 201 {
		t.Fatalf("case: %d %s", response.StatusCode, body)
	}
	response, body = f.request(http.MethodPost, "/api/v1/assurance/runs", map[string]any{"plan_id": "qualification", "plan_revision": 1, "case_id": "run-case"})
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("run: %d %s", response.StatusCode, body)
	}
	run := decodeMap(t, body)
	if run["plan_revision"] != float64(1) || run["plan_sha256"] == "" || run["state"] != "queued" || bytes.Contains(body, []byte("status=OK")) {
		t.Fatalf("invalid run metadata: %s", body)
	}
	id := run["id"].(string)
	run = waitAssuranceRun(t, f, id)
	if run["state"] != "passed" || run["finished_at"] == nil {
		t.Fatalf("run did not pass: %#v", run)
	}
	response, body = f.request(http.MethodGet, "/api/v1/assurance/runs/"+id, nil)
	if response.StatusCode != 200 || !bytes.Contains(body, []byte(id)) {
		t.Fatalf("get: %d %s", response.StatusCode, body)
	}
	response, body = f.request(http.MethodGet, "/api/v1/assurance/runs", nil)
	if response.StatusCode != 200 || !bytes.Contains(body, []byte(id)) {
		t.Fatalf("list: %d %s", response.StatusCode, body)
	}
	response, _ = f.request(http.MethodPost, "/api/v1/assurance/runs/"+id+"/cancel", map[string]any{})
	if response.StatusCode != 409 {
		t.Fatalf("terminal run cancelled: %d", response.StatusCode)
	}
	response, _ = f.request(http.MethodPost, "/api/v1/assurance/runs", map[string]any{"plan_id": "qualification", "plan_revision": 2, "case_id": "run-case"})
	if response.StatusCode != 404 {
		t.Fatalf("missing plan revision accepted: %d", response.StatusCode)
	}
}

func TestAssuranceRunScenariosProduceBoundedEvidence(t *testing.T) {
	f := newAPIFixture(t)
	scenarios := []any{"failure-recovery", "replay", "destination-integration", "performance", "security"}
	plan := map[string]any{"objective": "Exercise bounded runner checks", "environment": "synthetic-lab", "scenarios": scenarios, "pass_criteria": []any{"all local checks pass"}}
	response, body := f.request(http.MethodPost, "/api/v1/assurance/plans", map[string]any{"id": "scenario-plan", "name": "Scenario plan", "version": "1.0.0", "plan": plan})
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("plan: %d %s", response.StatusCode, body)
	}
	raw := "status=OK request_id=req-1"
	response, body = f.request(http.MethodPost, "/api/v1/assurance/cases", map[string]any{"id": "scenario-case", "name": "Scenario case", "retain": true, "case": map[string]any{"raw": raw, "desired": map[string]any{"status": "OK", "request_id": "req-1"}}})
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("case: %d %s", response.StatusCode, body)
	}

	for _, item := range scenarios {
		scenario := item.(string)
		t.Run(scenario, func(t *testing.T) {
			response, body := f.request(http.MethodPost, "/api/v1/assurance/runs", map[string]any{"plan_id": "scenario-plan", "plan_revision": 1, "case_id": "scenario-case", "scenario": scenario})
			if response.StatusCode != http.StatusAccepted {
				t.Fatalf("run: %d %s", response.StatusCode, body)
			}
			run := decodeMap(t, body)
			run = waitAssuranceRun(t, f, run["id"].(string))
			if run["state"] != "passed" || run["scenario"] != scenario {
				t.Fatalf("unexpected run: %#v", run)
			}
			encoded, _ := json.Marshal(run)
			if bytes.Contains(encoded, []byte(raw)) || len(encoded) > 32<<10 {
				t.Fatalf("run evidence is not bounded metadata: %d bytes %s", len(encoded), encoded)
			}
			evidence, ok := run["evidence"].(map[string]any)
			if !ok || evidence["raw_sha256"] == "" || evidence["passed"] != true {
				t.Fatalf("missing evidence: %s", body)
			}
			checks, ok := evidence["checks"].(map[string]any)
			if !ok || len(checks) == 0 {
				t.Fatalf("missing scenario checks: %s", body)
			}
		})
	}
}

func TestAssuranceRunCancelAndRestartRecoveryAreDurable(t *testing.T) {
	f := newAPIFixture(t)
	f.server.assuranceRuns.Close()
	now := time.Now().UTC()
	queued := assuranceRunRecord{SchemaVersion: 1, ID: "run-cancel-test", State: "queued", Scenario: "parse", CaseID: "case", QueuedAt: now, RequestedBy: "admin", RequestedRole: "admin"}
	if err := f.server.persistAssuranceRun(queued, "admin"); err != nil {
		t.Fatal(err)
	}
	cancelled, err := f.server.assuranceRuns.Cancel(queued.ID, "admin")
	if err != nil || cancelled.State != "cancelled" || cancelled.FinishedAt == nil {
		t.Fatalf("cancel: %#v %v", cancelled, err)
	}

	running := assuranceRunRecord{SchemaVersion: 1, ID: "run-recovery-test", State: "running", Scenario: "parse", CaseID: "case", QueuedAt: now, StartedAt: &now, RequestedBy: "admin", RequestedRole: "admin"}
	if err := f.server.persistAssuranceRun(running, "admin"); err != nil {
		t.Fatal(err)
	}
	probe := &assuranceRunScheduler{server: f.server}
	if pending, err := probe.recoverInterrupted(); err != nil {
		t.Fatal(err)
	} else if !pending {
		t.Fatal("expected recovered run to remain pending")
	}
	recovered, err := probe.load(running.ID)
	if err != nil || recovered.State != "queued" || recovered.StartedAt != nil || recovered.RecoveryCount != 1 {
		t.Fatalf("recovery: %#v %v", recovered, err)
	}
}

func TestAssuranceRunRejectsUnsupportedOrUnplannedScenario(t *testing.T) {
	f := newAPIFixture(t)
	plan := map[string]any{"objective": "Parse only", "environment": "synthetic-lab", "scenarios": []any{"parse"}, "pass_criteria": []any{"parse passes"}}
	response, body := f.request(http.MethodPost, "/api/v1/assurance/plans", map[string]any{"id": "parse-only", "name": "Parse only", "version": "1.0.0", "plan": plan})
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("plan: %d %s", response.StatusCode, body)
	}
	response, body = f.request(http.MethodPost, "/api/v1/assurance/cases", map[string]any{"id": "parse-case", "name": "Parse case", "retain": true, "case": map[string]any{"raw": "status=OK"}})
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("case: %d %s", response.StatusCode, body)
	}
	for _, scenario := range []string{"performance", "shell-command"} {
		response, _ = f.request(http.MethodPost, "/api/v1/assurance/runs", map[string]any{"plan_id": "parse-only", "plan_revision": 1, "case_id": "parse-case", "scenario": scenario})
		if response.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("scenario %q accepted: %d", scenario, response.StatusCode)
		}
	}
}
