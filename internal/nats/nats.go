package nats

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	natserver "github.com/nats-io/nats-server/v2/server"
	natsgo "github.com/nats-io/nats.go"
)

// Options configures the embedded NATS instance.
type Options struct {
	Memory   bool   // when true, JetStream stays in-memory (no StoreDir persisted)
	Offline  bool   // future: disable upstream leaf connections (not yet implemented)
	StoreDir string // optional override; defaults to ~/.tplug/nats
}

// Embedded wraps the running NATS server and an admin client connection.
type Embedded struct {
	srv      *natserver.Server
	conn     *natsgo.Conn
	storeDir string
	jsDir    string
	memory   bool
	logFile  string
}

// ClientConn exposes the management connection so other subsystems can reuse it.
func (e *Embedded) ClientConn() *natsgo.Conn {
	if e == nil {
		return nil
	}
	return e.conn
}

// StoreDir returns the base store directory used by the embedded server.
func (e *Embedded) StoreDir() string {
	if e == nil {
		return ""
	}
	return e.storeDir
}

// JetStreamDir returns the JetStream storage directory (empty when in-memory).
func (e *Embedded) JetStreamDir() string {
	if e == nil {
		return ""
	}
	return e.jsDir
}

// InMemory reports whether JetStream is running without disk persistence.
func (e *Embedded) InMemory() bool {
	if e == nil {
		return false
	}
	return e.memory
}

// LogFile returns the path to the server log file, if configured.
func (e *Embedded) LogFile() string {
	if e == nil {
		return ""
	}
	return e.logFile
}

// Shutdown drains the client connection and stops the embedded server.
func (e *Embedded) Shutdown() {
	if e == nil {
		return
	}
	if e.conn != nil {
		_ = e.conn.Drain()
		e.conn.Close()
	}
	if e.srv != nil {
		e.srv.Shutdown()
		e.srv.WaitForShutdown()
	}
}

// StartEmbedded spins up an in-process NATS with JetStream enabled and returns both
// the server handle and a JetStream context ready for stream/subject provisioning.
func StartEmbedded(ctx context.Context, opt Options) (*Embedded, natsgo.JetStreamContext, error) {
	storeDir := opt.StoreDir
	if storeDir == "" {
		storeDir = filepath.Join(".", ".data", "nats")
	}
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return nil, nil, err
	}
	jsStore := ""
	if !opt.Memory {
		jsStore = filepath.Join(storeDir, "jetstream")
		if err := os.MkdirAll(jsStore, 0o755); err != nil {
			return nil, nil, err
		}
	}
	logDir := filepath.Join(storeDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, nil, err
	}
	logFile := filepath.Join(logDir, "nats.log")

	sopts := &natserver.Options{
		Host:      "127.0.0.1",
		Port:      4222,
		HTTPPort:  -1,
		JetStream: true,
		StoreDir:  jsStore,
		LogFile:   logFile,
	}
	// TODO: wire Offline to control leaf nodes/upstream bridging once feature is implemented.

	addr := net.JoinHostPort(sopts.Host, fmt.Sprintf("%d", sopts.Port))
	if l, err := net.Listen("tcp", addr); err != nil {
		return nil, nil, fmt.Errorf("nats: client port %s unavailable: %w", addr, err)
	} else {
		_ = l.Close()
	}

	srv, err := natserver.NewServer(sopts)
	if err != nil {
		return nil, nil, err
	}

	ready := make(chan struct{})
	go func() {
		srv.ConfigureLogger()
		srv.Start()
		close(ready)
	}()

	select {
	case <-ctx.Done():
		srv.Shutdown()
		return nil, nil, ctx.Err()
	case <-ready:
	}

	if !srv.ReadyForConnections(10 * time.Second) {
		srv.Shutdown()
		return nil, nil, errors.New("nats: server not ready within timeout")
	}

	conn, err := natsgo.Connect(srv.ClientURL(),
		natsgo.RetryOnFailedConnect(true),
		natsgo.MaxReconnects(-1),
		natsgo.ReconnectWait(250*time.Millisecond),
	)
	if err != nil {
		srv.Shutdown()
		return nil, nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		srv.Shutdown()
		return nil, nil, err
	}

	log.Printf("Embedded NATS listening at %s", srv.ClientURL())
	if sopts.HTTPPort >= 0 {
		log.Printf("Embedded NATS monitoring at http://127.0.0.1:%d", sopts.HTTPPort)
	} else {
		log.Printf("Embedded NATS monitoring endpoint disabled")
	}
	if opt.Memory {
		log.Printf("Embedded NATS JetStream storage: in-memory (no persistence)")
	} else {
		log.Printf("Embedded NATS JetStream store dir: %s", jsStore)
	}
	log.Printf("Embedded NATS log file: %s", logFile)

	return &Embedded{srv: srv, conn: conn, storeDir: storeDir, jsDir: jsStore, memory: opt.Memory, logFile: logFile}, js, nil
}
