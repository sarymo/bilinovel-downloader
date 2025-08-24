package cmd

import (
	"bilinovel-downloader/epub"
	"fmt"

	"github.com/spf13/cobra"
)

type packArgs struct {
	DirPath string `validate:"required"`
}

var (
	pArgs packArgs
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "pack a epub file from directory",
	Long:  "pack a epub file from directory",
	RunE:  runPackage,
}

func init() {
	packCmd.Flags().StringVarP(&pArgs.DirPath, "dir-path", "d", "", "directory path")
	RootCmd.AddCommand(packCmd)
}

func runPackage(cmd *cobra.Command, args []string) error {
	err := epub.PackEpub(pArgs.DirPath)
	if err != nil {
		return fmt.Errorf("failed to create epub: %v", err)
	}
	return nil
}
