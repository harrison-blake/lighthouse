package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateOutputDirs(t *testing.T) {
	tempDir := t.TempDir()
	testDirs := []string{
		"one",
		"two",
		"three",
	}
	expectedVal := uint32(len(testDirs))

	gotVal, err := CreateOutputDirs(tempDir, testDirs)
	if err != nil {
		t.Fatalf("CreateOutputDirs failed: %v", err)
	}

	if gotVal != expectedVal {
		t.Errorf("expected to create %v directories, but got %v", expectedVal, gotVal)
	}

	for _, dir := range testDirs {
		dirPath := filepath.Join(tempDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("expected directory %v to be created, but it was not", dirPath)
		}
	}
}