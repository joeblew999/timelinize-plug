package main

import "github.com/spf13/cobra"

func init() {}

func addExtraCommands(root *cobra.Command) {
	root.AddCommand(cmdConfig())
}
