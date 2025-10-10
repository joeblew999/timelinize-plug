package main

import (
	"context"
	"log"

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

			// 1) NATS
			ns, js, err := nats.StartEmbedded(ctx, nats.Options{
				Memory:  memory,
				Offline: offline,
			})
			if err != nil { return err }
			defer ns.Shutdown()

			// 2) PocketBase (NO CGO)
			app, err := pb.StartEmbedded(pb.Options{})
			if err != nil {
				return err
			}
			log.Printf("PocketBase running")

			// 3) HTTP server
			return server.Start(ctx, server.Options{
				Addr:        "127.0.0.1:12002",
				OAuthCBAddr: "127.0.0.1:8008",
				JSCtx:       js,
				NATSConn:    ns.ClientConn(),
				// TODO: pass PB app once server needs it
			})
		},
	}
	c.Flags().BoolVar(&memory, "memory", false, "use in-memory JetStream (tests)")
	c.Flags().BoolVar(&offline, "offline", false, "disable upstream NATS bridge")
	return c
}
