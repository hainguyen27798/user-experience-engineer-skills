package scaffold

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	projecttemplates "github.com/hainguyen27798/user-experience-engineer-skills/templates"
)

var projectNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)
var nuxtFallbackCompatibilityDatePattern = regexp.MustCompile(`\bfallbackCompatibilityDate\s*=\s*["'](\d{4}-\d{2}-\d{2})["']`)

var workflowSkillIDs = []string{
	"create-implementation-plan",
	"design-frontend-ui",
	"implement-frontend-change",
	"review-frontend-change",
	"test-frontend-change",
	"verify-frontend-change",
	"fix-frontend-bug",
}

var workflowRouting = map[string]string{
	"create-implementation-plan": "" +
		"- Repository orientation: `caveman-explore`.\n" +
		"- Narrow feature scope: `lean-build`.\n" +
		"- Compatibility-safe transitions: `migration`.\n",
	"design-frontend-ui": "" +
		"- UI systems and accessibility: `ui-ux-pro-max`.\n" +
		"- Visual-quality review against the approved project direction: `design-taste-frontend`.\n",
	"implement-frontend-change": "" +
		"- Smallest complete implementation: `lean-build`.\n" +
		"- Behavior-preserving structure changes: `safe-refactor`.\n" +
		"- Existing-style visual implementation checks: `ui-ux-pro-max` or `design-taste-frontend`.\n",
	"review-frontend-change": "" +
		"- Concise diff review: `caveman-review`.\n" +
		"- UI, accessibility, responsive, and visual-quality review: `ui-ux-pro-max` or `design-taste-frontend`.\n",
	"verify-frontend-change": "- Final acceptance proof: `verify-and-stop`.\n",
	"fix-frontend-bug": "" +
		"- Root-cause analysis: `investigate-first`.\n" +
		"- Narrow fixes: `surgical-patch`.\n" +
		"- Behavior-preserving cleanup: `safe-refactor`.\n" +
		"- Visual defects: `ui-ux-pro-max` or `design-taste-frontend`.\n",
}

type Options struct {
	TargetDirectory string
	IncludeAPI      bool
	UseCodex        bool
	Runner          Runner
	Download        Downloader
	InstallSkills   SkillInstaller
}

type Result struct {
	TargetDirectory string
	UIDirectory     string
	APIDirectory    string
	APIModule       string
}

func Create(ctx context.Context, options Options) (Result, error) {
	targetDirectory, projectName, initializeGit, existingWorktree, err := prepareTarget(options.TargetDirectory, options.IncludeAPI)
	if err != nil {
		return Result{}, err
	}
	if err := rejectManagedSymlinks(targetDirectory); err != nil {
		return Result{}, err
	}

	runner := options.Runner
	if runner == nil {
		runner = commandRunner{stdout: os.Stdout, stderr: os.Stderr}
	}
	download := options.Download
	if download == nil {
		download = downloadReference
	}
	installSkills := options.InstallSkills
	if installSkills == nil {
		installSkills = installOfficialSkills
	}
	if options.UseCodex {
		if err := checkCodex(ctx, targetDirectory, runner); err != nil {
			return Result{}, err
		}
	}

	uiName := projectName + "-ui"
	uiDirectory := filepath.Join(targetDirectory, uiName)
	apiName := projectName + "-api"
	apiDirectory := filepath.Join(targetDirectory, apiName)

	fmt.Printf("\nInitializing %s workspace...\n", projectName)
	if initializeGit {
		if err := runner.Run(ctx, targetDirectory, "git", "init"); err != nil {
			return Result{}, err
		}
	}
	if err := mergeGitignore(filepath.Join(targetDirectory, ".gitignore")); err != nil {
		return Result{}, err
	}

	fmt.Printf("\nCreating %s from scratch...\n", uiName)
	if err := scaffoldFrontend(uiDirectory, uiName); err != nil {
		return Result{}, err
	}

	fmt.Println("\nInstalling Nuxt, Nuxt UI, and frontend tooling...")
	if err := runner.Run(
		ctx,
		uiDirectory,
		"npm",
		"install",
		"nuxt@latest",
		"vue@latest",
		"vue-router@latest",
		"@nuxt/ui@latest",
	); err != nil {
		return Result{}, err
	}
	compatibilityDate, err := readNuxtCompatibilityDate(uiDirectory)
	if err != nil {
		return Result{}, err
	}
	if err := writeFrontendNuxtConfig(uiDirectory, uiName, compatibilityDate); err != nil {
		return Result{}, err
	}
	if err := runner.Run(
		ctx,
		uiDirectory,
		"npm",
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
	); err != nil {
		return Result{}, err
	}
	apiModule := ""
	if options.IncludeAPI {
		fmt.Printf("\nCreating %s with FastAPI and uv...\n", apiName)
		if err := runner.Run(
			ctx,
			targetDirectory,
			"uv",
			"init",
			"--app",
			"--package",
			"--name",
			apiName,
			"--vcs",
			"none",
			"--no-workspace",
			apiName,
		); err != nil {
			return Result{}, err
		}
		if err := runner.Run(ctx, apiDirectory, "uv", "add", "fastapi[standard]"); err != nil {
			return Result{}, err
		}
		if err := runner.Run(ctx, apiDirectory, "uv", "add", "--dev", "mypy", "ruff"); err != nil {
			return Result{}, err
		}
		apiModule, err = configureAPI(apiDirectory)
		if err != nil {
			return Result{}, err
		}
	}

	fmt.Println("\nWriting workspace files and project agents...")
	if err := configureWorkspace(targetDirectory, projectName, apiModule, existingWorktree); err != nil {
		return Result{}, err
	}

	fmt.Println("\nInstalling official latest skills and recording their revisions...")
	locks, err := installSkills(ctx, targetDirectory, runner)
	if err != nil {
		return Result{}, err
	}
	if err := writeSkillLock(targetDirectory, locks, existingWorktree); err != nil {
		return Result{}, err
	}

	fmt.Println("\nDownloading current Nuxt and Nuxt UI references...")
	if err := installReferences(ctx, targetDirectory, download, existingWorktree); err != nil {
		return Result{}, err
	}

	fmt.Println("\nPreparing Nuxt types and ESLint config...")
	if err := runner.Run(ctx, uiDirectory, "npm", "run", "postinstall"); err != nil {
		return Result{}, err
	}
	if options.UseCodex {
		fmt.Println("\nAsking Codex to inspect the project and reconcile agent rules...")
		if err := reconcileProjectRules(ctx, targetDirectory, runner, codexCreatePrompt(existingWorktree)); err != nil {
			return Result{}, fmt.Errorf("run Codex rule reconciliation: %w", err)
		}
	}

	result := Result{
		TargetDirectory: targetDirectory,
		UIDirectory:     uiName,
		APIModule:       apiModule,
	}
	if options.IncludeAPI {
		result.APIDirectory = apiName
	}
	return result, nil
}

