package controlplane

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

var assuranceRunScenarios = map[string]struct{}{
	"parse":                   {},
	"parser-regression":       {},
	"failure-recovery":        {},
	"replay":                  {},
	"destination-integration": {},
	"performance":             {},
	"security":                {},
}

func supportedAssuranceRunScenario(scenario string) bool {
	_, ok := assuranceRunScenarios[scenario]
	return ok
}

// executeAssuranceRunScenario executes bounded local checks. It deliberately
// does not contact configured production destinations or mutate parser state.
// Evidence contains hashes, counts, timings and booleans only; customer fields
// and raw fixture content remain in the encrypted fixture store.
func (s *Server) executeAssuranceRunScenario(scenario string, request assuranceRequest) map[string]any {
	switch scenario {
	case "failure-recovery":
		return s.runFailureRecoveryScenario(request)
	case "replay":
		return s.runReplayScenario(request)
	case "destination-integration":
		return s.runDestinationIntegrationScenario(request)
	case "performance":
		return s.runPerformanceScenario(request)
	case "security":
		return s.runSecurityScenario(request)
	default:
		return baseRunEvidence(s.runAssurance(request))
	}
}

func baseRunEvidence(result map[string]any) map[string]any {
	return map[string]any{
		"raw_sha256": result["raw_sha256"],
		"passed":     result["passed"],
		"score":      result["score"],
		"counts":     result["counts"],
		"route":      result["route"],
		"latency_us": result["latency_us"],
	}
}

func (s *Server) runFailureRecoveryScenario(request assuranceRequest) map[string]any {
	// Force a deterministic comparison failure without changing the retained
	// fixture, then prove the original request can run successfully afterward.
	probe := request
	probe.Desired = map[string]any{"__nli_controlled_failure_probe__": "expected"}
	failed := s.runAssurance(probe)
	recovered := s.runAssurance(request)
	evidence := baseRunEvidence(recovered)
	failureObserved := failed["passed"] != true
	recoveryPassed := recovered["passed"] == true
	evidence["passed"] = failureObserved && recoveryPassed
	evidence["checks"] = map[string]any{
		"controlled_failure_observed": failureObserved,
		"recovery_passed":             recoveryPassed,
		"attempts":                    2,
	}
	return evidence
}

func (s *Server) runReplayScenario(request assuranceRequest) map[string]any {
	first := s.runAssurance(request)
	second := s.runAssurance(request)
	firstHash := stableResultHash(first)
	secondHash := stableResultHash(second)
	evidence := baseRunEvidence(second)
	stable := firstHash != "" && firstHash == secondHash
	evidence["passed"] = first["passed"] == true && second["passed"] == true && stable
	evidence["checks"] = map[string]any{
		"replays":          2,
		"deterministic":    stable,
		"first_sha256":     firstHash,
		"second_sha256":    secondHash,
		"both_runs_passed": first["passed"] == true && second["passed"] == true,
	}
	return evidence
}

func stableResultHash(result map[string]any) string {
	stable := map[string]any{"actual": result["actual"], "route": result["route"], "parser": result["parser"], "counts": result["counts"]}
	encoded, err := json.Marshal(stable)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func (s *Server) runDestinationIntegrationScenario(request assuranceRequest) map[string]any {
	result := s.runAssurance(request)
	// Validate the same bounded JSON document contract used at an output
	// boundary. No configured Loki/ClickHouse endpoint is contacted by a lab run.
	document := map[string]any{"fields": result["actual"], "route": result["route"], "parser": result["parser"]}
	encoded, encodeErr := json.Marshal(document)
	var decoded map[string]any
	decodeErr := json.Unmarshal(encoded, &decoded)
	digest := sha256.Sum256(encoded)
	roundTrip := encodeErr == nil && decodeErr == nil && len(encoded) > 2
	evidence := baseRunEvidence(result)
	evidence["passed"] = result["passed"] == true && roundTrip
	evidence["checks"] = map[string]any{
		"contract":         "local-json-boundary-v1",
		"json_round_trip":  roundTrip,
		"encoded_bytes":    len(encoded),
		"document_sha256":  hex.EncodeToString(digest[:]),
		"external_io_used": false,
	}
	return evidence
}

func (s *Server) runPerformanceScenario(request assuranceRequest) map[string]any {
	const iterations = 16
	latencies := make([]int64, 0, iterations)
	passed := 0
	started := time.Now()
	var last map[string]any
	for i := 0; i < iterations; i++ {
		begin := time.Now()
		last = s.runAssurance(request)
		latencies = append(latencies, time.Since(begin).Microseconds())
		if last["passed"] == true {
			passed++
		}
	}
	elapsed := time.Since(started)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	evidence := baseRunEvidence(last)
	evidence["passed"] = passed == iterations
	evidence["checks"] = map[string]any{
		"iterations":           iterations,
		"passed_iterations":    passed,
		"elapsed_us":           elapsed.Microseconds(),
		"p50_latency_us":       percentileLatency(latencies, 50),
		"p95_latency_us":       percentileLatency(latencies, 95),
		"records_per_second":   float64(iterations) / elapsed.Seconds(),
		"bounded_local_sample": true,
	}
	return evidence
}

func percentileLatency(sorted []int64, percentile int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	index := (len(sorted)*percentile + 99) / 100
	if index < 1 {
		index = 1
	}
	if index > len(sorted) {
		index = len(sorted)
	}
	return sorted[index-1]
}

func (s *Server) runSecurityScenario(request assuranceRequest) map[string]any {
	result := s.runAssurance(request)
	evidence := baseRunEvidence(result)
	encoded, err := json.Marshal(evidence)
	plaintextAbsent := err == nil && (request.Raw == "" || !strings.Contains(string(encoded), request.Raw))
	hashValid := evidence["raw_sha256"] == hex.EncodeToString(hashBytes([]byte(request.Raw)))
	evidence["passed"] = result["passed"] == true && plaintextAbsent && hashValid
	evidence["checks"] = map[string]any{
		"plaintext_absent":         plaintextAbsent,
		"fixture_hash_verified":    hashValid,
		"fixture_within_1mib":      len(request.Raw) <= 1<<20,
		"production_state_mutated": false,
		"external_io_used":         false,
	}
	return evidence
}
