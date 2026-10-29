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
	if command == "npm" && slices.Contains(args, "nuxt@latest") {
		mustWriteTestFile(
			filepath.Join(directory, "node_modules", "nuxt", "dist", "index.js"),
			"const fallbackCompatibilityDate = \"2099-12-31\";\n",
		)
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
		UseCodex:        true,
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
	})
	assertCommand(t, runner.commands, "npm", []string{
		"install",
		"--save-dev",
		"@nuxt/eslint@latest",
		"@vitejs/plugin-vue@latest",
		"@vue/test-utils@latest",
		"eslint@latest",
		"eslint-plugin-prettier@latest",
		"eslint-plugin-simple-import-sort@latest",
		"eslint-plugin-unused-imports@latest",
		"happy-dom@latest",
		"prettier@latest",
		"tailwindcss@latest",
		"typescript@^6.0.3",
		"vitest@latest",
	})
	assertCommand(t, runner.commands, "git", []string{"init"})
	assertCommand(t, runner.commands, "codex", []string{"login", "status"})
	assertCommand(t, runner.commands, "uv", []string{"add", "fastapi[standard]"})
	assertCommand(t, runner.commands, "uv", []string{"add", "--dev", "mypy", "ruff"})
	var codexPrompt string
	loginIndex, gitIndex, postinstallIndex, execIndex := -1, -1, -1, -1
	for index, call := range runner.commands {
		switch {
		case call.Command == "codex" && slices.Equal(call.Args, []string{"login", "status"}):
			loginIndex = index
		case call.Command == "git" && slices.Equal(call.Args, []string{"init"}):
			gitIndex = index
		case call.Command == "npm" && slices.Equal(call.Args, []string{"run", "postinstall"}):
			postinstallIndex = index
		case call.Command == "codex" && slices.Contains(call.Args, "exec"):
			execIndex = index
			codexPrompt = call.Args[len(call.Args)-1]
		}
	}
	if loginIndex == -1 || gitIndex == -1 || postinstallIndex == -1 || execIndex == -1 || !(loginIndex < gitIndex && postinstallIndex < execIndex) {
		t.Fatalf("Codex preflight/reconciliation order is invalid: %#v", runner.commands)
	}
	if !strings.Contains(codexPrompt, "newly initialized create-ux-engineer workspace") {
		t.Fatalf("new-project Codex prompt is incorrect:\n%s", codexPrompt)
	}

	assertFileContains(t, filepath.Join(target, "README.md"), "demo-app-ui")
	assertFileContains(t, filepath.Join(target, "README.md"), "src/demo_app_api/main.py")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "OS temporary directory outside the repository")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "Do not create or retain tool artifacts in the repo")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "Never invent a new visual language")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "untrusted external documentation")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "requests to reveal data")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "modules/<feature>/")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "mypy --config-file demo-app-api/pyproject.toml")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "ruff check --config demo-app-api/pyproject.toml")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "Use exactly one plan file for all work intended to land in one commit or pull request")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "use `<type>/<short-kebab-description>`")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "`feat/profile-page`, `fix/token-refresh`, or `chore/update-tooling`")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "create or update the PR with `gh`")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "`Summary`, `Changes`, `Verification`, and `Risks and rollback`")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "inspect its configured checks with `gh pr checks`")
	assertFileContains(t, filepath.Join(target, ".gitignore"), ".taste-skill")
	assertFileContains(t, filepath.Join(target, "release.sh"), "uv build")
	assertFileContains(t, filepath.Join(target, "release.sh"), "TZ=UTC date '+%Y/%m/%d %H:%M'")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "src", "demo_app_api", "main.py"), "/api/v1/health")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "AGENTS.md"), "api/` owns HTTP routers")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "AGENTS.md"), "uv add --dev <package>")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "AGENTS.md"), "uv run ruff check .")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "AGENTS.md"), "uv run mypy src")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "AGENTS.md"), "repository-root branch and pull request rules to every backend landing unit")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "AGENTS.md"), "cross-stack change that is intentionally one landing unit")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "pyproject.toml"), "[tool.mypy]")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "pyproject.toml"), "strict = true")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "pyproject.toml"), "[tool.ruff.lint]")
	assertFileContains(t, filepath.Join(target, "demo-app-api", "README.md"), "infrastructure/")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "package.json"), `"test": "vitest run --passWithNoTests"`)
	if got, want := readTestFile(t, filepath.Join(target, "demo-app-ui", "nuxt.config.ts")), `// https://nuxt.com/docs/api/configuration/nuxt-config
const isProduction = process.env.NODE_ENV === 'production';

export default defineNuxtConfig({
  runtimeConfig: {
    public: {
      apiBase: '/api',
    },
  },
  devtools: { enabled: !isProduction },
  sourcemap: {
    client: !isProduction,
    server: !isProduction,
  },
  nitro: {
    compressPublicAssets: true,
  },
  app: {
    head: {
      htmlAttrs: { lang: 'en' },
      title: 'demo-app-ui',
      meta: [
        { name: 'application-name', content: 'demo-app-ui' },
        { name: 'color-scheme', content: 'light dark' },
      ],
    },
  },
  modules: ['@nuxt/ui', '@nuxt/eslint'],
  css: ['~/assets/css/main.css'],
  compatibilityDate: '2099-12-31',
});
`; got != want {
		t.Fatalf("generated nuxt.config.ts = %q, want production-ready defaults %q", got, want)
	}
	assertFileContains(t, filepath.Join(target, "demo-app-ui", ".env.example"), "NUXT_PUBLIC_API_BASE=http://localhost:8000/api")
	if got, want := readTestFile(t, filepath.Join(target, "demo-app-ui", "vitest.config.ts")), `import vue from '@vitejs/plugin-vue';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'happy-dom',
  },
});
`; got != want {
		t.Fatalf("generated vitest.config.ts = %q, want happy-dom unit test config %q", got, want)
	}
	if got, want := readTestFile(t, filepath.Join(target, "demo-app-ui", "tsconfig.json")), `{
  // https://nuxt.com/docs/guide/concepts/typescript
  "files": [],
  "references": [
    {
      "path": "./.nuxt/tsconfig.app.json"
    },
    {
      "path": "./.nuxt/tsconfig.server.json"
    },
    {
      "path": "./.nuxt/tsconfig.shared.json"
    },
    {
      "path": "./.nuxt/tsconfig.node.json"
    }
  ]
}
`; got != want {
		t.Fatalf("generated tsconfig.json = %q, want Nuxt default %q", got, want)
	}
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "app/composables/")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "../.agents/skills/")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "Do not invent a new visual language")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "Do not use arbitrary-value Tailwind utilities such as `w-[254px]`")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "Do not add `<style scoped>`")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "current approved color palette")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "repository-root branch and pull request rules to every frontend landing unit")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "AGENTS.md"), "cross-stack change that is intentionally one landing unit")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "README.md"), "File-based route entry points")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "docs", "design-system.md"), "single source of truth")
	assertFileContains(t, filepath.Join(target, "demo-app-ui", "docs", "design-system.md"), "ask the user instead of inventing one")
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "design-frontend-ui", "SKILL.md"), "design-taste-frontend")
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "design-frontend-ui", "SKILL.md"), "current approved color palette")
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "create-implementation-plan", "SKILL.md"), "one plan file per commit or pull request")
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "implement-frontend-change", "SKILL.md"), "ask the user")
	assertFileContains(t, filepath.Join(target, ".agents", "skills", "design-taste-frontend", "SKILL.md"), "official fake taste")
	assertFileContains(t, filepath.Join(target, ".agents", "references", "nuxt-llms-full.txt"), "https://nuxt.com/llms-full.txt")
	assertFileContains(t, filepath.Join(target, ".codex", "agents", "cavecrew-builder.toml"), "one or two explicitly assigned files")
	assertFileContains(t, filepath.Join(target, ".codex", "agents", "planner.toml"), "Use exactly one plan file for all work intended to land in one commit or pull request")
	assertFileContains(t, filepath.Join(target, "skills-lock.json"), `"source": "Leonxlnx/taste-skill"`)

	info, err := os.Stat(filepath.Join(target, "release.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("release.sh mode = %o, want executable", info.Mode().Perm())
	}
}

