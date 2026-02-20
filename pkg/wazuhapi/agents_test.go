package wazuhapi

import (
	"testing"
)

func TestParseAgentsFrame(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"data": {
			"affected_items": [
				{
					"id": "001",
					"name": "ubuntu",
					"status": "active",
					"ip": "10.0.0.5",
					"version": "v4.8.0",
					"lastKeepAlive": "2026-05-21T10:00:00+00:00",
					"os": {"name": "Ubuntu", "platform": "ubuntu", "version": "22.04"}
				}
			]
		}
	}`)

	frame, err := ParseAgentsFrame(raw, "A")
	if err != nil {
		t.Fatal(err)
	}
	if frame.Fields[1].At(0) != "ubuntu" {
		t.Fatalf("unexpected agent name: %v", frame.Fields[1].At(0))
	}
	if frame.Fields[2].At(0) != "active" {
		t.Fatalf("unexpected status: %v", frame.Fields[2].At(0))
	}
}

func TestParseAgentOptions(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"data": {
			"affected_items": [
				{"id": "001", "name": "fedora"},
				{"id": "000", "name": "wazuh.manager"}
			]
		}
	}`)

	options, err := ParseAgentOptions(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}
	if options[0].Value != "fedora" || options[0].Label != "fedora (001)" {
		t.Fatalf("unexpected first option: %+v", options[0])
	}
}
