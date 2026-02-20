package wazuhapi

import (
	"testing"
)

func TestParseSCALiveTableFrame(t *testing.T) {
	t.Parallel()

	raw := []byte(`[
		{
			"agent": "fedora",
			"agent_id": "001",
			"policy_id": "cis_fedora",
			"policy": "CIS Benchmark",
			"description": "CIS",
			"score": 88.5,
			"pass": 80,
			"fail": 10,
			"total_checks": 90,
			"end_scan": "2026-05-21T10:00:00Z"
		}
	]`)

	frame, err := ParseSCALiveTableFrame(raw, "A")
	if err != nil {
		t.Fatal(err)
	}
	if frame.Fields[0].At(0) != "fedora" {
		t.Fatalf("unexpected agent: %v", frame.Fields[0].At(0))
	}
}
