package volunteer

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ansiGreen = "\x1b[32m"
	ansiRed   = "\x1b[31m"
	ansiReset = "\x1b[0m"
)

type ConnectionObserver struct {
	ListenHost string
	ListenPort int
	TargetHost string
	TargetPort int
	Output     io.Writer

	nextID atomic.Uint64
}

func ReserveLoopbackTCPPort() (string, int, error) {
	for _, host := range []string{"::1", "127.0.0.1"} {
		listener, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			continue
		}
		addr := listener.Addr().(*net.TCPAddr)
		_ = listener.Close()
		return host, addr.Port, nil
	}
	return "", 0, fmt.Errorf("reserve loopback TCP port: no loopback address available")
}

func (o *ConnectionObserver) Start(ctx context.Context) (<-chan error, error) {
	if o.Output == nil {
		o.Output = io.Discard
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(o.ListenHost, strconv.Itoa(o.ListenPort)))
	if err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()
	go func() {
		defer close(errCh)
		for {
			conn, err := listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				errCh <- err
				return
			}
			go o.handle(ctx, conn)
		}
	}()

	return errCh, nil
}

func (o *ConnectionObserver) handle(ctx context.Context, client net.Conn) {
	id := o.nextID.Add(1)
	started := time.Now()
	clientAddr := client.RemoteAddr().String()
	clientHost, clientPort := splitAddress(clientAddr)

	fmt.Fprintf(o.Output, "%sclient connected%s id=%d ip=%s port=%s remote=%s\n", ansiGreen, ansiReset, id, clientHost, clientPort, clientAddr)

	target, err := (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(o.TargetHost, strconv.Itoa(o.TargetPort)))
	if err != nil {
		_ = client.Close()
		fmt.Fprintf(o.Output, "%sclient disconnected%s id=%d ip=%s port=%s duration=%s error=%q\n", ansiRed, ansiReset, id, clientHost, clientPort, time.Since(started).Round(time.Millisecond), err.Error())
		return
	}

	var fromClient countingWriter
	var toClient countingWriter
	fromClient.w = target
	toClient.w = client
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(&fromClient, client)
		_ = target.Close()
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&toClient, target)
		_ = client.Close()
	}()

	wg.Wait()
	fmt.Fprintf(
		o.Output,
		"%sclient disconnected%s id=%d ip=%s port=%s remote=%s duration=%s bytes_from_client=%d bytes_to_client=%d\n",
		ansiRed,
		ansiReset,
		id,
		clientHost,
		clientPort,
		clientAddr,
		time.Since(started).Round(time.Millisecond),
		fromClient.Count(),
		toClient.Count(),
	)
}

func splitAddress(addr string) (string, string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, ""
	}
	return host, port
}

type countingWriter struct {
	n atomic.Int64
	w io.Writer
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.n.Add(int64(len(p)))
	if w.w == nil {
		return len(p), nil
	}
	return w.w.Write(p)
}

func (w *countingWriter) Count() int64 {
	return w.n.Load()
}
