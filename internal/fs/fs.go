package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReplaceDirs are the directories that get replaced during an update.
var ReplaceDirs = []string{"mods", "config", "scripts", "kubejs", "libraries", "defaultconfigs"}

// DetectPackRoot finds the directory containing a "mods" folder within depth 2.
func DetectPackRoot(extractedDir string) string {
	best := extractedDir
	filepath.WalkDir(extractedDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(extractedDir, path)
		depth := len(strings.Split(rel, string(filepath.Separator)))
		if rel == "." {
			depth = 0
		}
		if depth > 2 {
			return filepath.SkipDir
		}
		if d.Name() == "mods" {
			best = filepath.Dir(path)
			return filepath.SkipAll
		}
		return nil
	})
	return best
}

// ExtractZip extracts a ZIP file into destDir.
func ExtractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		// Prevent zip slip
		target := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) && filepath.Clean(target) != filepath.Clean(destDir) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0o755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("opening file in zip: %w", err)
		}

		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return fmt.Errorf("creating file: %w", err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return fmt.Errorf("extracting file: %w", err)
		}
		out.Close()
		rc.Close()
	}
	return nil
}

// CopyTreeContents copies all children of srcDir into destDir.
func CopyTreeContents(srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		src := filepath.Join(srcDir, entry.Name())
		dst := filepath.Join(destDir, entry.Name())
		if entry.IsDir() {
			os.RemoveAll(dst)
			if err := copyDir(src, dst); err != nil {
				return err
			}
		} else {
			if err := copyFile(src, dst); err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateFromPackRoot replaces modpack-managed directories and copies executables.
func UpdateFromPackRoot(packRoot, serverDir string) error {
	for _, d := range ReplaceDirs {
		src := filepath.Join(packRoot, d)
		info, err := os.Stat(src)
		if err != nil || !info.IsDir() {
			continue
		}
		dst := filepath.Join(serverDir, d)
		os.RemoveAll(dst)
		if err := copyDir(src, dst); err != nil {
			return fmt.Errorf("replacing %s: %w", d, err)
		}
	}

	// Copy top-level executables (*.jar, *.sh, *.bat) except user_jvm_args.txt
	entries, err := os.ReadDir(packRoot)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "user_jvm_args.txt" {
			continue
		}
		ext := filepath.Ext(name)
		if ext == ".jar" || ext == ".sh" || ext == ".bat" {
			if err := copyFile(filepath.Join(packRoot, name), filepath.Join(serverDir, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
