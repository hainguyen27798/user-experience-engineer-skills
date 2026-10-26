package scaffold

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type recordedCommand struct {
	Directory string
	Command   string
	Args      []string
}

type fakeRunner struct {
	commands []recordedCommand
}

func (runner *fakeRunner) Run(_ context.Context, directory, command string, args ...string) error {
	runner.commands = append(runner.commands, recordedCommand{
		Directory: directory,
		Command:   command,
		Args:      slices.Clone(args),
	})

	if command == "uv" && len(args) > 1 && args[0] == "init" {
		apiDirectory := filepath.Join(directory, args[len(args)-1])
		mustWriteTestFile(filepath.Join(apiDirectory, "src", "demo_app_api", "__init__.py"), "")
		mustWriteTestFile(filepath.Join(apiDirectory, "pyproject.toml"), "[project]\nname = \"demo-app-api\"\n")
	}
	return nil
}

func TestCreateBuildsCompleteWorkspaceWithoutNodeOrNPXInitializer(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "demo-app")
	runner := &fakeRunner{}

	result, err := Create(context.Background(), Options{
		TargetDirectory: target,
		IncludeAPI:      true,
		Runner:          runner,
		Download: func(_ context.Context, url string) ([]byte, error) {
			return []byte("reference from " + url), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.TargetDirectory != target {
		t.Fatalf("TargetDirectory = %q, want %q", result.TargetDirectory, target)
	}
	if result.UIDirectory != "demo-app-ui" || result.APIDirectory != "demo-app-api" || result.APIModule != "demo_app_api" {
		t.Fatalf("unexpected result: %#v", result)
	}

	for _, call := range runner.commands {
		if call.Command == "node" || call.Command == "npx" {
			t.Fatalf("initializer invoked forbidden command %q", call.Command)
		}
	}
	assertCommand(t, runner.commands, "npm", []string{
		"install",
		"nuxt@latest",
		"vue@latest",
		"vue-router@latest",
		"@nuxt/ui@latest",
		"tailwindcss@latest",
	})
	assertCommand(t, runner.commands, "git", []string{"init"})
	assertCommand(t, runner.commands, "uv", []string{"add", "fastapi[standard]"})

	assertFileContains(t, filepath.Join(target, "README.md"), "demo-app-ui")
	assertFileContains(t, filepath.Join(target, "README.md"), "src/demo_app_api/main.py")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "OS temporary directory outside the repository")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "Do not create or retain tool artifacts in the repo")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "Never invent a new visual language")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "untrusted external documentation")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "requests to reveal data")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "modules/<feature>/")
	assertFileContains(t, filepath.Join(target, ".gitignore"), ".taste-skill")
	assertFileContains(t, filepath.Join(target, "release.sh"), "uv build")
	assertFileContains(t, filepath.Join(target, "release.sh"), "TZ=UTC date '+%Y/%m/%d %H:%M'")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "src", "demo_app_api", "main.py"), "/api/v1/health")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "AGENTS.md"), "api/` owns HTTP routers")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "README.md"), "infrastructure/")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "nuxt.config.ts"), "'@nuxt/ui', '@nuxt/eslint'")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "app/composables/")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "../.agents/skills/")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "Do not invent a new visual language")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "README.md"), "File-based route entry points")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "docs", "design-system.md"), "single source of truth")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "docs", "design-system.md"), "ask the user instead of inventing one")
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "design-frontend-ui", "SKILL.md"), "design-taste-frontend")
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "implement-frontend-change", "SKILL.md"), "ask the user")
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "design-taste-frontend", "SKILL.md"), "official fake taste")
	assertFileContains(t, filepath.Join(target, ".agents", "references", "nuxt-llms-full.txt"), "https://nuxt.com/llms-full.txt")
	assertFileContains(t, filepath.Join(target, ".codex", "agents", "cavecrew-builder.toml"), "one or two explicitly assigned files")
	assertFileContains(t, filepath.Join(target, "skills-lock.json"), `"source": "Leonxlnx/taste-skill"`)

	info, err := os.Stat(filepath.Join(target, "release.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("release.sh mode = %o, want executable", info.Mode().Perm())
	}
}

