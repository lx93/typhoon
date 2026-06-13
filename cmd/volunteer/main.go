package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"typhoon/internal/relay"
	"typhoon/internal/volunteer"
)

const version = "dev"

func main() {
	var cfg cliConfig
	flag.StringVar(&cfg.BrokerURL, "broker", "http://localhost:8080", "broker base URL")
	flag.StringVar(&cfg.RegistrationToken, "registration-token", os.Getenv("TYPHOON_VOLUNTEER_TOKEN"), "volunteer registration token")
	flag.StringVar(&cfg.XrayPath, "xray", "xray", "path to xray binary")
	flag.StringVar(&cfg.ListenHost, "listen-host", "::", "local listen host")
	flag.IntVar(&cfg.ListenPort, "listen-port", 443, "local listen port")
	flag.StringVar(&cfg.PublicHost, "public-host", "", "public hostname or IP clients can reach; defaults to this machine's first global IPv6 address")
	flag.IntVar(&cfg.PublicPort, "public-port", 443, "public port clients can reach")
	flag.StringVar(&cfg.ServerName, "server-name", "www.microsoft.com", "Reality server name")
	flag.StringVar(&cfg.RealityDest, "reality-dest", "www.microsoft.com:443", "Reality dest")
	flag.StringVar(&cfg.ClientID, "client-id", "", "VLESS client UUID; generated when empty")
	flag.StringVar(&cfg.RealityPrivateKey, "reality-private-key", "", "Reality private key; generated with xray x25519 when empty")
	flag.StringVar(&cfg.RealityPublicKey, "reality-public-key", "", "Reality public key; generated with xray x25519 when empty")
	flag.StringVar(&cfg.ShortID, "short-id", "", "Reality short ID; generated when empty")
	flag.IntVar(&cfg.MaxSessions, "max-sessions", 8, "advertised max client sessions")
	flag.IntVar(&cfg.MaxMbps, "max-mbps", 20, "advertised max Mbps")
	flag.DurationVar(&cfg.HeartbeatInterval, "heartbeat-interval", 30*time.Second, "broker heartbeat interval")
	flag.StringVar(&cfg.ConfigOut, "config-out", "", "write generated Xray config to this path")
	flag.BoolVar(&cfg.ConnectionLog, "connection-log", true, "print colored client connect and disconnect events")
	flag.BoolVar(&cfg.PrintConfigOnly, "print-config-only", false, "print generated Xray config and exit")
	flag.BoolVar(&cfg.SkipXrayRun, "skip-xray-run", false, "register and heartbeat without launching xray")
	flag.Parse()

	if err := cfg.ApplyDefaults(); err != nil {
		slog.Error("invalid volunteer config", "error", err)
		os.Exit(2)
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid volunteer config", "error", err)
		os.Exit(2)
	}

	if err := run(cfg); err != nil {
		slog.Error("volunteer stopped", "error", err)
		os.Exit(1)
	}
}

type cliConfig struct {
	BrokerURL         string
	RegistrationToken string
	XrayPath          string
	ListenHost        string
	ListenPort        int
	PublicHost        string
	PublicPort        int
	ServerName        string
	RealityDest       string
	ClientID          string
	RealityPrivateKey string
	RealityPublicKey  string
	ShortID           string
	MaxSessions       int
	MaxMbps           int
	HeartbeatInterval time.Duration
	ConfigOut         string
	ConnectionLog     bool
	PrintConfigOnly   bool
	SkipXrayRun       bool
}

func (c *cliConfig) ApplyDefaults() error {
	if c.PublicHost != "" || c.PrintConfigOnly {
		return nil
	}
	publicIPv6, err := volunteer.DefaultPublicIPv6Address()
	if err != nil {
		return fmt.Errorf("public-host is required when no global IPv6 address can be auto-detected: %w", err)
	}
	c.PublicHost = publicIPv6
	return nil
}

