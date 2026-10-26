package scaffold

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type SkillInstaller func(ctx context.Context, targetDirectory string, runner Runner) (map[string]SkillLock, error)

type SkillLock struct {
	Source       string `json:"source"`
	Ref          string `json:"ref"`
	SourceType   string `json:"sourceType"`
	SkillPath    string `json:"skillPath"`
	ComputedHash string `json:"computedHash"`
}

type skillSource struct {
	Repository string
	Skills     []sourceSkill
}

type sourceSkill struct {
	Name string
	Path string
}

var officialSkillSources = []skillSource{
	{
		Repository: "JuliusBrussee/caveman",
		Skills: skillNames("skills", []string{
			"caveman",
			"ultracave",
			"megacave",
			"cavecrew",
			"caveman-explore",
			"caveman-compress",
			"lean-build",
			"investigate-first",
			"surgical-patch",
			"safe-refactor",
			"migration",
			"caveman-review",
			"verify-and-stop",
			"caveman-commit",
			"caveman-learn",
		}),
	},
	{
		Repository: "nextlevelbuilder/ui-ux-pro-max-skill",
		Skills: []sourceSkill{{
			Name: "ui-ux-pro-max",
			Path: ".claude/skills/ui-ux-pro-max",
		}},
	},
	{
		Repository: "Leonxlnx/taste-skill",
		Skills: []sourceSkill{{
			Name: "design-taste-frontend",
			Path: "skills/taste-skill",
		}},
	},
}

func skillNames(parent string, names []string) []sourceSkill {
	result := make([]sourceSkill, 0, len(names))
	for _, name := range names {
		result = append(result, sourceSkill{Name: name, Path: filepath.ToSlash(filepath.Join(parent, name))})
	}
	return result
}

func installOfficialSkills(ctx context.Context, targetDirectory string, runner Runner) (map[string]SkillLock, error) {
	temporaryDirectory, err := os.MkdirTemp("", "create-ux-engineer-skills-")
	if err != nil {
		return nil, fmt.Errorf("create skill staging directory: %w", err)
	}
	defer os.RemoveAll(temporaryDirectory)

	locks := make(map[string]SkillLock)
	for index, source := range officialSkillSources {
		repositoryDirectory := filepath.Join(temporaryDirectory, fmt.Sprintf("source-%d", index))
		repositoryURL := "https://github.com/" + source.Repository + ".git"
		if err := runner.Run(ctx, temporaryDirectory, "git", "clone", "--depth", "1", repositoryURL, repositoryDirectory); err != nil {
			return nil, err
		}
		revision, err := gitOutput(ctx, repositoryDirectory, "rev-parse", "HEAD")
		if err != nil {
			return nil, err
		}

		for _, skill := range source.Skills {
			sourceDirectory := filepath.Join(repositoryDirectory, filepath.FromSlash(skill.Path))
			if _, err := os.Stat(filepath.Join(sourceDirectory, "SKILL.md")); err != nil {
				return nil, fmt.Errorf("official skill %s is missing at %s#%s: %w", skill.Name, source.Repository, revision, err)
			}
			destination := filepath.Join(targetDirectory, ".agents", "skills", skill.Name)
			if err := copyDirectory(sourceDirectory, destination); err != nil {
				return nil, fmt.Errorf("install official skill %s: %w", skill.Name, err)
			}
			hash, err := hashDirectory(destination)
			if err != nil {
				return nil, err
			}
			locks[skill.Name] = SkillLock{
				Source:       source.Repository,
				Ref:          revision,
				SourceType:   "github",
				SkillPath:    skill.Path + "/SKILL.md",
				ComputedHash: hash,
			}
		}
	}
	return locks, nil
}

func gitOutput(ctx context.Context, directory string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", directory}, args...)...)
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("git %v: %w", args, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func copyDirectory(source, destination string) error {
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination already exists: %s", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(destination, relativePath)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refuse symlink in skill source: %s", path)
		}
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copyFile(path, targetPath, info.Mode().Perm())
	})
}

func copyFile(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func hashDirectory(directory string) (string, error) {
	hash := sha256.New()
	var paths []string
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	for _, path := range paths {
		relativePath, err := filepath.Rel(directory, path)
		if err != nil {
			return "", err
		}
		hash.Write([]byte(filepath.ToSlash(relativePath)))
		contents, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		hash.Write(contents)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeSkillLock(targetDirectory string, skills map[string]SkillLock) error {
	contents, err := json.MarshalIndent(struct {
		Version int                  `json:"version"`
		Skills  map[string]SkillLock `json:"skills"`
	}{Version: 1, Skills: skills}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode skills-lock.json: %w", err)
	}
	return os.WriteFile(filepath.Join(targetDirectory, "skills-lock.json"), append(contents, '\n'), 0o644)
}
