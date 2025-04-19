package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version = "dev"
)

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version: ", Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
