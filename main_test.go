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

func TestParseBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		hasError bool
	}{
		{name: "lowercase true", input: "true", expected: true, hasError: false},
		{name: "titlecase True", input: "True", expected: true, hasError: false},
		{name: "uppercase TRUE", input: "TRUE", expected: true, hasError: false},
		{name: "char t", input: "t", expected: true, hasError: false},
		{name: "char T", input: "T", expected: true, hasError: false},
		{name: "number 1", input: "1", expected: true, hasError: false},
		{name: "lowercase false", input: "false", expected: false, hasError: false},
		{name: "titlecase False", input: "False", expected: false, hasError: false},
		{name: "uppercase FALSE", input: "FALSE", expected: false, hasError: false},
		{name: "char f", input: "f", expected: false, hasError: false},
		{name: "char F", input: "F", expected: false, hasError: false},
		{name: "number 0", input: "0", expected: false, hasError: false},
		{name: "invalid string", input: "invalid", expected: false, hasError: true},
		{name: "empty string", input: "", expected: false, hasError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBool(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("ParseBool() with input %q expected an error, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseBool() with input %q returned an unexpected error: %v", tt.input, err)
				}
				if got != tt.expected {
					t.Errorf("ParseBool() with input %q got = %v, want %v", tt.input, got, tt.expected)
				}
			}
		})
	}
}
