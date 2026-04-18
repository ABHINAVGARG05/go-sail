package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RenameModulePath walks a project directory and replaces all occurrences of
// oldModulePath with newModulePath in .go files and go.mod.
// This fixes the template's internal imports to match the user's project name.
func RenameModulePath(projectDir, oldModulePath, newModulePath string) error {
	return filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, vendor, and .git
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vendor" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files and go.mod
		ext := filepath.Ext(path)
		base := filepath.Base(path)
		if ext != ".go" && base != "go.mod" && base != "go.sum" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", path, err)
		}

		original := string(content)
		replaced := strings.ReplaceAll(original, oldModulePath, newModulePath)

		if replaced != original {
			if err := os.WriteFile(path, []byte(replaced), info.Mode()); err != nil {
				return fmt.Errorf("failed to write %s: %v", path, err)
			}
		}

		return nil
	})
}
