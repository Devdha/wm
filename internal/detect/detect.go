package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// DetectionResult contains detected package manager info
type DetectionResult struct {
	PackageManager string
	InstallCommand string
	IsMonorepo     bool
}

// Detect analyzes a directory and returns package manager info
func Detect(dir string) DetectionResult {
	// Check pnpm first (pnpm-workspace.yaml)
	if fileExists(filepath.Join(dir, "pnpm-workspace.yaml")) {
		return DetectionResult{
			PackageManager: "pnpm",
			InstallCommand: "pnpm install",
			IsMonorepo:     true,
		}
	}

	// Check for pnpm-lock.yaml
	if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
		return DetectionResult{
			PackageManager: "pnpm",
			InstallCommand: "pnpm install",
			IsMonorepo:     false,
		}
	}

	// Check yarn (yarn.lock)
	if fileExists(filepath.Join(dir, "yarn.lock")) {
		isMonorepo := hasWorkspacesInPackageJson(dir)
		return DetectionResult{
			PackageManager: "yarn",
			InstallCommand: "yarn install",
			IsMonorepo:     isMonorepo,
		}
	}

	// Check npm (package-lock.json or package.json with workspaces)
	if fileExists(filepath.Join(dir, "package-lock.json")) || fileExists(filepath.Join(dir, "package.json")) {
		isMonorepo := hasWorkspacesInPackageJson(dir)
		pm := "npm"
		if isMonorepo && !fileExists(filepath.Join(dir, "package-lock.json")) {
			// Has workspaces but no lock file - could be npm
		}
		return DetectionResult{
			PackageManager: pm,
			InstallCommand: "npm install",
			IsMonorepo:     isMonorepo,
		}
	}

	// Check Cargo (Cargo.toml with workspace)
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		isMonorepo := isCargoWorkspace(dir)
		return DetectionResult{
			PackageManager: "cargo",
			InstallCommand: "cargo build",
			IsMonorepo:     isMonorepo,
		}
	}

	// Check Go (go.work for workspace)
	if fileExists(filepath.Join(dir, "go.work")) {
		return DetectionResult{
			PackageManager: "go",
			InstallCommand: "go mod download",
			IsMonorepo:     true,
		}
	}

	// Check Go (go.mod for single module)
	if fileExists(filepath.Join(dir, "go.mod")) {
		return DetectionResult{
			PackageManager: "go",
			InstallCommand: "go mod download",
			IsMonorepo:     false,
		}
	}

	return DetectionResult{}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasWorkspacesInPackageJson(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}

	var pkg struct {
		Workspaces interface{} `json:"workspaces"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	return pkg.Workspaces != nil
}

func isCargoWorkspace(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	if err != nil {
		return false
	}

	return strings.Contains(string(data), "[workspace]")
}
