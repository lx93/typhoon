.PHONY: test fmt broker volunteer

fmt:
	gofmt -w cmd internal

test:
	go test ./...

broker:
	go run ./cmd/broker -addr :8080

volunteer:
	go run ./cmd/volunteer \
		-broker http://localhost:8080 \
		-public-host 127.0.0.1 \
		-public-port 443 \
		-client-id 2c08df10-4ef4-4ab9-95c6-cb1e94cdb2ff \
		-reality-private-key dev-private-key \
		-reality-public-key dev-public-key \
		-short-id 5f7a8d9c01ab23cd \
		-skip-xray-run
