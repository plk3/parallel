package main

import (
	"os"
	"path/filepath"
	"sync"
)

func recursiveSearch(root string, filter func(string) bool, wg *sync.WaitGroup, results *[]string, errChan chan error) {
	defer wg.Done() // Ensure this goroutine is done when it finishes
	entries, err := os.ReadDir(root)
	if err != nil {
		errChan <- err // エラーが発生した場合、エラーをチャネルに送信
		return
	}

	var mu sync.Mutex
	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			wg.Add(1)
			go recursiveSearch(fullPath, filter, wg, results, errChan)
		} else if filter(fullPath) {
			mu.Lock()
			*results = append(*results, fullPath)
			mu.Unlock()
		}
	}
}

func SearchFiles(root string, filter func(string) bool) ([]string, error) {
	var results []string
	var wg sync.WaitGroup
	errChan := make(chan error, 1) // エラーを受け取るチャネルを作成
	wg.Add(1)
	go recursiveSearch(root, filter, &wg, &results, errChan)

	wg.Wait()      // 全ての goroutine が完了するのを待つ
	close(errChan) // チャネルを閉じる

	// チャネルからエラーを受け取って返す
	if err := <-errChan; err != nil {
		return nil, err
	}

	return results, nil
}
