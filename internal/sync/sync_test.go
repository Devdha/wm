package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/donghun/wm/internal/config"
)

func TestSyncCopy(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(srcDir, ".env")
	if err := os.WriteFile(srcFile, []byte("SECRET=123"), 0644); err != nil {
		t.Fatal(err)
	}

	item := config.SyncItem{
		Src:  ".env",
		Dst:  ".env",
		Mode: "copy",
		When: "always",
	}

	if err := SyncFile(srcDir, dstDir, item); err != nil {
		t.Fatalf("SyncFile failed: %v", err)
	}

	// Verify copy
	dstFile := filepath.Join(dstDir, ".env")
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}

	if string(content) != "SECRET=123" {
		t.Errorf("expected 'SECRET=123', got '%s'", content)
	}
}

func TestSyncSymlink(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(srcDir, ".env")
	if err := os.WriteFile(srcFile, []byte("SECRET=456"), 0644); err != nil {
		t.Fatal(err)
	}

	item := config.SyncItem{
		Src:  ".env",
		Dst:  ".env",
		Mode: "symlink",
		When: "always",
	}

	if err := SyncFile(srcDir, dstDir, item); err != nil {
		t.Fatalf("SyncFile failed: %v", err)
	}

	// Verify symlink
	dstFile := filepath.Join(dstDir, ".env")
	info, err := os.Lstat(dstFile)
	if err != nil {
		t.Fatalf("failed to stat dest file: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink, got regular file")
	}
}

func TestSyncWhenMissing(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source and existing dest
	srcFile := filepath.Join(srcDir, ".env.example")
	dstFile := filepath.Join(dstDir, ".env")
	if err := os.WriteFile(srcFile, []byte("NEW"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstFile, []byte("EXISTING"), 0644); err != nil {
		t.Fatal(err)
	}

	item := config.SyncItem{
		Src:  ".env.example",
		Dst:  ".env",
		Mode: "copy",
		When: "missing",
	}

	if err := SyncFile(srcDir, dstDir, item); err != nil {
		t.Fatalf("SyncFile failed: %v", err)
	}

	// Should not overwrite
	content, _ := os.ReadFile(dstFile)
	if string(content) != "EXISTING" {
		t.Errorf("file was overwritten, expected 'EXISTING', got '%s'", content)
	}
}

func TestSyncWhenMissingCopiesIfNotExists(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source only (no dest)
	srcFile := filepath.Join(srcDir, ".env.example")
	if err := os.WriteFile(srcFile, []byte("NEW"), 0644); err != nil {
		t.Fatal(err)
	}

	item := config.SyncItem{
		Src:  ".env.example",
		Dst:  ".env",
		Mode: "copy",
		When: "missing",
	}

	if err := SyncFile(srcDir, dstDir, item); err != nil {
		t.Fatalf("SyncFile failed: %v", err)
	}

	// Should copy since dest doesn't exist
	dstFile := filepath.Join(dstDir, ".env")
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("expected file to be created: %v", err)
	}

	if string(content) != "NEW" {
		t.Errorf("expected 'NEW', got '%s'", content)
	}
}

func TestSyncSourceNotExists(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	item := config.SyncItem{
		Src:  ".env",
		Dst:  ".env",
		Mode: "copy",
		When: "always",
	}

	// Should not error when source doesn't exist
	if err := SyncFile(srcDir, dstDir, item); err != nil {
		t.Fatalf("SyncFile failed: %v", err)
	}

	// Dest should not exist
	dstFile := filepath.Join(dstDir, ".env")
	if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
		t.Error("dest file should not exist when source doesn't exist")
	}
}

func TestSyncAllWithGlob(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create multiple source files matching glob
	if err := os.WriteFile(filepath.Join(srcDir, ".env"), []byte("ROOT"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, ".env.local"), []byte("LOCAL"), 0644); err != nil {
		t.Fatal(err)
	}

	items := []config.SyncItem{
		{
			Src:  ".env*",
			Dst:  ".env*",
			Mode: "copy",
			When: "always",
		},
	}

	if err := SyncAll(srcDir, dstDir, items); err != nil {
		t.Fatalf("SyncAll failed: %v", err)
	}

	// Verify both files were copied
	content1, err := os.ReadFile(filepath.Join(dstDir, ".env"))
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}
	if string(content1) != "ROOT" {
		t.Errorf("expected 'ROOT', got '%s'", content1)
	}

	content2, err := os.ReadFile(filepath.Join(dstDir, ".env.local"))
	if err != nil {
		t.Fatalf("failed to read .env.local: %v", err)
	}
	if string(content2) != "LOCAL" {
		t.Errorf("expected 'LOCAL', got '%s'", content2)
	}
}

func TestSyncAllMultipleItems(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source files
	if err := os.WriteFile(filepath.Join(srcDir, ".env"), []byte("ENV"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "config.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	items := []config.SyncItem{
		{Src: ".env", Dst: ".env", Mode: "copy", When: "always"},
		{Src: "config.json", Dst: "config.json", Mode: "copy", When: "always"},
	}

	if err := SyncAll(srcDir, dstDir, items); err != nil {
		t.Fatalf("SyncAll failed: %v", err)
	}

	// Verify both files were copied
	if _, err := os.Stat(filepath.Join(dstDir, ".env")); os.IsNotExist(err) {
		t.Error(".env should exist")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "config.json")); os.IsNotExist(err) {
		t.Error("config.json should exist")
	}
}

func TestSyncCreatesDestDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create nested source file
	nestedDir := filepath.Join(srcDir, "config", "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "app.yaml"), []byte("key: value"), 0644); err != nil {
		t.Fatal(err)
	}

	item := config.SyncItem{
		Src:  "config/nested/app.yaml",
		Dst:  "config/nested/app.yaml",
		Mode: "copy",
		When: "always",
	}

	if err := SyncFile(srcDir, dstDir, item); err != nil {
		t.Fatalf("SyncFile failed: %v", err)
	}

	// Verify file was created with directory
	dstFile := filepath.Join(dstDir, "config", "nested", "app.yaml")
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}
	if string(content) != "key: value" {
		t.Errorf("expected 'key: value', got '%s'", content)
	}
}

func TestSyncPreservesFileMode(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source file with executable mode
	srcFile := filepath.Join(srcDir, "script.sh")
	if err := os.WriteFile(srcFile, []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatal(err)
	}

	item := config.SyncItem{
		Src:  "script.sh",
		Dst:  "script.sh",
		Mode: "copy",
		When: "always",
	}

	if err := SyncFile(srcDir, dstDir, item); err != nil {
		t.Fatalf("SyncFile failed: %v", err)
	}

	// Verify mode was preserved
	dstFile := filepath.Join(dstDir, "script.sh")
	info, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("failed to stat dest file: %v", err)
	}

	// Check executable bits are set
	if info.Mode().Perm()&0100 == 0 {
		t.Errorf("expected executable bit, got %v", info.Mode().Perm())
	}
}
