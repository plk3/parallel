package search

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// recursiveSearch handles recursive directory traversal with goroutine limits
func recursiveSearch(root string, filter func(string) bool, wg *sync.WaitGroup, resultChan chan<- string, errChan chan<- error, sem chan struct{}) {
	defer wg.Done()

	// Acquire a slot in the semaphore
	sem <- struct{}{}
	defer func() { <-sem }() // Release the slot when done

	entries, err := os.ReadDir(root)
	if err != nil {
		errChan <- err
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			wg.Add(1)
			go recursiveSearch(fullPath, filter, wg, resultChan, errChan, sem)
		} else if filter(fullPath) {
			resultChan <- fullPath
		}
	}
}

// SearchFiles searches files matching the filter in the given root directory
// with a limit on the maximum number of concurrent goroutines
func SearchFiles(root string, filter func(string) bool) ([]string, error) {
	resultChan := make(chan string, 100)
	errChan := make(chan error)
	var wg sync.WaitGroup

	// Semaphore to limit concurrent goroutines
	sem := make(chan struct{}, runtime.NumCPU()*2)

	// Start the recursive search
	wg.Add(1)
	go recursiveSearch(root, filter, &wg, resultChan, errChan, sem)

	// Close channels when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	var foundFiles []string
	var errs []error

	// Collect results and errors
	for resultChan != nil || errChan != nil {
		select {
		case result, ok := <-resultChan:
			if ok {
				foundFiles = append(foundFiles, result)
			} else {
				resultChan = nil
			}
		case err, ok := <-errChan:
			if ok {
				errs = append(errs, err)
			} else {
				errChan = nil
			}
		}
	}

	// Handle accumulated errors
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return foundFiles, nil
}
