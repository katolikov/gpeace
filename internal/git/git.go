package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// RepoRoot returns the root directory of the current git repository.
func RepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// GetConflictedFiles returns file paths (relative to repo root) with unmerged conflicts.
func GetConflictedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

// StageFile runs git add on a resolved file.
func StageFile(path string) error {
	cmd := exec.Command("git", "add", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add %s: %w", path, err)
	}
	return nil
}
