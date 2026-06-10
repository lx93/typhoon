package relay

import "time"

const (
	ProtocolVLESSRealityVision = "vless-reality-vision"
	FlowVision                 = "xtls-rprx-vision"
	ExitModeDirect             = "direct"
	ExitModeDedicated          = "dedicated"
)

type RegisterRequest struct {
	PublicHost       string `json:"public_host"`
	PublicPort       int    `json:"public_port"`
	Protocol         string `json:"protocol"`
	ClientID         string `json:"client_id"`
	RealityPublicKey string `json:"reality_public_key"`
	ShortID          string `json:"short_id"`
	ServerName       string `json:"server_name"`
	Flow             string `json:"flow"`
	ExitMode         string `json:"exit_mode"`
	MaxSessions      int    `json:"max_sessions"`
	MaxMbps          int    `json:"max_mbps"`
	VolunteerVersion string `json:"volunteer_version"`
}

type Descriptor struct {
	ID               string    `json:"id"`
	PublicHost       string    `json:"public_host"`
	PublicPort       int       `json:"public_port"`
	Protocol         string    `json:"protocol"`
	ClientID         string    `json:"client_id"`
	RealityPublicKey string    `json:"reality_public_key"`
	ShortID          string    `json:"short_id"`
	ServerName       string    `json:"server_name"`
	Flow             string    `json:"flow"`
	ExitMode         string    `json:"exit_mode"`
	MaxSessions      int       `json:"max_sessions"`
	MaxMbps          int       `json:"max_mbps"`
	VolunteerVersion string    `json:"volunteer_version"`
	RegisteredAt     time.Time `json:"registered_at"`
	LastHeartbeatAt  time.Time `json:"last_heartbeat_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type ListResponse struct {
	Count      int          `json:"count"`
	ServerTime time.Time    `json:"server_time"`
	Relays     []Descriptor `json:"relays"`
}

type HeartbeatResponse struct {
	OK        bool      `json:"ok"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