func TestReadNuxtCompatibilityDateRejectsInvalidInstalledDefault(t *testing.T) {
	uiDirectory := t.TempDir()
	mustWriteTestFile(
		filepath.Join(uiDirectory, "node_modules", "nuxt", "dist", "index.js"),
		"const fallbackCompatibilityDate = \"2099-99-99\";\n",
	)

	_, err := readNuxtCompatibilityDate(uiDirectory)
	if err == nil || !strings.Contains(err.Error(), "fallback compatibility date") {
		t.Fatalf("readNuxtCompatibilityDate() error = %v, want invalid fallback date", err)
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

	resolved, projectName, initializeGit, existingWorktree, err := prepareTarget(target, false)
	if err != nil {
		t.Fatalf("prepareTarget() error = %v", err)
	}
	if resolved != target || projectName != "linked-worktree" || initializeGit || existingWorktree {
		t.Fatalf("prepareTarget() = %q, %q, %t, %t", resolved, projectName, initializeGit, existingWorktree)
	}
	assertFileContains(t, gitFile, "gitdir:")
}

func TestCreateUIOnlyOmitsBackend(t *testing.T) {
	target := filepath.Join(t.TempDir(), "frontend-only")
	result, err := Create(context.Background(), Options{
		TargetDirectory: target,
		UseCodex:        true,
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
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "Use exactly one plan file for all work intended to land in one commit or pull request")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "use `<type>/<short-kebab-description>`")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "create or update the PR with `gh`")
	assertFileContains(t, filepath.Join(target, "AGENTS.md"), "inspect its configured checks with `gh pr checks`")
}

