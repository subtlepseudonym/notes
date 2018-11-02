package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultNotesDirectory = "/Users/cdemille/workspace/log"
	defaultModTimeFormat  = time.RFC822
)

func Execute() {
	root := buildCommands()
	root.Execute()
}

func buildCommands() *cobra.Command {
	root := root()

	root.AddCommand(ls())

	return root
}

func root() *cobra.Command {
	var root = &cobra.Command{
		Use: "nts",
	}
	return root
}
