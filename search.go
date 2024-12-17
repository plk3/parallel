package main

import (
	"os"
	"path/filepath"
	"sync"
)

func recursiveSearch(root string, maxDepth int, filter func(string) bool, wg *sync.WaitGroup, resultChan chan string, errChan chan error) {
	defer wg.Done()
	// -2以下で渡せば、最大までループ
	if maxDepth == -1 {
		return
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		errChan <- err
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			wg.Add(1)
			go recursiveSearch(fullPath, maxDepth-1, filter, wg, resultChan, errChan)
		} else if filter(fullPath) {
			resultChan <- fullPath
		}
	}
}

func SearchFiles(root string, filter func(string) bool) ([]string, error) {
	return SearchFilesWithDepth(root, -2, filter)
}

// depth 0でrootだけ検索
func SearchFilesWithDepth(root string, maxDepth int, filter func(string) bool) ([]string, error) {
	results := make(chan string)
	errors := make(chan error)

	var wg sync.WaitGroup
	wg.Add(1)
	go recursiveSearch(root, maxDepth, filter, &wg, results, errors)

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	var foundFiles []string
	for {
		select {
		case result, ok := <-results:
			if !ok {
				results = nil
			} else {
				foundFiles = append(foundFiles, result)
			}
		case err, ok := <-errors:
			if !ok {
				errors = nil
			} else {
				return nil, err
			}
		}
		if results == nil && errors == nil {
			break
		}
	}

	return foundFiles, nil
}