func prepareTarget(value string, includeAPI bool) (string, string, bool, bool, error) {
	if strings.TrimSpace(value) == "" {
		value = "."
	}
	targetDirectory, err := filepath.Abs(value)
	if err != nil {
		return "", "", false, false, fmt.Errorf("resolve target path: %w", err)
	}
	projectName := filepath.Base(targetDirectory)
	if !projectNamePattern.MatchString(projectName) {
		return "", "", false, false, errors.New("project directory name must use lowercase letters, numbers, dots, hyphens, or underscores")
	}

	initializeGit := true
	existingWorktree := false
	entries, err := os.ReadDir(targetDirectory)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(targetDirectory, 0o755); err != nil {
			return "", "", false, false, fmt.Errorf("create target directory: %w", err)
		}
	case err != nil:
		return "", "", false, false, fmt.Errorf("inspect target directory: %w", err)
	default:
		worktreeEntries := 0
		for _, entry := range entries {
			if entry.Name() == ".gitignore" {
				info, infoErr := entry.Info()
				if infoErr != nil {
					return "", "", false, false, fmt.Errorf("inspect existing .gitignore: %w", infoErr)
				}
				if entry.Type()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
					return "", "", false, false, errors.New("existing .gitignore must be a regular file")
				}
			}
			if entry.Name() != ".git" {
				worktreeEntries++
				continue
			}
			info, infoErr := entry.Info()
			if infoErr != nil {
				return "", "", false, false, fmt.Errorf("inspect existing .git metadata: %w", infoErr)
			}
			if entry.Type()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
				return "", "", false, false, errors.New("existing .git metadata must be a directory or regular gitfile")
			}
			initializeGit = false
		}
		existingWorktree = worktreeEntries > 0
		if existingWorktree && includeAPI {
			return "", "", false, false, fmt.Errorf("cannot add FastAPI to non-empty target directory: %s; use --no-api to add only the UI", targetDirectory)
		}
	}

	uiDirectory := filepath.Join(targetDirectory, projectName+"-ui")
	if _, err := os.Lstat(uiDirectory); err == nil {
		return "", "", false, false, fmt.Errorf("UI target already exists: %s", uiDirectory)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", "", false, false, fmt.Errorf("inspect UI target: %w", err)
	}

	return targetDirectory, projectName, initializeGit, existingWorktree, nil
}

