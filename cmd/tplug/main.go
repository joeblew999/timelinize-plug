package main

import (
	"fmt"
	"os"

	"github.com/joeblew999/timelinize-plug/internal/version"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "tplug", Short: "Timelinize plug: CLI + Server with embedded NATS and PB"}
	root.AddCommand(cmdServe(), cmdAuth(), cmdImport(), cmdStatus(), cmdSync(), cmdSelfUpdate())
	addExtraCommands(root)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(version.String())
			return nil
		},
	}
}
