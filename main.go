package main

import (
	"bilinovel-downloader/cmd"
	"log"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
