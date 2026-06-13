package volunteer

import (
	"context"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConnectionObserverLogsAndForwardsTraffic(t *testing.T) {
	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen target: %v", err)
	}
	defer targetListener.Close()

	go func() {
		conn, err := targetListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = io.Copy(conn, conn)
	}()

	listenHost, listenPort := reserveTestPort(t)
	targetPort := targetListener.Addr().(*net.TCPAddr).Port
	var output syncBuffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	observer := &ConnectionObserver{
		ListenHost: listenHost,
		ListenPort: listenPort,
		TargetHost: "127.0.0.1",
		TargetPort: targetPort,
		Output:     &output,
	}
	errCh, err := observer.Start(ctx)
	if err != nil {
		t.Fatalf("start observer: %v", err)
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(listenHost, strconv.Itoa(listenPort)))
	if err != nil {
		t.Fatalf("dial observer: %v", err)
	}
	if _, err := conn.Write([]byte("hello")); err != nil {
		t.Fatalf("write through observer: %v", err)
	}
	buf := make([]byte, 5)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read echo through observer: %v", err)
	}
	if string(buf) != "hello" {
		t.Fatalf("expected echo %q, got %q", "hello", string(buf))
	}
	_ = conn.Close()

	waitForLog(t, &output, "client disconnected")
	logs := output.String()
	for _, want := range []string{"client connected", "client disconnected", "ip=127.0.0.1", "bytes_from_client=5", "bytes_to_client=5"} {
		if !strings.Contains(logs, want) {
			t.Fatalf("expected logs to contain %q, got:\n%s", want, logs)
		}
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("observer returned error after cancel: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("observer did not stop after cancel")
	}
}

func reserveTestPort(t *testing.T) (string, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve listen port: %v", err)
	}
	defer listener.Close()
	return "127.0.0.1", listener.Addr().(*net.TCPAddr).Port
}

func waitForLog(t *testing.T, output *syncBuffer, needle string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(output.String(), needle) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for log %q; got:\n%s", needle, output.String())
}

type syncBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}
