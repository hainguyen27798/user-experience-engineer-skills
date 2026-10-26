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
		fmt.Fprintf(os.Stderr, "\nInitialization failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var includeAPI bool
	var skipAPI bool
	var showVersion bool
	flag.BoolVar(&includeAPI, "api", false, "create a FastAPI service managed by uv")
	flag.BoolVar(&skipAPI, "no-api", false, "create only the Nuxt frontend")
	flag.BoolVar(&showVersion, "version", false, "print the create-ux-engineer version")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Create a Nuxt workspace in the current directory.")
		fmt.Fprintln(flag.CommandLine.Output(), "\nUsage:")
		fmt.Fprintln(flag.CommandLine.Output(), "  create-ux-engineer [--api|--no-api|--version]")
		fmt.Fprintln(flag.CommandLine.Output(), "\nThe directory may contain existing .git metadata, but no worktree files.")
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
	if flag.NArg() != 0 {
		return errors.New("do not pass a path; cd into the empty project directory first")
	}

	if !includeAPI && !skipAPI {
		answer, err := ask("Create a FastAPI service with uv? [y/N]: ")
		if err != nil {
			return err
		}
		switch strings.ToLower(answer) {
		case "", "n", "no":
		case "y", "yes":
			includeAPI = true
		default:
			return errors.New("answer y or n for FastAPI setup")
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	result, err := scaffold.Create(ctx, scaffold.Options{
		TargetDirectory: ".",
		IncludeAPI:      includeAPI,
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

func ask(prompt string) (string, error) {
	fmt.Print(prompt)
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(answer), nil
}
