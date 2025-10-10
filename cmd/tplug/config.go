package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joeblew999/timelinize-plug/internal/config"
	"github.com/joeblew999/timelinize-plug/internal/pb"
	"github.com/spf13/cobra"
)

func cmdConfig() *cobra.Command {
	var key, value string
	c := &cobra.Command{
		Use:   "config",
		Short: "Set/get PB KV config values",
	}
	set := &cobra.Command{
		Use:   "set",
		Short: "Set a KV entry (json)",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := pb.StartEmbedded(pb.Options{})
			if err != nil { return err }
			_ = config.EnsureKV(app)
			var raw json.RawMessage
			if value == "-" {
				data, _ := os.ReadFile("/dev/stdin")
				raw = data
			} else { raw = json.RawMessage(value) }
			return config.Set(app, key, raw)
		},
	}
	set.Flags().StringVar(&key, "key", "", "key")
	set.Flags().StringVar(&value, "value", "", "json value or '-' for stdin")
	_ = set.MarkFlagRequired("key")

	get := &cobra.Command{
		Use:   "get",
		Short: "Get a KV entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := pb.StartEmbedded(pb.Options{})
			if err != nil { return err }
			_ = config.EnsureKV(app)
			b, err := config.Get(app, key)
			if err != nil { return err }
			fmt.Println(string(b))
			return nil
		},
	}
	get.Flags().StringVar(&key, "key", "", "key")
	_ = get.MarkFlagRequired("key")

	c.AddCommand(set, get)
	return c
}
