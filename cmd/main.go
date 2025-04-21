package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wovenet",
	Short: "Wovenet CLI",
	Long: `Wovenet is a tool to establish transport-layer tunnels for applications
across isolated private networks. This make you can access a remote
service which local in a private network as if it is on your local
network.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
