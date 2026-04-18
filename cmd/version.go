package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/TejasGhatte/go-sail/cmd.Version=v1.0.0"
var Version = "dev"

var VersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Print the version of go-sail",
	Long:  "Displays the current version of the go-sail CLI tool.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-sail %s\n", Version)
	},
}
