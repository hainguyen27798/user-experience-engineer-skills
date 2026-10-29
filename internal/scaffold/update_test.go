package scaffold

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type updateErrorRunner struct {
	err error
}

func (runner updateErrorRunner) Run(_ context.Context, _ string, _ string, _ ...string) error {
	return runner.err
}

type codexOutputRunner struct {
	commands       []recordedCommand
	outputPath     string
	outputContents string
}

func (runner *codexOutputRunner) Run(_ context.Context, directory, command string, args ...string) error {
	runner.commands = append(runner.commands, recordedCommand{Directory: directory, Command: command, Args: args})
	if command == "codex" && slices.Contains(args, "exec") {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(directory, runner.outputPath)), 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, runner.outputPath), []byte(runner.outputContents), 0o644)
	}
	return nil
}

func TestUpdateRefreshesManagedInputsAndCallsCodexWithoutScaffolding(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	mustWriteTestFile(filepath.Join(target, ".agents", "skills", "design-taste-frontend", "SKILL.md"), "old skill")
	mustWriteTestFile(filepath.Join(target, ".agents", "references", "nuxt-llms-full.txt"), "old reference")
	mustWriteTestFile(filepath.Join(target, "skills-lock.json"), "old lock")
	runner := &fakeRunner{}

	err := Update(context.Background(), UpdateOptions{
		TargetDirectory: target,
		Runner:          runner,
		Download: func(_ context.Context, url string) ([]byte, error) {
			return []byte("updated reference from " + url), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	assertFileContains(t, filepath.Join(target, ".agents", "skills", "design-taste-frontend", "SKILL.md"), "official fake taste")
	assertFileContains(t, filepath.Join(target, ".agents", "references", "nuxt-llms-full.txt"), "updated reference from https://nuxt.com/llms-full.txt")
	assertFileContains(t, filepath.Join(target, "skills-lock.json"), `"computedHash": "fake-hash"`)
	if matches, err := filepath.Glob(filepath.Join(target, ".create-ux-engineer-update-*")); err != nil || len(matches) != 0 {
		t.Fatalf("update staging directories remain: %v, error = %v", matches, err)
	}
	if _, err := os.Stat(filepath.Join(target, "existing-project-ui")); !os.IsNotExist(err) {
		t.Fatalf("update mode should not scaffold a UI, error = %v", err)
	}

	var codexCall *recordedCommand
	for index := range runner.commands {
		call := &runner.commands[index]
		if call.Command == "npm" || call.Command == "uv" {
			t.Fatalf("update mode invoked scaffold dependency command: %#v", call)
		}
		if call.Command == "codex" {
			codexCall = call
		}
	}
	if codexCall == nil {
		t.Fatalf("update mode did not invoke Codex: %#v", runner.commands)
	}
	assertCommand(t, runner.commands, "codex", []string{"login", "status"})
	if codexCall.Directory == target || !slices.Contains(codexCall.Args, "--ephemeral") || !slices.Contains(codexCall.Args, "--skip-git-repo-check") || !slices.Contains(codexCall.Args, "workspace-write") {
		t.Fatalf("unexpected Codex invocation: %#v", codexCall)
	}
	approvalIndex, execIndex := slices.Index(codexCall.Args, "--ask-for-approval"), slices.Index(codexCall.Args, "exec")
	if approvalIndex == -1 || approvalIndex+1 >= len(codexCall.Args) || execIndex == -1 || approvalIndex > execIndex || codexCall.Args[approvalIndex+1] != "never" {
		t.Fatalf("Codex approval policy must be a global option before exec: %#v", codexCall.Args)
	}
	prompt := codexCall.Args[len(codexCall.Args)-1]
	for _, expected := range []string{"Inspect the repository before editing", "templates and CLI conventions should be applied to this specific project", "Do not modify application source", "Do not create a branch, commit, push, or open a pull request", "read-only to this run", target} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("Codex prompt missing %q:\n%s", expected, prompt)
		}
	}
}

func TestUpdateRejectsNonGitDirectoryBeforeCommands(t *testing.T) {
	runner := &fakeRunner{}
	err := Update(context.Background(), UpdateOptions{
		TargetDirectory: t.TempDir(),
		Runner:          runner,
	})
	if err == nil || !strings.Contains(err.Error(), "existing Git project") {
		t.Fatalf("Update() error = %v, want Git project rejection", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("commands ran before update target rejection: %#v", runner.commands)
	}
}

func TestUpdateRequiresAuthenticatedCodexBeforeDownloading(t *testing.T) {
	target := t.TempDir()
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	installedSkills := false
	codexError := errors.New("codex is not authenticated")

	err := Update(context.Background(), UpdateOptions{
		TargetDirectory: target,
		Runner:          updateErrorRunner{err: codexError},
		InstallSkills: func(_ context.Context, _ string, _ Runner) (map[string]SkillLock, error) {
			installedSkills = true
			return nil, nil
		},
	})
	if !errors.Is(err, codexError) || !strings.Contains(err.Error(), "installed and authenticated") {
		t.Fatalf("Update() error = %v, want Codex authentication failure", err)
	}
	if installedSkills {
		t.Fatalf("skills were downloaded before Codex authentication passed")
	}
}

func TestUpdateRemovesSkillsNoLongerPresentInManagedLock(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	mustWriteTestFile(filepath.Join(target, ".agents", "skills", "retired-official-skill", "SKILL.md"), "retired skill")
	retiredHash, err := hashDirectory(filepath.Join(target, ".agents", "skills", "retired-official-skill"))
	if err != nil {
		t.Fatal(err)
	}
	mustWriteTestFile(filepath.Join(target, "skills-lock.json"), `{"version":1,"skills":{"retired-official-skill":{"source":"official/repository","computedHash":"`+retiredHash+`"}}}`)

	err = Update(context.Background(), UpdateOptions{
		TargetDirectory: target,
		Runner:          &fakeRunner{},
		Download: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("reference"), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target, ".agents", "skills", "retired-official-skill")); !os.IsNotExist(err) {
		t.Fatalf("retired managed skill remains after update: %v", err)
	}
}

func TestUpdatePreservesModifiedSkillEvenWhenOldLockListsIt(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	mustWriteTestFile(filepath.Join(target, ".agents", "skills", "retired-official-skill", "SKILL.md"), "user-modified skill")
	mustWriteTestFile(filepath.Join(target, "skills-lock.json"), `{"version":1,"skills":{"retired-official-skill":{"computedHash":"old-managed-hash"}}}`)

	err := Update(context.Background(), UpdateOptions{
		TargetDirectory: target,
		Runner:          &fakeRunner{},
		Download: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("reference"), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "retired-official-skill", "SKILL.md"), "user-modified skill")
}

func TestReadManagedSkillNamesIgnoresSymlinkedLock(t *testing.T) {
	target := t.TempDir()
	externalLock := filepath.Join(t.TempDir(), "skills-lock.json")
	mustWriteTestFile(externalLock, `{"version":1,"skills":{}}`)
	if err := os.Symlink(externalLock, filepath.Join(target, "skills-lock.json")); err != nil {
		t.Fatal(err)
	}
	if names := readManagedSkillNames(target); len(names) != 0 {
		t.Fatalf("readManagedSkillNames() followed a symlinked lock: %v", names)
	}
}

func TestPromoteStagedPathsValidatesEverySourceBeforeReplacing(t *testing.T) {
	root := t.TempDir()
	staging := filepath.Join(root, "staging")
	destination := filepath.Join(root, "destination", "skill")
	source := filepath.Join(staging, "skill")
	mustWriteTestFile(filepath.Join(destination, "SKILL.md"), "existing skill")
	mustWriteTestFile(filepath.Join(source, "SKILL.md"), "updated skill")

	err := promoteStagedPaths(root, staging, []stagedReplacement{
		{Source: source, Destination: destination},
		{Source: filepath.Join(staging, "missing"), Destination: filepath.Join(root, "destination", "missing")},
	})
	if err == nil || !strings.Contains(err.Error(), "inspect staged path") {
		t.Fatalf("promoteStagedPaths() error = %v, want missing staged path", err)
	}
	assertFileContains(t, filepath.Join(destination, "SKILL.md"), "existing skill")
	assertFileContains(t, filepath.Join(source, "SKILL.md"), "updated skill")
}

func TestUpdateRejectsManagedSymlinkBeforeCommands(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	external := filepath.Join(t.TempDir(), "external")
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(target, ".agents")); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{}

	err := Update(context.Background(), UpdateOptions{TargetDirectory: target, Runner: runner})
	if err == nil || !strings.Contains(err.Error(), "must not contain symlinks") {
		t.Fatalf("Update() error = %v, want managed symlink rejection", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("commands ran before managed symlink rejection: %#v", runner.commands)
	}
}

func TestCodexReconciliationPromotesOnlyAllowedRuleFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	mustWriteTestFile(filepath.Join(target, "app.py"), "preserve application code")
	runner := &codexOutputRunner{outputPath: filepath.Join(codexRuleOutputDirectory, "AGENTS.md"), outputContents: "project-specific rules"}

	err := Update(context.Background(), UpdateOptions{
		TargetDirectory: target,
		Runner:          runner,
		Download: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("reference"), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "project-specific rules")
	assertFileContains(t, filepath.Join(target, "app.py"), "preserve application code")
}

func TestCodexReconciliationRejectsApplicationOutput(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	mustWriteTestFile(filepath.Join(target, "app.py"), "original application code")
	runner := &codexOutputRunner{outputPath: filepath.Join(codexRuleOutputDirectory, "app.py"), outputContents: "unauthorized change"}

	err := Update(context.Background(), UpdateOptions{
		TargetDirectory: target,
		Runner:          runner,
		Download: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("reference"), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err == nil || !strings.Contains(err.Error(), "outside the rule allowlist") {
		t.Fatalf("Update() error = %v, want application output rejection", err)
	}
	assertFileContains(t, filepath.Join(target, "app.py"), "original application code")
}

func TestCodexReconciliationDoesNotAllowOfficialSkillFiles(t *testing.T) {
	target := t.TempDir()
	mustWriteTestFile(filepath.Join(target, ".agents", "skills", "design-taste-frontend", "SKILL.md"), "official skill")
	relativePath := filepath.Join(".agents", "skills", "design-taste-frontend", "AGENTS.md")
	if isReconciliationRulePath(target, relativePath) {
		t.Fatalf("official skill output was included in the Codex reconciliation allowlist: %s", relativePath)
	}
}

func TestStageProjectRulesDoesNotExposeProjectInstructionsAtWorkspaceRoot(t *testing.T) {
	target := t.TempDir()
	staging := t.TempDir()
	mustWriteTestFile(filepath.Join(target, "AGENTS.md"), "existing agent instructions")
	mustWriteTestFile(filepath.Join(target, ".codex", "config.toml"), "existing project configuration")

	if err := stageProjectRules(target, staging); err != nil {
		t.Fatalf("stageProjectRules() error = %v", err)
	}
	assertFileContains(t, filepath.Join(staging, codexRuleOutputDirectory, "AGENTS.md"), "existing agent instructions")
	assertFileContains(t, filepath.Join(staging, codexRuleOutputDirectory, ".codex", "config.toml"), "existing project configuration")
	for _, path := range []string{filepath.Join(staging, "AGENTS.md"), filepath.Join(staging, ".codex")} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("project instructions escaped neutral rules directory at %s: %v", path, err)
		}
	}
}

func TestStageProjectRulesRejectsOversizedExistingRules(t *testing.T) {
	target := t.TempDir()
	staging := t.TempDir()
	mustWriteTestFile(filepath.Join(target, ".codex", "config.toml"), string(make([]byte, maxCodexRuleBytes+1)))

	err := stageProjectRules(target, staging)
	if err == nil || !strings.Contains(err.Error(), "existing project rules exceed") {
		t.Fatalf("stageProjectRules() error = %v, want existing-rule size rejection", err)
	}
}

func TestRollbackReplacementsRestoresPreviousContents(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "managed-skill")
	backup := filepath.Join(root, "backups", "managed-skill")
	mustWriteTestFile(filepath.Join(destination, "SKILL.md"), "updated skill")
	mustWriteTestFile(filepath.Join(backup, "SKILL.md"), "previous skill")

	cause := errors.New("later promotion failed")
	err := rollbackReplacements([]appliedReplacement{{destination: destination, backup: backup}}, cause)
	if !errors.Is(err, cause) {
		t.Fatalf("rollbackReplacements() error = %v, want original cause", err)
	}
	assertFileContains(t, filepath.Join(destination, "SKILL.md"), "previous skill")
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Fatalf("backup should be consumed after rollback, error = %v", err)
	}
}
