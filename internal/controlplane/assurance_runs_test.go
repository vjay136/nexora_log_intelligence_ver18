package controlplane

import (
	"bytes"
	"net/http"
	"testing"
)

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
	if response.StatusCode != 201 {
		t.Fatalf("run: %d %s", response.StatusCode, body)
	}
	run := decodeMap(t, body)
	if run["plan_revision"] != float64(1) || run["plan_sha256"] == "" || run["finished_at"] == nil || bytes.Contains(body, []byte("status=OK")) {
		t.Fatalf("invalid run metadata: %s", body)
	}
	id := run["id"].(string)
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
