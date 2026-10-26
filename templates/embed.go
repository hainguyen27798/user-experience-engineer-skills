package templates

import (
	"embed"
	"fmt"
)

// Files contains every file emitted by the initializer.
//
//go:embed AGENTS.md api codex frontend skills workspace
var Files embed.FS

func Read(name string) ([]byte, error) {
	contents, err := Files.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read template %q: %w", name, err)
	}
	return contents, nil
}