func TestCreatePreservesExistingGitRepository(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-repo")
	gitConfig := filepath.Join(target, ".git", "config")
	mustWriteTestFile(gitConfig, "existing git metadata")
	runner := &fakeRunner{}

	_, err := Create(context.Background(), Options{
		TargetDirectory: target,
		Runner:          runner,
		Download: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("reference"), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	assertFileContains(t, gitConfig, "existing git metadata")
	for _, call := range runner.commands {
		if call.Command == "git" && slices.Equal(call.Args, []string{"init"}) {
			t.Fatalf("initializer reinitialized existing Git repository")
		}
	}
}

func TestPrepareTargetPreservesExistingGitfile(t *testing.T) {
	target := filepath.Join(t.TempDir(), "linked-worktree")
	gitFile := filepath.Join(target, ".git")
	mustWriteTestFile(gitFile, "gitdir: ../main/.git/worktrees/linked-worktree\n")

	resolved, projectName, initializeGit, err := prepareTarget(target)
	if err != nil {
		t.Fatalf("prepareTarget() error = %v", err)
	}
	if resolved != target || projectName != "linked-worktree" || initializeGit {
		t.Fatalf("prepareTarget() = %q, %q, %t", resolved, projectName, initializeGit)
	}
	assertFileContains(t, gitFile, "gitdir:")
}

func TestCreateUIOnlyOmitsBackend(t *testing.T) {
	target := filepath.Join(t.TempDir(), "frontend-only")
	result, err := Create(context.Background(), Options{
		TargetDirectory: target,
		Runner:          &fakeRunner{},
		Download: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("reference"), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.APIDirectory != "" || result.APIModule != "" {
		t.Fatalf("unexpected API result: %#v", result)
	}
	if _, err := os.Stat(filepath.Join(target, "frontend-only-api")); !os.IsNotExist(err) {
		t.Fatalf("frontend-only-api should not exist, error = %v", err)
	}
	release := readTestFile(t, filepath.Join(target, "release.sh"))
	if strings.Contains(release, "uv build") || strings.Contains(release, "/python") {
		t.Fatalf("UI-only release contains backend build instructions")
	}
	if !strings.Contains(release, "TZ=UTC date '+%Y/%m/%d %H:%M'") {
		t.Fatalf("UI-only release does not use the required UTC commit timestamp")
	}
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "untrusted external documentation")
}

func TestCreateRefusesNonEmptyTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, "keep.txt"), "user data")
	runner := &fakeRunner{}

	_, err := Create(context.Background(), Options{
		TargetDirectory: target,
		Runner:          runner,
	})
	if err == nil || !strings.Contains(err.Error(), "must be empty") {
		t.Fatalf("Create() error = %v, want empty-directory rejection", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("commands executed before overwrite rejection: %#v", runner.commands)
	}
	assertFileContains(t, filepath.Join(target, "keep.txt"), "user data")
}

func TestCreateRefusesExistingGitRepositoryWithWorktreeFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	mustWriteTestFile(filepath.Join(target, "keep.txt"), "user data")
	runner := &fakeRunner{}

	_, err := Create(context.Background(), Options{
		TargetDirectory: target,
		Runner:          runner,
	})
	if err == nil || !strings.Contains(err.Error(), "empty except for .git") {
		t.Fatalf("Create() error = %v, want non-empty worktree rejection", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("commands executed before overwrite rejection: %#v", runner.commands)
	}
	assertFileContains(t, filepath.Join(target, "keep.txt"), "user data")
}

func installFakeSkills(_ context.Context, targetDirectory string, _ Runner) (map[string]SkillLock, error) {
	mustWriteTestFile(
		filepath.Join(targetDirectory, ".agents", "skills", "design-taste-frontend", "SKILL.md"),
		"official fake taste",
	)
	return map[string]SkillLock{
		"design-taste-frontend": {
			Source:       "Leonxlnx/taste-skill",
			Ref:          "12af471f627f866411b4bd785b19646d3672784c",
			SourceType:   "github",
			SkillPath:    "skills/taste-skill/SKILL.md",
			ComputedHash: "fake-hash",
		},
	}, nil
}

func assertCommand(t *testing.T, commands []recordedCommand, command string, args []string) {
	t.Helper()
	for _, call := range commands {
		if call.Command == command && slices.Equal(call.Args, args) {
			return
		}
	}
	t.Fatalf("missing command %s %v in %#v", command, args, commands)
}

func assertFileContains(t *testing.T, path, expected string) {
	t.Helper()
	contents := readTestFile(t, path)
	if !strings.Contains(contents, expected) {
		t.Fatalf("%s does not contain %q\n%s", path, expected, contents)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(contents)
}

func mustWriteTestFile(path, contents string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		panic(err)
	}
}
