package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"typhoon/internal/broker"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	leaseTTL := flag.Duration("lease-ttl", 3*time.Minute, "volunteer relay lease TTL")
	flag.Parse()

	store := broker.NewStore()
	handler := broker.NewServer(store, broker.Config{
		RegistrationToken: os.Getenv("TYPHOON_VOLUNTEER_TOKEN"),
		VolunteerLeaseTTL: *leaseTTL,
	})

	slog.Info("starting broker", "addr", *addr, "lease_ttl", leaseTTL.String())
	if err := http.ListenAndServe(*addr, handler); err != nil {
		slog.Error("broker stopped", "error", err)
		os.Exit(1)
	}
}
