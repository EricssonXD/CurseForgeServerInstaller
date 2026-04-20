package fs

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackRoot_Direct(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "mods"), 0o755)
	got := DetectPackRoot(tmp)
	if got != tmp {
		t.Errorf("DetectPackRoot() = %q, want %q", got, tmp)
	}
}

func TestDetectPackRoot_Nested(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "server", "mods"), 0o755)
	got := DetectPackRoot(tmp)
	if got != filepath.Join(tmp, "server") {
		t.Errorf("DetectPackRoot() = %q", got)
	}
}

func TestDetectPackRoot_NoMods(t *testing.T) {
	tmp := t.TempDir()
	got := DetectPackRoot(tmp)
	if got != tmp {
		t.Errorf("expected fallback to root, got %q", got)
	}
}

func TestExtractZip(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "test.zip")
	destDir := filepath.Join(tmp, "out")

	// Create test zip
	f, _ := os.Create(zipPath)
	w := zip.NewWriter(f)
	fw, _ := w.Create("hello.txt")
	fw.Write([]byte("world"))
	fw, _ = w.Create("sub/nested.txt")
	fw.Write([]byte("nested"))
	w.Close()
	f.Close()

	if err := ExtractZip(zipPath, destDir); err != nil {
		t.Fatalf("ExtractZip: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil || string(data) != "world" {
		t.Errorf("hello.txt: %q, %v", data, err)
	}
	data, err = os.ReadFile(filepath.Join(destDir, "sub", "nested.txt"))
	if err != nil || string(data) != "nested" {
		t.Errorf("sub/nested.txt: %q, %v", data, err)
	}
}

func TestExtractZip_ZipSlip(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "evil.zip")
	destDir := filepath.Join(tmp, "out")

	f, _ := os.Create(zipPath)
	w := zip.NewWriter(f)
	fw, _ := w.Create("../../../etc/evil.txt")
	fw.Write([]byte("pwned"))
	w.Close()
	f.Close()

	err := ExtractZip(zipPath, destDir)
	if err == nil {
		t.Fatal("expected zip slip error")
	}
}

func TestCopyTreeContents(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0o644)
	os.MkdirAll(filepath.Join(src, "sub"), 0o755)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("bbb"), 0o644)

	if err := CopyTreeContents(src, dst); err != nil {
		t.Fatalf("CopyTreeContents: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(data) != "aaa" {
		t.Errorf("a.txt = %q", data)
	}
	data, _ = os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if string(data) != "bbb" {
		t.Errorf("sub/b.txt = %q", data)
	}
}

func TestUpdateFromPackRoot(t *testing.T) {
	packRoot := t.TempDir()
	serverDir := t.TempDir()

	// Create mods dir in pack root
	os.MkdirAll(filepath.Join(packRoot, "mods"), 0o755)
	os.WriteFile(filepath.Join(packRoot, "mods", "mod1.jar"), []byte("mod"), 0o644)
	os.WriteFile(filepath.Join(packRoot, "start.sh"), []byte("#!/bin/sh"), 0o755)
	os.WriteFile(filepath.Join(packRoot, "user_jvm_args.txt"), []byte("skip"), 0o644)

	// Create old mods in server dir
	os.MkdirAll(filepath.Join(serverDir, "mods"), 0o755)
	os.WriteFile(filepath.Join(serverDir, "mods", "old.jar"), []byte("old"), 0o644)

	if err := UpdateFromPackRoot(packRoot, serverDir); err != nil {
		t.Fatalf("UpdateFromPackRoot: %v", err)
	}

	// New mod should exist
	if _, err := os.Stat(filepath.Join(serverDir, "mods", "mod1.jar")); err != nil {
		t.Error("mod1.jar missing")
	}
	// Old mod should be gone
	if _, err := os.Stat(filepath.Join(serverDir, "mods", "old.jar")); !os.IsNotExist(err) {
		t.Error("old.jar should be removed")
	}
	// start.sh copied
	if _, err := os.Stat(filepath.Join(serverDir, "start.sh")); err != nil {
		t.Error("start.sh missing")
	}
	// user_jvm_args.txt should NOT be copied
	if _, err := os.Stat(filepath.Join(serverDir, "user_jvm_args.txt")); !os.IsNotExist(err) {
		t.Error("user_jvm_args.txt should not be copied")
	}
}
