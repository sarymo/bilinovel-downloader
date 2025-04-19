package cmd

import (
	"bilinovel-downloader/downloader"
	"fmt"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a novel or volume",
	Long:  "Download a novel or volume",
}

var downloadNovelCmd = &cobra.Command{
	Use:   "novel",
	Short: "Download a novel, default download all volumes",
	Long:  "Download a novel, default download all volumes",
	RunE:  runDownloadNovel,
}

var downloadVolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Download a volume",
	Long:  "Download a volume",
	RunE:  runDownloadVolume,
}

type downloadNovelArgs struct {
	NovelId    int `validate:"required"`
	outputPath string
}

type downloadVolumeArgs struct {
	NovelId    int `validate:"required"`
	VolumeId   int `validate:"required"`
	outputPath string
}

var (
	novelArgs  downloadNovelArgs
	volumeArgs downloadVolumeArgs
)

func init() {
	downloadNovelCmd.Flags().IntVarP(&novelArgs.NovelId, "novel-id", "n", 0, "novel id")
	downloadNovelCmd.Flags().StringVarP(&novelArgs.outputPath, "output-path", "o", "./novels", "output path")

	downloadVolumeCmd.Flags().IntVarP(&volumeArgs.NovelId, "novel-id", "n", 0, "novel id")
	downloadVolumeCmd.Flags().IntVarP(&volumeArgs.VolumeId, "volume-id", "v", 0, "volume id")
	downloadVolumeCmd.Flags().StringVarP(&volumeArgs.outputPath, "output-path", "o", "./novels", "output path")

	downloadCmd.AddCommand(downloadNovelCmd)
	downloadCmd.AddCommand(downloadVolumeCmd)
	RootCmd.AddCommand(downloadCmd)
}

func runDownloadNovel(cmd *cobra.Command, args []string) error {
	if novelArgs.NovelId == 0 {
		return fmt.Errorf("novel id is required")
	}
	err := downloader.DownloadNovel(novelArgs.NovelId, novelArgs.outputPath)
	if err != nil {
		return fmt.Errorf("failed to download novel: %v", err)
	}

	return nil
}

func runDownloadVolume(cmd *cobra.Command, args []string) error {
	if volumeArgs.NovelId == 0 {
		return fmt.Errorf("novel id is required")
	}
	if volumeArgs.VolumeId == 0 {
		return fmt.Errorf("volume id is required")
	}
	err := downloader.DownloadVolume(volumeArgs.NovelId, volumeArgs.VolumeId, volumeArgs.outputPath)
	if err != nil {
		return fmt.Errorf("failed to download volume: %v", err)
	}

	return nil
}
