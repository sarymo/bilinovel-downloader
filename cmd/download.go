package cmd

import (
	"bilinovel-downloader/downloader/bilinovel"
	"bilinovel-downloader/epub"
	"bilinovel-downloader/model"
	"bilinovel-downloader/text"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a novel or volume",
	Long:  "Download a novel or volume",
	Run: func(cmd *cobra.Command, args []string) {
		err := runDownloadNovel()
		if err != nil {
			log.Printf("failed to download novel: %v", err)
		}
	},
}

type downloadCmdArgs struct {
	NovelId    int `validate:"required"`
	VolumeId   int `validate:"required"`
	outputPath string
	outputType string
}

var (
	downloadArgs downloadCmdArgs
)

func init() {
	downloadCmd.Flags().IntVarP(&downloadArgs.NovelId, "novel-id", "n", 0, "novel id")
	downloadCmd.Flags().IntVarP(&downloadArgs.VolumeId, "volume-id", "v", 0, "volume id")
	downloadCmd.Flags().StringVarP(&downloadArgs.outputPath, "output-path", "o", "novels", "output path")
	downloadCmd.Flags().StringVarP(&downloadArgs.outputType, "output-type", "t", "epub", "output type, epub or text")
	RootCmd.AddCommand(downloadCmd)
}

func runDownloadNovel() error {
	downloader, err := bilinovel.New()
	if err != nil {
		return fmt.Errorf("failed to create downloader: %v", err)
	}
	if downloadArgs.NovelId == 0 {
		return fmt.Errorf("novel id is required")
	}
	if downloadArgs.VolumeId == 0 {
		novel, err := downloadNovel(downloader)
		if err != nil {
			return fmt.Errorf("failed to get novel: %v", err)
		}
		switch downloadArgs.outputType {
		case "epub":
			for _, volume := range novel.Volumes {
				err = epub.PackVolumeToEpub(volume, downloadArgs.outputPath, downloader.GetStyleCSS(), downloader.GetExtraFiles())
				if err != nil {
					return fmt.Errorf("failed to pack volume: %v", err)
				}
			}
		case "text":
			for _, volume := range novel.Volumes {
				err = text.PackVolumeToText(volume, downloadArgs.outputPath)
				if err != nil {
					return fmt.Errorf("failed to pack volume: %v", err)
				}
			}
		}
	} else {
		// 下载单卷
		volume, err := downloadVolume(downloader)
		if err != nil {
			return fmt.Errorf("failed to get volume: %v", err)
		}
		switch downloadArgs.outputType {
		case "epub":
			err = epub.PackVolumeToEpub(volume, downloadArgs.outputPath, downloader.GetStyleCSS(), downloader.GetExtraFiles())
			if err != nil {
				return fmt.Errorf("failed to pack volume: %v", err)
			}
		case "text":
			err = text.PackVolumeToText(volume, downloadArgs.outputPath)
			if err != nil {
				return fmt.Errorf("failed to pack volume: %v", err)
			}
		}
	}

	return nil
}

func downloadNovel(downloader model.Downloader) (*model.Novel, error) {
	jsonPath := filepath.Join(downloadArgs.outputPath, fmt.Sprintf("novel-%d.json", downloadArgs.NovelId))
	err := os.MkdirAll(filepath.Dir(jsonPath), 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}
	_, err = os.Stat(jsonPath)
	novel := &model.Novel{}
	if err != nil {
		if os.IsNotExist(err) {
			novel, err = downloader.GetNovel(downloadArgs.NovelId)
			if err != nil {
				return nil, fmt.Errorf("failed to get novel: %v", err)
			}
			jsonFile, err := os.Create(jsonPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create json file: %v", err)
			}
			defer jsonFile.Close()
			err = json.NewEncoder(jsonFile).Encode(novel)
			if err != nil {
				return nil, fmt.Errorf("failed to encode json file: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get novel: %v", err)
		}
	} else {
		jsonFile, err := os.Open(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open json file: %v", err)
		}
		defer jsonFile.Close()
		err = json.NewDecoder(jsonFile).Decode(novel)
		if err != nil {
			return nil, fmt.Errorf("failed to decode json file: %v", err)
		}
	}
	return novel, nil
}

func downloadVolume(downloader model.Downloader) (*model.Volume, error) {
	jsonPath := filepath.Join(downloadArgs.outputPath, fmt.Sprintf("volume-%d-%d.json", downloadArgs.NovelId, downloadArgs.VolumeId))
	err := os.MkdirAll(filepath.Dir(jsonPath), 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}
	_, err = os.Stat(jsonPath)
	volume := &model.Volume{}
	if err != nil {
		if os.IsNotExist(err) {
			volume, err = downloader.GetVolume(downloadArgs.NovelId, downloadArgs.VolumeId)
			if err != nil {
				return nil, fmt.Errorf("failed to get volume: %v", err)
			}
			jsonFile, err := os.Create(jsonPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create json file: %v", err)
			}
			err = json.NewEncoder(jsonFile).Encode(volume)
			if err != nil {
				return nil, fmt.Errorf("failed to encode json file: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get volume: %v", err)
		}
	} else {
		jsonFile, err := os.Open(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open json file: %v", err)
		}
		defer jsonFile.Close()
		err = json.NewDecoder(jsonFile).Decode(volume)
		if err != nil {
			return nil, fmt.Errorf("failed to decode json file: %v", err)
		}
	}
	return volume, nil
}
