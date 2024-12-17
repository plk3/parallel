package main

import (
	"os"
	"path/filepath"
	"sync"
)

func SearchFiles(root string, filter func(string) bool) ([]string, error) {
	var results []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	resultChan := make(chan string)

	// Worker function to process directories in parallel
	var worker func(path string, wg *sync.WaitGroup)
	worker = func(path string, wg *sync.WaitGroup) {
		defer wg.Done() // Ensure this goroutine is done when it finishes

		entries, err := os.ReadDir(path)
		if err != nil {
			errChan <- err
			return
		}

		for _, entry := range entries {
			fullPath := filepath.Join(path, entry.Name())
			if filter(fullPath) {
				resultChan <- fullPath
			}
			if entry.IsDir() {
				// Add a new goroutine for subdirectories before calling it
				wg.Add(1)
				go worker(fullPath, wg)
			}
		}
	}

	// Start the search with the root directory
	wg.Add(1)
	go worker(root, &wg)

	// Collect results and handle errors
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// Process results and errors
	for {
		select {
		case err := <-errChan:
			return nil, err
		case result, ok := <-resultChan:
			if !ok {
				return results, nil
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}
	}
}
