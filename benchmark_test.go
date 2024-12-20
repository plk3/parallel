package search

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func CreateTestDirectory(root string, depth int, filesPerDir int) error {
	// CreateTestDirectory generates a test directory structure with the given depth and file count per directory.func CreateTestDirectory(root string, depth int, filesPerDir int) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errList []error
	var generate func(root string, depth int)

	// Semaphore to limit simultaneous goroutines
	sem := make(chan struct{}, 100) // Limit to 100 concurrent goroutines

	// Start generating directories and files
	generate = func(root string, depth int) {
		defer wg.Done()

		// Base case: Stop if depth is zero
		if depth == 0 {
			return
		}

		// Generate files in the current directory
		for i := 0; i < filesPerDir; i++ {
			wg.Add(1)
			sem <- struct{}{} // Acquire a semaphore slot
			go func(fileName string) {
				defer wg.Done()
				defer func() { <-sem }() // Release the semaphore slot
				if err := os.WriteFile(fileName, []byte("test"), 0644); err != nil {
					mu.Lock()
					errList = append(errList, err)
					mu.Unlock()
				}
			}(filepath.Join(root, randomFileName()))
		}

		// Generate subdirectories recursively
		for i := 0; i < filesPerDir/2; i++ {
			subDir := filepath.Join(root, randomFileName())
			if err := os.Mkdir(subDir, 0755); err != nil {
				mu.Lock()
				errList = append(errList, err)
				mu.Unlock()
				continue
			}
			wg.Add(1)
			go generate(subDir, depth-1)
		}
	}

	// Start the root directory generation
	wg.Add(1)
	go generate(root, depth)

	// Wait for all goroutines to finish
	wg.Wait()

	// Aggregate errors
	if len(errList) > 0 {
		return fmt.Errorf("errors occurred: %v", errList)
	}
	return nil
}

// randomFileName generates a random file or directory name
func randomFileName() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Test setup
func setupTestEnvironment() (string, error) {
	tempDir, err := os.MkdirTemp("", "benchmark_test")
	if err != nil {
		return "", err
	}

	if err := CreateTestDirectory(tempDir, 3, 10); err != nil { // Depth: 3, 10 files per directory
		return "", err
	}

	return tempDir, nil
}

func SingleThreadSearchFiles(root string, filter func(string) bool) ([]string, error) {
	var foundFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filter(path) {
			foundFiles = append(foundFiles, path)
		}
		return nil
	})
	return foundFiles, err
}

// BenchmarkSingleThreadSearchFiles benchmarks filepath.Walk-based file search
func BenchmarkSingleThreadSearchFiles(b *testing.B) {
	tempDir, err := setupTestEnvironment()
	if err != nil {
		b.Fatalf("Failed to set up test environment: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filter := func(path string) bool {
		return strings.Contains(path, "abc")
	}

	b.ResetTimer() // Reset timer to exclude setup time
	for i := 0; i < b.N; i++ {
		_, err := SingleThreadSearchFiles(tempDir, filter)
		if err != nil {
			b.Fatalf("SingleThreadSearchFiles failed: %v", err)
		}
	}
}

// BenchmarkParallelSearchFiles benchmarks the parallel file search
func BenchmarkParallelSearchFiles(b *testing.B) {
	tempDir, err := setupTestEnvironment()
	if err != nil {
		b.Fatalf("Failed to set up test environment: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filter := func(path string) bool {
		return strings.Contains(path, "abc")
	}

	b.ResetTimer() // Reset timer to exclude setup time
	for i := 0; i < b.N; i++ {
		_, err := SearchFiles(tempDir, filter)
		if err != nil {
			b.Fatalf("SearchFiles failed: %v", err)
		}
	}
}
