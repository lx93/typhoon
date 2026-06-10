package broker

import (
	"testing"

	"typhoon/internal/relay"
)

func TestValidateRegisterRequest(t *testing.T) {
	req := validRegisterRequest()
	if err := validateRegisterRequest(req); err != nil {
		t.Fatalf("expected valid request: %v", err)
	}

	req.Protocol = "unknown"
	if err := validateRegisterRequest(req); err == nil {
		t.Fatal("expected protocol validation error")
	}
}

func TestHeartbeatRelayID(t *testing.T) {
	id, ok := heartbeatRelayID("/api/v1/volunteers/relay_abc/heartbeat")
	if !ok || id != "relay_abc" {
		t.Fatalf("expected relay_abc, got id=%q ok=%v", id, ok)
	}
}

func TestValidateExitModes(t *testing.T) {
	req := validRegisterRequest()
	req.ExitMode = relay.ExitModeDedicated
	if err := validateRegisterRequest(req); err != nil {
		t.Fatalf("expected dedicated exit mode to be schema-valid: %v", err)
	}
}
