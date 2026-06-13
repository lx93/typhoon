package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"typhoon/internal/client"
	"typhoon/internal/relay"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "check":
		return runCheck(args[1:])
	case "config":
		return runConfig(args[1:])
	case "connect":
		return runConnect(args[1:])
	case "-h", "--help", "help":
		printUsage()
		return nil
	default:
		return usageError()
	}
}

func runCheck(args []string) error {
	cfg, err := parseCommonFlags("check", args)
	if err != nil {
		return err
	}

	selected, _, err := fetchSelectedRelay(context.Background(), cfg)
	if err != nil {
		return err
	}
	printSelectedRelay(os.Stdout, selected)
	return nil
}

func runConfig(args []string) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	cfg := commonConfig{}
	addCommonFlags(fs, &cfg)
	outPath := fs.String("out", "", "write generated sing-box config to this path; defaults to stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	selected, configJSON, err := fetchSelectedRelay(context.Background(), cfg)
	if err != nil {
		return err
	}

	if *outPath == "" || *outPath == "-" {
		fmt.Print(string(configJSON))
		return nil
	}
	if err := os.WriteFile(*outPath, configJSON, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Fprintf(os.Stdout, "wrote sing-box config for relay %s to %s\n", selected.ID, *outPath)
	return nil
}

func runConnect(args []string) error {
	fs := flag.NewFlagSet("connect", flag.ContinueOnError)
	cfg := commonConfig{}
	addCommonFlags(fs, &cfg)
	singBoxPath := fs.String("sing-box", "sing-box", "path to sing-box binary")
	configOut := fs.String("config-out", "", "optional path for generated sing-box config")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	selected, configJSON, err := fetchSelectedRelay(ctx, cfg)
	if err != nil {
		return err
	}

	configPath, cleanup, err := writeConnectConfig(*configOut, configJSON)
	if err != nil {
		return err
	}
	defer cleanup()

	printSelectedRelay(os.Stdout, selected)
	fmt.Fprintf(os.Stdout, "starting sing-box with config %s\n", configPath)

	runner := client.SingBoxRunner{
		Path:   *singBoxPath,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return runner.Run(ctx, configPath)
}

type commonConfig struct {
	BrokerURL string
	Limit     int
	MTU       int
}

func parseCommonFlags(name string, args []string) (commonConfig, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	cfg := commonConfig{}
	addCommonFlags(fs, &cfg)
	if err := fs.Parse(args); err != nil {
		return commonConfig{}, err
	}
	return cfg, nil
}

func addCommonFlags(fs *flag.FlagSet, cfg *commonConfig) {
	fs.StringVar(&cfg.BrokerURL, "broker", "http://localhost:8080", "broker base URL")
	fs.IntVar(&cfg.Limit, "limit", 5, "relay candidate limit")
	fs.IntVar(&cfg.MTU, "mtu", 0, "TUN MTU; defaults to sing-box config default")
}

func fetchSelectedRelay(ctx context.Context, cfg commonConfig) (relay.Descriptor, []byte, error) {
	broker := client.BrokerClient{BaseURL: cfg.BrokerURL}
	resp, err := broker.ListRelays(ctx, cfg.Limit)
	if err != nil {
		return relay.Descriptor{}, nil, err
	}

	selected, err := client.SelectRelay(resp)
	if err != nil {
		if errors.Is(err, client.ErrNoUsableRelay) {
			return relay.Descriptor{}, nil, fmt.Errorf("no usable relay returned by broker")
		}
		return relay.Descriptor{}, nil, err
	}

	configJSON, err := client.BuildSingBoxConfig(client.SingBoxConfigInput{Relay: selected, MTU: cfg.MTU})
	if err != nil {
		return relay.Descriptor{}, nil, err
	}
	return selected, configJSON, nil
}

func writeConnectConfig(configOut string, configJSON []byte) (string, func(), error) {
	if configOut != "" {
		if err := os.WriteFile(configOut, configJSON, 0o600); err != nil {
			return "", func() {}, fmt.Errorf("write config: %w", err)
		}
		return configOut, func() {}, nil
	}

	file, err := os.CreateTemp("", "typhoon-sing-box-*.json")
	if err != nil {
		return "", func() {}, fmt.Errorf("create temp config: %w", err)
	}
	path := file.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}

	if _, err := file.Write(configJSON); err != nil {
		_ = file.Close()
		cleanup()
		return "", func() {}, fmt.Errorf("write temp config: %w", err)
	}
	if err := file.Close(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("close temp config: %w", err)
	}
	return path, cleanup, nil
}

func printSelectedRelay(out *os.File, selected relay.Descriptor) {
	expires := selected.ExpiresAt.Format(time.RFC3339)
	fmt.Fprintf(
		out,
		"selected relay %s at %s:%d, expires %s\n",
		selected.ID,
		selected.PublicHost,
		selected.PublicPort,
		expires,
	)
}

func usageError() error {
	printUsage()
	return fmt.Errorf("expected one of: check, config, connect")
}

func printUsage() {
	program := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, `Usage:
  %[1]s check   -broker http://localhost:8080
  %[1]s config  -broker http://localhost:8080 -out typhoon-sing-box.json
  %[1]s connect -broker http://localhost:8080 -sing-box /opt/homebrew/bin/sing-box

Commands:
  check    Fetch relay candidates and print the selected usable relay.
  config   Generate a sing-box TUN client config for the selected relay.
  connect  Generate a config and run sing-box to route traffic through Typhoon.

Common flags:
  -mtu     Override the generated TUN MTU, e.g. -mtu 1280 for IPv6 path tests.

`, program)
}
