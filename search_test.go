package search

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
	t.Run("normal case", func(t *testing.T) {
		// 一時ディレクトリを作成 (テスト終了時に自動で削除される)
		tempDir := t.TempDir()
		files := []string{
			"test1.txt",
			"keyword.go",
			"subdir/keyword.go",
			"subdir/test4.txt",
			"subdir/keyword/test5.go",
		}
		initCreateFile(tempDir, files)

		// テスト対象の関数を呼び出し
		foundFiles, err := SearchFiles(tempDir, func(path string) bool {
			return strings.Contains(path, "keyword")
		})
		if err != nil {
			t.Fatalf("SearchFiles failed: %v", err)
		}

		// 期待される結果
		expected := []string{
			filepath.Join(tempDir, "keyword.go"),
			filepath.Join(tempDir, "subdir/keyword.go"),
			filepath.Join(tempDir, "subdir/keyword/test5.go"),
		}

		// 比較
		sort.Strings(foundFiles)
		sort.Strings(expected)
		if !slices.Equal(foundFiles, expected) {
			t.Errorf("Expected files %v, but got %v", expected, foundFiles)
		}
	})

	t.Run("invalid directory", func(t *testing.T) {
		// 存在しないディレクトリを指定
		_, err := SearchFiles("/invalid/path", func(path string) bool {
			return strings.Contains(path, "keyword")
		})
		if err == nil {
			t.Fatalf("Expected error, but got nil")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		// 空のディレクトリ
		tempDir := t.TempDir()

		// テスト対象の関数を呼び出し
		foundFiles, err := SearchFiles(tempDir, func(path string) bool {
			return strings.Contains(path, "keyword")
		})
		if err != nil {
			t.Fatalf("SearchFiles failed: %v", err)
		}

		// 結果が空であることを確認
		if len(foundFiles) != 0 {
			t.Errorf("Expected no files, but got %v", foundFiles)
		}
	})
}
