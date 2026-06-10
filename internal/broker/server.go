package broker

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"typhoon/internal/relay"
)

type Config struct {
	RegistrationToken string
	VolunteerLeaseTTL time.Duration
}

func NewServer(store *Store, cfg Config) http.Handler {
	if cfg.VolunteerLeaseTTL == 0 {
		cfg.VolunteerLeaseTTL = 3 * time.Minute
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	mux.HandleFunc("POST /api/v1/volunteers/register", registerHandler(store, cfg))
	mux.HandleFunc("POST /api/v1/volunteers/", heartbeatHandler(store, cfg))
	mux.HandleFunc("GET /api/v1/relays", listRelaysHandler(store))

	return mux
}

func registerHandler(store *Store, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r, cfg.RegistrationToken) {
			writeError(w, http.StatusUnauthorized, "missing or invalid volunteer registration token")
			return
		}

		var req relay.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON request")
			return
		}

		if err := validateRegisterRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		desc, err := store.Register(req, time.Now().UTC(), cfg.VolunteerLeaseTTL)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not register relay")
			return
		}

		writeJSON(w, http.StatusCreated, desc)
	}
}

func heartbeatHandler(store *Store, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r, cfg.RegistrationToken) {
			writeError(w, http.StatusUnauthorized, "missing or invalid volunteer registration token")
			return
		}

		id, ok := heartbeatRelayID(r.URL.Path)
		if !ok {
			writeError(w, http.StatusNotFound, "unknown volunteer endpoint")
			return
		}

		desc, err := store.Heartbeat(id, time.Now().UTC(), cfg.VolunteerLeaseTTL)
		if errors.Is(err, ErrRelayNotFound) {
			writeError(w, http.StatusNotFound, "relay not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not update relay heartbeat")
			return
		}

		writeJSON(w, http.StatusOK, relay.HeartbeatResponse{OK: true, ExpiresAt: desc.ExpiresAt})
	}
}

func listRelaysHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 5
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > 20 {
				writeError(w, http.StatusBadRequest, "limit must be between 1 and 20")
				return
			}
			limit = parsed
		}

		now := time.Now().UTC()
		relays := store.List(now, limit)
		writeJSON(w, http.StatusOK, relay.ListResponse{
			Count:      len(relays),
			ServerTime: now,
			Relays:     relays,
		})
	}
}

func validateRegisterRequest(req relay.RegisterRequest) error {
	switch {
	case req.PublicHost == "":
		return errors.New("public_host is required")
	case req.PublicPort < 1 || req.PublicPort > 65535:
		return errors.New("public_port must be between 1 and 65535")
	case req.Protocol != relay.ProtocolVLESSRealityVision:
		return errors.New("protocol must be vless-reality-vision")
	case req.ClientID == "":
		return errors.New("client_id is required")
	case req.RealityPublicKey == "":
		return errors.New("reality_public_key is required")
	case req.ShortID == "":
		return errors.New("short_id is required")
	case req.ServerName == "":
		return errors.New("server_name is required")
	case req.Flow != relay.FlowVision:
		return errors.New("flow must be xtls-rprx-vision")
	case req.ExitMode != relay.ExitModeDirect && req.ExitMode != relay.ExitModeDedicated:
		return errors.New("exit_mode must be direct or dedicated")
	case req.MaxSessions < 1:
		return errors.New("max_sessions must be at least 1")
	case req.MaxMbps < 1:
		return errors.New("max_mbps must be at least 1")
	default:
		return nil
	}
}

func heartbeatRelayID(path string) (string, bool) {
	const prefix = "/api/v1/volunteers/"
	const suffix = "/heartbeat"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	return id, id != ""
}

func authorized(r *http.Request, token string) bool {
	if token == "" {
		return true
	}
	return r.Header.Get("Authorization") == "Bearer "+token
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, relay.ErrorResponse{Error: message})
}