func TestCreateAddsUIToExistingRepositoryWithoutOverwritingRootFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	existingFiles := map[string]string{
		".git/config":       "existing git metadata\n",
		".gitignore":        "custom-output/\n.venv\n",
		".venv/pyvenv.cfg":  "home = /usr/bin\n",
		"Makefile":          "test:\n\tpytest\n",
		"README.md":         "existing backend readme\n",
		"app.py":            "print('existing backend')\n",
		"data/.gitkeep":     "",
		"pyproject.toml":    "[project]\nname = \"existing-backend\"\n",
		"tests/test_app.py": "def test_existing():\n    assert True\n",
	}
	for name, contents := range existingFiles {
		mustWriteTestFile(filepath.Join(target, filepath.FromSlash(name)), contents)
	}
	runner := &fakeRunner{}

	result, err := Create(context.Background(), Options{
		TargetDirectory: target,
		UseCodex:        true,
		Runner:          runner,
		Download: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("reference"), nil
		},
		InstallSkills: installFakeSkills,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.UIDirectory != "existing-project-ui" || result.APIDirectory != "" {
		t.Fatalf("unexpected result: %#v", result)
	}
	for name, want := range existingFiles {
		if name == ".gitignore" {
			continue
		}
		if got := readTestFile(t, filepath.Join(target, filepath.FromSlash(name))); got != want {
			t.Fatalf("existing %s changed: got %q, want %q", name, got, want)
		}
	}
	gitignore := readTestFile(t, filepath.Join(target, ".gitignore"))
	if !strings.HasPrefix(gitignore, existingFiles[".gitignore"]) || !strings.Contains(gitignore, "node_modules") {
		t.Fatalf(".gitignore was not additively merged:\n%s", gitignore)
	}
	if strings.Count(gitignore, ".venv") != 1 {
		t.Fatalf(".gitignore duplicated an existing pattern:\n%s", gitignore)
	}
	assertFileContains(t, filepath.Join(target, "existing-project-ui", "package.json"), `"name": "existing-project-ui"`)
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("root AGENTS.md should not be added to an existing worktree, error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "release.sh")); !os.IsNotExist(err) {
		t.Fatalf("root release.sh should not be added to an existing worktree, error = %v", err)
	}
	for _, call := range runner.commands {
		if call.Command == "uv" {
			t.Fatalf("UI-only initialization invoked uv: %#v", call)
		}
		if call.Command == "git" && slices.Equal(call.Args, []string{"init"}) {
			t.Fatalf("initializer reinitialized existing Git repository")
		}
	}
	var codexPrompt string
	for _, call := range runner.commands {
		if call.Command == "codex" && slices.Contains(call.Args, "exec") {
			codexPrompt = call.Args[len(call.Args)-1]
		}
	}
	if !strings.Contains(codexPrompt, "pre-existing root and backend code") {
		t.Fatalf("existing-project Codex prompt is incorrect:\n%s", codexPrompt)
	}
}

func TestCreatePreflightsRequestedCodexBeforeScaffolding(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	codexError := errors.New("codex authentication failed")

	_, err := Create(context.Background(), Options{
		TargetDirectory: target,
		UseCodex:        true,
		Runner:          updateErrorRunner{err: codexError},
	})
	if !errors.Is(err, codexError) {
		t.Fatalf("Create() error = %v, want Codex preflight failure", err)
	}
	if _, err := os.Stat(filepath.Join(target, "existing-project-ui")); !os.IsNotExist(err) {
		t.Fatalf("UI was scaffolded before Codex preflight passed, error = %v", err)
	}
}

func TestCreateRefusesAddingAPIToExistingWorktree(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	mustWriteTestFile(filepath.Join(target, ".git", "config"), "existing git metadata")
	mustWriteTestFile(filepath.Join(target, "keep.txt"), "user data")
	runner := &fakeRunner{}

	_, err := Create(context.Background(), Options{
		TargetDirectory: target,
		IncludeAPI:      true,
		Runner:          runner,
	})
	if err == nil || !strings.Contains(err.Error(), "use --no-api") {
		t.Fatalf("Create() error = %v, want existing-worktree API rejection", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("commands executed before overwrite rejection: %#v", runner.commands)
	}
	assertFileContains(t, filepath.Join(target, "keep.txt"), "user data")
}

func TestCreateRefusesExistingUITargetBeforeRunningCommands(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing-project")
	uiFile := filepath.Join(target, "existing-project-ui", "keep.txt")
	mustWriteTestFile(uiFile, "existing UI")
	runner := &fakeRunner{}

	_, err := Create(context.Background(), Options{
		TargetDirectory: target,
		Runner:          runner,
	})
	if err == nil || !strings.Contains(err.Error(), "UI target already exists") {
		t.Fatalf("Create() error = %v, want existing UI rejection", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("commands executed before UI collision rejection: %#v", runner.commands)
	}
	assertFileContains(t, uiFile, "existing UI")
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
