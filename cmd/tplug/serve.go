package main

import (
	"context"
	"log"
	"os"

	"github.com/joeblew999/timelinize-plug/internal/nats"
	"github.com/joeblew999/timelinize-plug/internal/pb"
	"github.com/joeblew999/timelinize-plug/internal/server"
	"github.com/spf13/cobra"
)

func cmdServe() *cobra.Command {
	var memory bool
	var offline bool

	c := &cobra.Command{
		Use:   "serve",
		Short: "Start server (embedded NATS + PB + Chi UI + SSE)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			tenant := os.Getenv("TPLUG_TENANT_ID")
			if tenant == "" {
				tenant = "local"
			}

			// 1) NATS
			ns, js, err := nats.StartEmbedded(ctx, nats.Options{
				Memory:  memory,
				Offline: offline,
			})
			if err != nil {
				return err
			}
			defer ns.Shutdown()

			// 2) PocketBase (NO CGO)
			app, err := pb.StartEmbedded(pb.Options{})
			if err != nil {
				return err
			}
			log.Printf("PocketBase running")

			bridge := pb.NewRealtimeBridge(app, ns.ClientConn(), js, tenant)
			bridge.Start()

			// 3) HTTP server
			opts := server.Options{
				Addr:        "127.0.0.1:12002",
				OAuthCBAddr: "127.0.0.1:8008",
				JSCtx:       js,
				NATSConn:    ns.ClientConn(),
				PBApp:       app,
				Tenant:      tenant,
			}

			log.Printf("UI available at http://%s/", opts.Addr)
			log.Printf("Datastar panel: http://%s/ui/datastar/status", opts.Addr)
			if ns.InMemory() {
				log.Printf("NATS JetStream storage: in-memory (no persistence)")
			} else if jsDir := ns.JetStreamDir(); jsDir != "" {
				log.Printf("NATS JetStream storage dir: %s", jsDir)
			}
			if storeDir := ns.StoreDir(); storeDir != "" {
				log.Printf("NATS base store dir: %s", storeDir)
			}
			if logFile := ns.LogFile(); logFile != "" {
				log.Printf("NATS log file: %s", logFile)
			}
			log.Printf("PocketBase data dir: %s", app.DataDir())

			return server.Start(ctx, opts)
		},
	}
	c.Flags().BoolVar(&memory, "memory", false, "use in-memory JetStream (tests)")
	c.Flags().BoolVar(&offline, "offline", false, "disable upstream NATS bridge")
	return c
}
