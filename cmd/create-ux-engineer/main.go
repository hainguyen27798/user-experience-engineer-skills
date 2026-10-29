package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/hainguyen27798/user-experience-engineer-skills/internal/scaffold"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nOperation failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var includeAPI bool
	var skipAPI bool
	var updateOnly bool
	var showVersion bool
	flag.BoolVar(&includeAPI, "api", false, "create a FastAPI service managed by uv")
	flag.BoolVar(&skipAPI, "no-api", false, "create only the Nuxt frontend")
	flag.BoolVar(&updateOnly, "update", false, "refresh skills and use Codex to reconcile rules in an existing project")
	flag.BoolVar(&showVersion, "version", false, "print the create-ux-engineer version")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Create or update a UX engineering workspace in the current directory.")
		fmt.Fprintln(flag.CommandLine.Output(), "\nUsage:")
		fmt.Fprintln(flag.CommandLine.Output(), "  create-ux-engineer [--api|--no-api|--update|--version]")
		fmt.Fprintln(flag.CommandLine.Output(), "\nFull-stack setup requires an empty worktree. UI-only setup may add a new frontend to an existing worktree.")
		fmt.Fprintln(flag.CommandLine.Output(), "The current directory base name becomes the project name.")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("create-ux-engineer %s\n", version)
		return nil
	}

	if includeAPI && skipAPI {
		return errors.New("choose only one of --api or --no-api")
	}
	if updateOnly && (includeAPI || skipAPI) {
		return errors.New("--update cannot be combined with --api or --no-api")
	}
	if flag.NArg() != 0 {
		return errors.New("do not pass a path; cd into the target project directory first")
	}
	if updateOnly {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := scaffold.Update(ctx, scaffold.UpdateOptions{TargetDirectory: "."}); err != nil {
			return err
		}
		fmt.Println("\nUpdated project skills, references, and agent rules.")
		return nil
	}

	input := bufio.NewReader(os.Stdin)
	if !includeAPI && !skipAPI {
		answer, err := askYesNo(input, "Create a FastAPI service with uv? [y/N]: ")
		if err != nil {
			return err
		}
		if answer {
			includeAPI = true
		}
	}
	useCodex, err := askYesNo(input, "Let Codex inspect this project, including any existing code, and reconcile agent rules after setup? [y/N]: ")
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	result, err := scaffold.Create(ctx, scaffold.Options{
		TargetDirectory: ".",
		IncludeAPI:      includeAPI,
		UseCodex:        useCodex,
	})
	if err != nil {
		return err
	}

	fmt.Printf("\nReady: %s\n", result.TargetDirectory)
	fmt.Printf("\n  npm --prefix %s run dev\n", result.UIDirectory)
	if result.APIDirectory != "" {
		fmt.Printf("  uv run --project %s fastapi dev %s/src/%s/main.py\n", result.APIDirectory, result.APIDirectory, result.APIModule)
	}
	return nil
}

func ask(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(answer), nil
}

func askYesNo(reader *bufio.Reader, prompt string) (bool, error) {
	answer, err := ask(reader, prompt)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(answer) {
	case "", "n", "no":
		return false, nil
	case "y", "yes":
		return true, nil
	default:
		return false, errors.New("answer y or n")
	}
}
