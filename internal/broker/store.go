package broker

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"time"

	"typhoon/internal/relay"
)

var ErrRelayNotFound = errors.New("relay not found")

type Store struct {
	mu     sync.RWMutex
	relays map[string]relay.Descriptor
}

func NewStore() *Store {
	return &Store{relays: make(map[string]relay.Descriptor)}
}

func (s *Store) Register(req relay.RegisterRequest, now time.Time, ttl time.Duration) (relay.Descriptor, error) {
	id, err := newRelayID()
	if err != nil {
		return relay.Descriptor{}, err
	}

	desc := relay.Descriptor{
		ID:               id,
		PublicHost:       req.PublicHost,
		PublicPort:       req.PublicPort,
		Protocol:         req.Protocol,
		ClientID:         req.ClientID,
		RealityPublicKey: req.RealityPublicKey,
		ShortID:          req.ShortID,
		ServerName:       req.ServerName,
		Flow:             req.Flow,
		ExitMode:         req.ExitMode,
		MaxSessions:      req.MaxSessions,
		MaxMbps:          req.MaxMbps,
		VolunteerVersion: req.VolunteerVersion,
		RegisteredAt:     now,
		LastHeartbeatAt:  now,
		ExpiresAt:        now.Add(ttl),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.relays[id] = desc

	return desc, nil
}

func (s *Store) Heartbeat(id string, now time.Time, ttl time.Duration) (relay.Descriptor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	desc, ok := s.relays[id]
	if !ok {
		return relay.Descriptor{}, ErrRelayNotFound
	}

	desc.LastHeartbeatAt = now
	desc.ExpiresAt = now.Add(ttl)
	s.relays[id] = desc

	return desc, nil
}

func (s *Store) List(now time.Time, limit int) []relay.Descriptor {
	s.Prune(now)

	s.mu.RLock()
	defer s.mu.RUnlock()

	relays := make([]relay.Descriptor, 0, len(s.relays))
	for _, desc := range s.relays {
		if desc.ExpiresAt.After(now) {
			relays = append(relays, desc)
		}
	}

	sort.Slice(relays, func(i, j int) bool {
		iIPv6 := relay.IsIPv6Host(relays[i].PublicHost)
		jIPv6 := relay.IsIPv6Host(relays[j].PublicHost)
		if iIPv6 != jIPv6 {
			return iIPv6
		}
		return relays[i].LastHeartbeatAt.After(relays[j].LastHeartbeatAt)
	})

	if limit > 0 && len(relays) > limit {
		return relays[:limit]
	}

	return relays
}

func (s *Store) Prune(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, desc := range s.relays {
		if !desc.ExpiresAt.After(now) {
			delete(s.relays, id)
		}
	}
}

func newRelayID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "relay_" + hex.EncodeToString(b[:]), nil
}