func (c cliConfig) Validate() error {
	if c.BrokerURL == "" {
		return fmt.Errorf("broker is required")
	}
	if c.PublicHost == "" && !c.PrintConfigOnly {
		return fmt.Errorf("public-host is required")
	}
	if c.ListenPort < 1 || c.ListenPort > 65535 {
		return fmt.Errorf("listen-port must be between 1 and 65535")
	}
	if c.PublicPort < 1 || c.PublicPort > 65535 {
		return fmt.Errorf("public-port must be between 1 and 65535")
	}
	if c.MaxSessions < 1 {
		return fmt.Errorf("max-sessions must be at least 1")
	}
	if c.MaxMbps < 1 {
		return fmt.Errorf("max-mbps must be at least 1")
	}
	if c.HeartbeatInterval < 5*time.Second {
		return fmt.Errorf("heartbeat-interval must be at least 5s")
	}
	return nil
}

func run(cfg cliConfig) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	xrayCfg := cfg
	if cfg.ConnectionLog && !cfg.SkipXrayRun && !cfg.PrintConfigOnly {
		targetHost, targetPort, err := volunteer.ReserveLoopbackTCPPort()
		if err != nil {
			return err
		}
		xrayCfg.ListenHost = targetHost
		xrayCfg.ListenPort = targetPort
	}

	prepared, err := prepareRuntime(xrayCfg)
	if err != nil {
		return err
	}

	if cfg.PrintConfigOnly {
		fmt.Println(string(prepared.XrayConfig))
		return nil
	}

	configPath := cfg.ConfigOut
	if configPath == "" {
		configPath = filepath.Join(os.TempDir(), "typhoon-xray-config.json")
	}
	if err := os.WriteFile(configPath, prepared.XrayConfig, 0o600); err != nil {
		return fmt.Errorf("write xray config: %w", err)
	}
	slog.Info("wrote xray config", "path", configPath)

	var xrayCmd *exec.Cmd
	var errCh <-chan error
	var observerErrCh <-chan error
	if !cfg.SkipXrayRun {
		xrayCmd = exec.CommandContext(ctx, cfg.XrayPath, "run", "-config", configPath)
		xrayCmd.Stdout = os.Stdout
		xrayCmd.Stderr = os.Stderr
		if err := xrayCmd.Start(); err != nil {
			return fmt.Errorf("start xray: %w", err)
		}
		waitCh := make(chan error, 1)
		go func() {
			waitCh <- xrayCmd.Wait()
		}()
		errCh = waitCh
		slog.Info("started xray", "pid", xrayCmd.Process.Pid)

		if cfg.ConnectionLog {
			observer := &volunteer.ConnectionObserver{
				ListenHost: cfg.ListenHost,
				ListenPort: cfg.ListenPort,
				TargetHost: xrayCfg.ListenHost,
				TargetPort: xrayCfg.ListenPort,
				Output:     os.Stdout,
			}
			observerErrCh, err = observer.Start(ctx)
			if err != nil {
				stopProcess(xrayCmd, errCh)
				return fmt.Errorf("start connection observer: %w", err)
			}
			slog.Info(
				"started connection observer",
				"listen",
				fmt.Sprintf("%s:%d", cfg.ListenHost, cfg.ListenPort),
				"target",
				fmt.Sprintf("%s:%d", xrayCfg.ListenHost, xrayCfg.ListenPort),
			)
		}
	}

	desc, err := register(ctx, cfg, prepared)
	if err != nil {
		if xrayCmd != nil {
			stopProcess(xrayCmd, errCh)
		}
		return err
	}
	slog.Info("registered relay", "id", desc.ID, "public", fmt.Sprintf("%s:%d", desc.PublicHost, desc.PublicPort))

	ticker := time.NewTicker(cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if xrayCmd != nil {
				stopProcess(xrayCmd, errCh)
			}
			return nil
		case err := <-errCh:
			if err == nil {
				return fmt.Errorf("xray exited")
			}
			return fmt.Errorf("xray exited: %w", err)
		case err, ok := <-observerErrCh:
			if !ok {
				observerErrCh = nil
				continue
			}
			if err != nil {
				if xrayCmd != nil {
					stopProcess(xrayCmd, errCh)
				}
				return fmt.Errorf("connection observer stopped: %w", err)
			}
		case <-ticker.C:
			if err := heartbeat(ctx, cfg, desc.ID); err != nil {
				slog.Warn("heartbeat failed", "error", err)
				continue
			}
			slog.Info("heartbeat ok", "id", desc.ID)
		}
	}
}

