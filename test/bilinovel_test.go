package test

import (
	"bilinovel-downloader/downloader/bilinovel"
	"encoding/json"
	"fmt"
	"testing"
)

func TestBilinovel_GetNovel(t *testing.T) {
	bilinovel, err := bilinovel.New()
	bilinovel.SetTextOnly(true)
	bilinovel.SetDebug(true)
	if err != nil {
		t.Fatalf("failed to create bilinovel: %v", err)
	}
	novel, err := bilinovel.GetNovel(4519)
	if err != nil {
		t.Fatalf("failed to get novel: %v", err)
	}
	jsonBytes, err := json.Marshal(novel)
	if err != nil {
		t.Fatalf("failed to marshal novel: %v", err)
	}
	fmt.Println(string(jsonBytes))
}

func TestBilinovel_GetVolume(t *testing.T) {
	bilinovel, err := bilinovel.New()
	bilinovel.SetTextOnly(true)
	if err != nil {
		t.Fatalf("failed to create bilinovel: %v", err)
	}
	volume, err := bilinovel.GetVolume(1410, 52748)
	if err != nil {
		t.Fatalf("failed to get volume: %v", err)
	}
	jsonBytes, err := json.Marshal(volume)
	if err != nil {
		t.Fatalf("failed to marshal volume: %v", err)
	}
	fmt.Println(string(jsonBytes))
}

func TestBilinovel_GetChapter(t *testing.T) {
	bilinovel, err := bilinovel.New()
	if err != nil {
		t.Fatalf("failed to create bilinovel: %v", err)
	}
	bilinovel.SetDebug(true)
	chapter, err := bilinovel.GetChapter(1410, 52748, 52752)
	if err != nil {
		t.Fatalf("failed to get chapter: %v", err)
	}
	jsonBytes, err := json.Marshal(chapter)
	if err != nil {
		t.Fatalf("failed to marshal chapter: %v", err)
	}
	fmt.Println(string(jsonBytes))
}
