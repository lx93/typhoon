package broker

import (
	"testing"
	"time"

	"typhoon/internal/relay"
)

func TestStorePrunesExpiredRelays(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 6, 9, 7, 0, 0, 0, time.UTC)

	desc, err := store.Register(validRegisterRequest(), now, time.Minute)
	if err != nil {
		t.Fatalf("register relay: %v", err)
	}

	if got := store.List(now.Add(30*time.Second), 10); len(got) != 1 {
		t.Fatalf("expected relay before expiration, got %d", len(got))
	}

	if got := store.List(desc.ExpiresAt.Add(time.Nanosecond), 10); len(got) != 0 {
		t.Fatalf("expected relay to be pruned after expiration, got %d", len(got))
	}
}

func TestHeartbeatExtendsRelayLease(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 6, 9, 7, 0, 0, 0, time.UTC)

	desc, err := store.Register(validRegisterRequest(), now, time.Minute)
	if err != nil {
		t.Fatalf("register relay: %v", err)
	}

	heartbeatAt := now.Add(30 * time.Second)
	updated, err := store.Heartbeat(desc.ID, heartbeatAt, time.Minute)
	if err != nil {
		t.Fatalf("heartbeat relay: %v", err)
	}

	if !updated.ExpiresAt.Equal(heartbeatAt.Add(time.Minute)) {
		t.Fatalf("expected expiration %s, got %s", heartbeatAt.Add(time.Minute), updated.ExpiresAt)
	}
}

func validRegisterRequest() relay.RegisterRequest {
	return relay.RegisterRequest{
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
		VolunteerVersion: "test",
	}
}