type preparedRuntime struct {
	ClientID         string
	RealityPublicKey string
	ShortID          string
	XrayConfig       []byte
}

func prepareRuntime(cfg cliConfig) (preparedRuntime, error) {
	clientID := cfg.ClientID
	if clientID == "" {
		generated, err := volunteer.GenerateUUID()
		if err != nil {
			return preparedRuntime{}, fmt.Errorf("generate client ID: %w", err)
		}
		clientID = generated
	}

	shortID := cfg.ShortID
	if shortID == "" {
		generated, err := volunteer.GenerateShortID()
		if err != nil {
			return preparedRuntime{}, fmt.Errorf("generate short ID: %w", err)
		}
		shortID = generated
	}

	privateKey := cfg.RealityPrivateKey
	publicKey := cfg.RealityPublicKey
	if privateKey == "" || publicKey == "" {
		keyPair, err := volunteer.GenerateRealityKeyPair(cfg.XrayPath)
		if err != nil {
			return preparedRuntime{}, err
		}
		privateKey = keyPair.PrivateKey
		publicKey = keyPair.PublicKey
	}

	xrayConfig, err := volunteer.BuildXrayConfig(volunteer.XrayConfigInput{
		ListenHost:        cfg.ListenHost,
		ListenPort:        cfg.ListenPort,
		ClientID:          clientID,
		Flow:              relay.FlowVision,
		Dest:              cfg.RealityDest,
		ServerName:        cfg.ServerName,
		RealityPrivateKey: privateKey,
		ShortID:           shortID,
	})
	if err != nil {
		return preparedRuntime{}, err
	}

	return preparedRuntime{
		ClientID:         clientID,
		RealityPublicKey: publicKey,
		ShortID:          shortID,
		XrayConfig:       xrayConfig,
	}, nil
}

func register(ctx context.Context, cfg cliConfig, prepared preparedRuntime) (relay.Descriptor, error) {
	req := relay.RegisterRequest{
		PublicHost:       cfg.PublicHost,
		PublicPort:       cfg.PublicPort,
		Protocol:         relay.ProtocolVLESSRealityVision,
		ClientID:         prepared.ClientID,
		RealityPublicKey: prepared.RealityPublicKey,
		ShortID:          prepared.ShortID,
		ServerName:       cfg.ServerName,
		Flow:             relay.FlowVision,
		ExitMode:         relay.ExitModeDirect,
		MaxSessions:      cfg.MaxSessions,
		MaxMbps:          cfg.MaxMbps,
		VolunteerVersion: version,
	}

	var desc relay.Descriptor
	if err := postJSON(ctx, cfg, "/api/v1/volunteers/register", req, &desc); err != nil {
		return relay.Descriptor{}, err
	}
	return desc, nil
}

func heartbeat(ctx context.Context, cfg cliConfig, id string) error {
	var resp relay.HeartbeatResponse
	return postJSON(ctx, cfg, "/api/v1/volunteers/"+id+"/heartbeat", map[string]bool{"ok": true}, &resp)
}

func postJSON(ctx context.Context, cfg cliConfig, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := strings.TrimRight(cfg.BrokerURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.RegistrationToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.RegistrationToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr relay.ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Error == "" {
			apiErr.Error = resp.Status
		}
		return fmt.Errorf("broker %s: %s", path, apiErr.Error)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return err
		}
	}

	return nil
}

func stopProcess(cmd *exec.Cmd, errCh <-chan error) {
	if cmd.Process == nil {
		return
	}

	_ = cmd.Process.Signal(os.Interrupt)

	select {
	case <-errCh:
		return
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
	}

	select {
	case <-errCh:
	case <-time.After(time.Second):
	}
}