func scaffoldFrontend(targetDirectory, packageName string) error {
	replacements := map[string]string{"UI_PACKAGE": packageName}
	files := map[string]string{
		"package.json":            "frontend/package.json",
		"tsconfig.json":           "frontend/tsconfig.json",
		"vitest.config.ts":        "frontend/vitest.config.ts",
		".env.example":            "frontend/env.example",
		".prettierrc":             "frontend/prettierrc",
		"eslint.config.mjs":       "frontend/eslint.config.mjs",
		"AGENTS.md":               "AGENTS.md",
		"README.md":               "frontend/README.md",
		"app/app.vue":             "frontend/app/app.vue",
		"app/pages/index.vue":     "frontend/app/pages/index.vue",
		"app/assets/css/main.css": "frontend/app/assets/css/main.css",
		"docs/design-system.md":   "",
	}
	for relativePath, templateName := range files {
		destination := filepath.Join(targetDirectory, filepath.FromSlash(relativePath))
		if templateName == "" {
			if err := writeFile(destination, []byte(designSystem), 0o644); err != nil {
				return err
			}
			continue
		}
		if err := writeRenderedTemplate(destination, templateName, replacements, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func readNuxtCompatibilityDate(uiDirectory string) (string, error) {
	path := filepath.Join(uiDirectory, "node_modules", "nuxt", "dist", "index.js")
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read installed Nuxt compatibility defaults: %w", err)
	}
	matches := nuxtFallbackCompatibilityDatePattern.FindSubmatch(contents)
	if len(matches) != 2 {
		return "", errors.New("installed Nuxt does not expose a fallback compatibility date")
	}
	compatibilityDate := string(matches[1])
	if _, err := time.Parse(time.DateOnly, compatibilityDate); err != nil {
		return "", fmt.Errorf("installed Nuxt fallback compatibility date %q is invalid: %w", compatibilityDate, err)
	}
	return compatibilityDate, nil
}

func writeFrontendNuxtConfig(targetDirectory, packageName, compatibilityDate string) error {
	return writeRenderedTemplate(
		filepath.Join(targetDirectory, "nuxt.config.ts"),
		"frontend/nuxt.config.ts",
		map[string]string{
			"UI_PACKAGE":              packageName,
			"NUXT_COMPATIBILITY_DATE": compatibilityDate,
		},
		0o644,
	)
}

func configureAPI(targetDirectory string) (string, error) {
	entries, err := os.ReadDir(filepath.Join(targetDirectory, "src"))
	if err != nil {
		return "", fmt.Errorf("inspect generated uv package: %w", err)
	}
	var modules []string
	for _, entry := range entries {
		if entry.IsDir() {
			modules = append(modules, entry.Name())
		}
	}
	if len(modules) != 1 {
		return "", errors.New("could not recognize the generated uv package layout")
	}
	replacements := map[string]string{"API_MODULE": modules[0]}
	if err := writeRenderedTemplate(filepath.Join(targetDirectory, "src", modules[0], "main.py"), "api/main.py", replacements, 0o644); err != nil {
		return "", err
	}
	if err := writeRenderedTemplate(filepath.Join(targetDirectory, "README.md"), "api/README.md", replacements, 0o644); err != nil {
		return "", err
	}
	if err := writeRenderedTemplate(filepath.Join(targetDirectory, "AGENTS.md"), "api/AGENTS.md", replacements, 0o644); err != nil {
		return "", err
	}
	if err := appendRenderedTemplate(filepath.Join(targetDirectory, "pyproject.toml"), "api/pyproject-tools.toml"); err != nil {
		return "", err
	}
	return modules[0], nil
}

func appendRenderedTemplate(destination, templateName string) error {
	existing, err := os.ReadFile(destination)
	if err != nil {
		return fmt.Errorf("read %s: %w", destination, err)
	}
	contents, err := renderTemplate(templateName, nil)
	if err != nil {
		return err
	}
	separator := "\n"
	if len(existing) > 0 && existing[len(existing)-1] == '\n' {
		separator = ""
	}
	return writeFile(destination, append(existing, []byte(separator+contents)...), 0o644)
}

func configureWorkspace(targetDirectory, projectName, apiModule string, existingWorktree bool) error {
	variant := "ui-only"
	if apiModule != "" {
		variant = "with-api"
	}
	replacements := map[string]string{
		"PROJECT_NAME":  projectName,
		"UI_DIRECTORY":  projectName + "-ui",
		"API_DIRECTORY": projectName + "-api",
		"API_MODULE":    apiModule,
	}
	if !existingWorktree {
		for _, name := range []string{"README.md", "AGENTS.md", "release.sh"} {
			mode := os.FileMode(0o644)
			if name == "release.sh" {
				mode = 0o755
			}
			if err := writeRenderedTemplate(
				filepath.Join(targetDirectory, name),
				filepath.ToSlash(filepath.Join("workspace", variant, name)),
				replacements,
				mode,
			); err != nil {
				return err
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(targetDirectory, "plan"), 0o755); err != nil {
		return err
	}

	for _, skillID := range workflowSkillIDs {
		contents, err := renderTemplate(filepath.ToSlash(filepath.Join("skills", skillID, "SKILL.md")), replacements)
		if err != nil {
			return err
		}
		if routing := workflowRouting[skillID]; routing != "" {
			contents = strings.TrimRight(contents, "\n") + "\n\n## Installed Base Skill Routing\n\n" + routing
		}
		destination := filepath.Join(targetDirectory, ".agents", "skills", skillID, "SKILL.md")
		if err := writeWorkspaceFile(destination, []byte(contents), 0o644, existingWorktree); err != nil {
			return err
		}
	}

	for _, relativePath := range []string{
		"config.toml",
		"agents/planner.toml",
		"agents/reviewer.toml",
		"agents/cavecrew-investigator.toml",
		"agents/cavecrew-builder.toml",
		"agents/cavecrew-reviewer.toml",
	} {
		contents, err := renderTemplate("codex/"+relativePath, nil)
		if err != nil {
			return err
		}
		if err := writeWorkspaceFile(filepath.Join(targetDirectory, ".codex", filepath.FromSlash(relativePath)), []byte(contents), 0o644, existingWorktree); err != nil {
			return err
		}
	}
	return nil
}

func mergeGitignore(path string) error {
	generated, err := renderTemplate("workspace/gitignore", nil)
	if err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return writeFile(path, []byte(generated), 0o644)
	}
	if err != nil {
		return fmt.Errorf("inspect %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("existing .gitignore must be a regular file: %s", path)
	}
	existing, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	present := make(map[string]bool)
	for _, line := range strings.Split(string(existing), "\n") {
		present[strings.TrimSpace(line)] = true
	}
	var missing []string
	for _, line := range strings.Split(generated, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !present[line] {
			missing = append(missing, line)
			present[line] = true
		}
	}
	if len(missing) == 0 {
		return nil
	}
	separator := "\n\n"
	if len(existing) == 0 {
		separator = ""
	} else if existing[len(existing)-1] == '\n' {
		separator = "\n"
	}
	contents := append(existing, []byte(separator+"# create-ux-engineer\n"+strings.Join(missing, "\n")+"\n")...)
	return writeFile(path, contents, 0o644)
}

func writeWorkspaceFile(path string, contents []byte, mode os.FileMode, preserveExisting bool) error {
	if preserveExisting {
		if _, err := os.Lstat(path); err == nil {
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return writeFile(path, contents, mode)
}

func installReferences(ctx context.Context, targetDirectory string, download Downloader, preserveExisting bool) error {
	for _, source := range referenceSources {
		contents, err := download(ctx, source.URL)
		if err != nil {
			return err
		}
		if len(contents) == 0 || len(contents) > maxReferenceBytes {
			return fmt.Errorf("invalid reference %s size: %d", source.URL, len(contents))
		}
		if err := writeWorkspaceFile(
			filepath.Join(targetDirectory, ".agents", "references", source.FileName),
			contents,
			0o644,
			preserveExisting,
		); err != nil {
			return err
		}
	}
	return nil
}

func renderTemplate(name string, replacements map[string]string) (string, error) {
	contents, err := projecttemplates.Read(name)
	if err != nil {
		return "", err
	}
	rendered := string(contents)
	for key, value := range replacements {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
	}
	if strings.Contains(rendered, "{{") || strings.Contains(rendered, "}}") {
		return "", fmt.Errorf("template %s contains an unresolved placeholder", name)
	}
	return rendered, nil
}

func writeRenderedTemplate(destination, templateName string, replacements map[string]string, mode os.FileMode) error {
	contents, err := renderTemplate(templateName, replacements)
	if err != nil {
		return err
	}
	return writeFile(destination, []byte(contents), mode)
}

func writeFile(path string, contents []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, contents, mode); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

const designSystem = `# Project Design System

This file is the single source of truth for this project's visual language. Design, implementation, and review must use the same decisions.

## Project Direction

The existing UI is the baseline. Record only decisions supported by current screens, components, tokens, or explicit user direction. When a material visual choice has no clear evidence, ask the user instead of inventing one.

## Required Decisions

- Semantic color roles and light/dark behavior.
- Typography families, scale, weights, and line heights.
- Spacing, container, breakpoint, grid, and density rules.
- Radius, border, shadow, and elevation rules.
- Preferred Nuxt UI variants and custom patterns.
- Interaction states, motion, keyboard behavior, contrast, and target sizes.

## Enforcement

- Reuse Nuxt UI components, semantic tokens, CSS variables, and established patterns.
- Match existing layout, color, typography, sizing, spacing, shape, depth, responsive, and interaction patterns by default.
- Do not import a generic style from UI UX Pro Max, Taste, or another design tool. Map guidance back to project evidence.
- Keep raw reusable visual values in the theme or token layer.
- Record approved reusable decisions here before scattering one-off values through components.
`
