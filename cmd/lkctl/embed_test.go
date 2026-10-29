package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackage(t *testing.T) {
	dir := t.TempDir()

	// 空目录：返回空
	if got := detectPackage(dir, "key.go"); got != "" {
		t.Fatalf("empty dir: got %q, want empty", got)
	}

	// 写一个 main 包文件
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := detectPackage(dir, "key.go"); got != "main" {
		t.Fatalf("got %q, want main", got)
	}

	// exclude 生效：排除 main.go 后目录里只剩另一个包的文件
	other := filepath.Join(dir, "other")
	if err := os.Mkdir(other, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "x.go"), []byte("package license\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := detectPackage(dir, "main.go"); got != "" {
		t.Fatalf("subdir should be ignored, got %q", got)
	}

	// 自身文件被排除（重新生成场景）：旧 key.go 的包名不影响检测
	if err := os.WriteFile(filepath.Join(dir, "key.go"), []byte("package license\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := detectPackage(dir, "key.go"); got != "main" {
		t.Fatalf("regenerate: got %q, want main", got)
	}
}
