package main

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

func initCreateFile(root string, files []string) error {
	for _, file := range files {
		path := filepath.Join(root, file)
		if strings.Contains(file, "/") {
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

func TestSearchFiles(t *testing.T) {
	// 一時ディレクトリを作成 (テスト終了時に自動で削除される)
	tempDir := t.TempDir()
	files := []string{
		"test1.txt",
		"test2.go",
		"subdir/test3.go",
		"subdir/test4.txt",
		"subdir/nested/test5.go",
	}
	initCreateFile(tempDir, files)

	// テスト対象の関数を呼び出し
	foundFiles, err := SearchFiles(tempDir, func(path string) bool {
		return strings.HasSuffix(path, ".go")
	})
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	// 期待される結果
	expected := []string{
		filepath.Join(tempDir, "test2.go"),
		filepath.Join(tempDir, "subdir/test3.go"),
		filepath.Join(tempDir, "subdir/nested/test5.go"),
	}

	// 比較
	sort.Strings(foundFiles)
	sort.Strings(expected)
	if !slices.Equal(foundFiles, expected) {
		t.Errorf("Expected files %v, but got %v", expected, foundFiles)
	}
}
