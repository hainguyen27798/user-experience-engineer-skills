//go:build integration

package scaffold

import (
	"context"
	"io"
	"os"
	"testing"
)

func TestOfficialSkillSources(t *testing.T) {
	locks, err := installOfficialSkills(
		context.Background(),
		t.TempDir(),
		commandRunner{stdout: io.Discard, stderr: os.Stderr},
	)
	if err != nil {
		t.Fatalf("installOfficialSkills() error = %v", err)
	}
	for _, skill := range []string{"caveman", "ui-ux-pro-max", "design-taste-frontend"} {
		lock, ok := locks[skill]
		if !ok {
			t.Fatalf("missing lock for %s", skill)
		}
		if len(lock.Ref) != 40 || len(lock.ComputedHash) != 64 {
			t.Fatalf("invalid lock for %s: %#v", skill, lock)
		}
	}
}
