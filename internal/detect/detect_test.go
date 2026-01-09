package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPnpm(t *testing.T) {
	tmpDir := t.TempDir()

	// Create pnpm-workspace.yaml
	if err := os.WriteFile(filepath.Join(tmpDir, "pnpm-workspace.yaml"), []byte("packages:\n  - 'apps/*'"), 0644); err != nil {
		t.Fatal(err)
	}

	result := Detect(tmpDir)

	if result.PackageManager != "pnpm" {
		t.Errorf("expected pnpm, got %s", result.PackageManager)
	}

	if result.InstallCommand != "pnpm install" {
		t.Errorf("expected 'pnpm install', got %s", result.InstallCommand)
	}
}

func TestDetectNpm(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package.json with workspaces
	content := `{"workspaces": ["packages/*"]}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result := Detect(tmpDir)

	if result.PackageManager != "npm" {
		t.Errorf("expected npm, got %s", result.PackageManager)
	}
}

func TestDetectYarn(t *testing.T) {
	tmpDir := t.TempDir()

	// Create yarn.lock
	if err := os.WriteFile(filepath.Join(tmpDir, "yarn.lock"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	// Create package.json
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	result := Detect(tmpDir)

	if result.PackageManager != "yarn" {
		t.Errorf("expected yarn, got %s", result.PackageManager)
	}
}

func TestDetectCargo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Cargo.toml with workspace
	content := `[workspace]
members = ["crates/*"]`
	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result := Detect(tmpDir)

	if result.PackageManager != "cargo" {
		t.Errorf("expected cargo, got %s", result.PackageManager)
	}

	if result.InstallCommand != "cargo build" {
		t.Errorf("expected 'cargo build', got %s", result.InstallCommand)
	}
}

func TestDetectGo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.work
	if err := os.WriteFile(filepath.Join(tmpDir, "go.work"), []byte("go 1.21\nuse ./cmd"), 0644); err != nil {
		t.Fatal(err)
	}

	result := Detect(tmpDir)

	if result.PackageManager != "go" {
		t.Errorf("expected go, got %s", result.PackageManager)
	}

	if result.InstallCommand != "go mod download" {
		t.Errorf("expected 'go mod download', got %s", result.InstallCommand)
	}
}

func TestDetectPoetry(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "poetry.lock"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	result := Detect(tmpDir)

	if result.PackageManager != "poetry" {
		t.Errorf("expected poetry, got %s", result.PackageManager)
	}
	if result.InstallCommand != "poetry install" {
		t.Errorf("expected 'poetry install', got %s", result.InstallCommand)
	}
}

func TestDetectPipenv(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "Pipfile"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	result := Detect(tmpDir)

	if result.PackageManager != "pipenv" {
		t.Errorf("expected pipenv, got %s", result.PackageManager)
	}
	if result.InstallCommand != "pipenv install" {
		t.Errorf("expected 'pipenv install', got %s", result.InstallCommand)
	}
}

func TestDetectPipRequirements(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("flask==2.0"), 0644); err != nil {
		t.Fatal(err)
	}

	result := Detect(tmpDir)

	if result.PackageManager != "pip" {
		t.Errorf("expected pip, got %s", result.PackageManager)
	}
	if result.InstallCommand != "pip install -r requirements.txt" {
		t.Errorf("expected 'pip install -r requirements.txt', got %s", result.InstallCommand)
	}
}

func TestDetectNone(t *testing.T) {
	tmpDir := t.TempDir()

	result := Detect(tmpDir)

	if result.PackageManager != "" {
		t.Errorf("expected empty, got %s", result.PackageManager)
	}
}
