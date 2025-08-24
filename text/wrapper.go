package text

import (
	"bilinovel-downloader/model"
	"bilinovel-downloader/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func PackVolumeToText(volume *model.Volume, outputPath string) error {
	outputPath = filepath.Join(outputPath, utils.CleanDirName(volume.Title))
	_, err := os.Stat(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(outputPath, 0755)
			if err != nil {
				return fmt.Errorf("failed to create output directory: %v", err)
			}
		} else {
			return fmt.Errorf("failed to get output directory: %v", err)
		}
	} else {
		err = os.RemoveAll(outputPath)
		if err != nil {
			return fmt.Errorf("failed to remove output directory: %v", err)
		}
		err = os.MkdirAll(outputPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}
	for i, chapter := range volume.Chapters {
		chapterPath := filepath.Join(outputPath, fmt.Sprintf("%03d-%s.txt", i, chapter.Title))
		chapterFile, err := os.Create(chapterPath)
		if err != nil {
			return fmt.Errorf("failed to create chapter file: %v", err)
		}
		defer chapterFile.Close()
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(chapter.Content.Html))
		if err != nil {
			return fmt.Errorf("failed to create chapter file: %v", err)
		}
		doc.Find("img").Remove()
		text := doc.Text()
		_, err = chapterFile.WriteString(text)
		if err != nil {
			return fmt.Errorf("failed to write chapter file: %v", err)
		}
	}
	return nil
}
