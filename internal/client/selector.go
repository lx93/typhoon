package client

import (
	"errors"
	"time"

	"typhoon/internal/relay"
)

var ErrNoUsableRelay = errors.New("no usable relay")

func SelectRelay(resp relay.ListResponse) (relay.Descriptor, error) {
	now := resp.ServerTime
	if now.IsZero() {
		now = time.Now()
	}

	var fallback relay.Descriptor
	hasFallback := false
	for _, candidate := range resp.Relays {
		if !IsUsableRelay(candidate, now) {
			continue
		}
		if relay.IsIPv6Host(candidate.PublicHost) {
			return candidate, nil
		}
		if !hasFallback {
			fallback = candidate
			hasFallback = true
		}
	}
	if hasFallback {
		return fallback, nil
	}
	return relay.Descriptor{}, ErrNoUsableRelay
}

func IsUsableRelay(candidate relay.Descriptor, now time.Time) bool {
	return candidate.Protocol == relay.ProtocolVLESSRealityVision &&
		candidate.Flow == relay.FlowVision &&
		candidate.ExitMode == relay.ExitModeDirect &&
		candidate.ExpiresAt.After(now) &&
		hasRequiredConnectionFields(candidate)
}

func hasRequiredConnectionFields(candidate relay.Descriptor) bool {
	return candidate.PublicHost != "" &&
		candidate.PublicPort > 0 &&
		candidate.ClientID != "" &&
		candidate.RealityPublicKey != "" &&
		candidate.ShortID != "" &&
		candidate.ServerName != ""
}
