package bilinovel

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
)

func CreateEpub(path string) error {
	log.Printf("Creating epub for %s", path)

	savePath := path + ".epub"
	zipFile, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = addStringToZip(zipWriter, "mimetype", "application/epub+zip", zip.Store)
	if err != nil {
		return err
	}

	err = addDirContentToZip(zipWriter, path, zip.Deflate)
	if err != nil {
		return err
	}

	err = addStringToZip(zipWriter, "style.css", StyleCSS, zip.Deflate)
	if err != nil {
		return err
	}

	return nil
}

// func addFileToZip(zipWriter *zip.Writer, filename string, relPath string, method uint16) error {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	info, err := file.Stat()
// 	if err != nil {
// 		return err
// 	}

// 	header, err := zip.FileInfoHeader(info)
// 	if err != nil {
// 		return err
// 	}
// 	header.Name = relPath
// 	header.Method = method

// 	writer, err := zipWriter.CreateHeader(header)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = io.Copy(writer, file)
// 	return err
// }

func addStringToZip(zipWriter *zip.Writer, relPath, content string, method uint16) error {
	header := &zip.FileHeader{
		Name:   relPath,
		Method: method,
	}
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(content))
	return err
}

func addDirContentToZip(zipWriter *zip.Writer, dirPath string, method uint16) error {
	return filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if filepath.Base(filePath) == "volume.json" {
			return nil
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			return err
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = method

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})
}
