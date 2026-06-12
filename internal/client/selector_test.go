package client

import (
	"errors"
	"testing"
	"time"

	"typhoon/internal/relay"
)

func TestSelectRelaySelectsFirstUsableRelay(t *testing.T) {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	wrongProtocol := validRelay(now)
	wrongProtocol.ID = "wrong_protocol"
	wrongProtocol.Protocol = "unknown"

	second := validRelay(now)
	second.ID = "selected"

	selected, err := SelectRelay(relay.ListResponse{
		ServerTime: now,
		Relays:     []relay.Descriptor{wrongProtocol, second},
	})
	if err != nil {
		t.Fatalf("select relay: %v", err)
	}
	if selected.ID != "selected" {
		t.Fatalf("expected selected relay, got %q", selected.ID)
	}
}

func TestSelectRelaySkipsExpiredRelay(t *testing.T) {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	expired := validRelay(now)
	expired.ExpiresAt = now.Add(-time.Second)

	_, err := SelectRelay(relay.ListResponse{
		ServerTime: now,
		Relays:     []relay.Descriptor{expired},
	})
	if !errors.Is(err, ErrNoUsableRelay) {
		t.Fatalf("expected ErrNoUsableRelay, got %v", err)
	}
}

func TestSelectRelaySkipsIncompleteRelay(t *testing.T) {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	incomplete := validRelay(now)
	incomplete.RealityPublicKey = ""

	_, err := SelectRelay(relay.ListResponse{
		ServerTime: now,
		Relays:     []relay.Descriptor{incomplete},
	})
	if !errors.Is(err, ErrNoUsableRelay) {
		t.Fatalf("expected ErrNoUsableRelay, got %v", err)
	}
}

func TestIsUsableRelayRequiresDirectVLESSRealityVision(t *testing.T) {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	candidate := validRelay(now)
	candidate.ExitMode = relay.ExitModeDedicated

	if IsUsableRelay(candidate, now) {
		t.Fatal("expected dedicated exit relay to be unusable for MVP client")
	}
}

func validRelay(now time.Time) relay.Descriptor {
	return relay.Descriptor{
		ID:               "relay_123",
		PublicHost:       "volunteer.example.com",
		PublicPort:       443,
		Protocol:         relay.ProtocolVLESSRealityVision,
		ClientID:         "2c08df10-4ef4-4ab9-95c6-cb1e94cdb2ff",
		RealityPublicKey: "public-key",
		ShortID:          "5f7a8d9c01ab23cd",
		ServerName:       "www.microsoft.com",
		Flow:             relay.FlowVision,
		ExitMode:         relay.ExitModeDirect,
		MaxSessions:      8,
		MaxMbps:          20,
		VolunteerVersion: "dev",
		RegisteredAt:     now.Add(-time.Minute),
		LastHeartbeatAt:  now.Add(-time.Second),
		ExpiresAt:        now.Add(time.Minute),
	}
}
