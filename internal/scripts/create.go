package scripts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/TejasGhatte/go-sail/internal/helpers"
	"github.com/TejasGhatte/go-sail/internal/initializers"
	"github.com/TejasGhatte/go-sail/internal/models"
	"github.com/TejasGhatte/go-sail/internal/prompts"
	"github.com/briandowns/spinner"
)

// CreateOptions holds CLI flags for the create command.
type CreateOptions struct {
	DryRun     bool
	ModulePath string
	GitInit    bool
}

func CreateProject(ctx context.Context, name string, opts CreateOptions) error {
	framework, err := prompts.SelectFramework(ctx)
	if err != nil {
		return err
	}

	database, err := prompts.SelectDatabase(ctx)
	if err != nil {
		return err
	}

	var orm string
	if database != "" {
		orm, err = prompts.SelectORM(ctx, database)
		if err != nil {
			return err
		}
	}

	// Ask for Docker setup
	var wantDocker bool
	if database != "" {
		wantDocker, err = prompts.AskDockerSetup(ctx)
		if err != nil {
			return err
		}
	}

	// Ask for .env.example
	var wantEnv bool
	if database != "" {
		wantEnv, err = prompts.AskEnvSetup(ctx)
		if err != nil {
			return err
		}
	}

	// Resolve module path
	modulePath := name
	if opts.ModulePath != "" {
		modulePath = opts.ModulePath
	}

	options := &models.Options{
		ProjectName: name,
		Framework:   framework,
		Database:    database,
		ORM:         orm,
		ModulePath:  modulePath,
		DryRun:      opts.DryRun,
		GitInit:     opts.GitInit,
		Docker:      wantDocker,
		EnvFile:     wantEnv,
	}

	// Dry run: just print summary and exit
	if options.DryRun {
		printDryRun(options)
		return nil
	}

	fmt.Println("\nGenerating project with the following options:")
	fmt.Printf("  Project:    %s\n", name)
	fmt.Printf("  Module:     %s\n", modulePath)
	fmt.Printf("  Framework:  %s\n", framework)
	if database != "" {
		fmt.Printf("  Database:   %s\n", database)
		fmt.Printf("  ORM:        %s\n", orm)
	}
	if wantDocker {
		fmt.Printf("  Docker:     yes\n")
	}
	if wantEnv {
		fmt.Printf("  .env file:  yes\n")
	}
	fmt.Println()

	// Spinner
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	if err := PopulateDirectory(ctx, options); err != nil {
		return err
	}
	if err := runGoImports(name); err != nil {
		return fmt.Errorf("failed to run goimports: %v", err)
	}
	if err := runGoModCommands(name); err != nil {
		return fmt.Errorf("failed to run go mod commands: %v", err)
	}

	// Git init if requested
	if options.GitInit {
		if err := initGitRepo(name); err != nil {
			return fmt.Errorf("failed to initialize git repo: %v", err)
		}
	}

	s.Stop()

	fmt.Println("\nProject created successfully!")
	fmt.Printf("\n  cd %s\n", name)
	fmt.Println("  go run main.go")
	fmt.Println()

	return nil
}

func PopulateDirectory(ctx context.Context, options *models.Options) error {
	if err := GitClone(ctx, options.ProjectName, options.Framework, initializers.Config.Repositories[options.Framework]); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}

	currentDir, _ := os.Getwd()
	projectDir := filepath.Join(currentDir, options.ProjectName)
	initializersFolder := filepath.Join(projectDir, "initializers")

	// Rename module path in all .go files and go.mod
	templateModulePath := initializers.Config.TemplateModules[options.Framework]
	if templateModulePath != "" && options.ModulePath != "" {
		if err := helpers.RenameModulePath(projectDir, templateModulePath, options.ModulePath); err != nil {
			return fmt.Errorf("error renaming module path: %v", err)
		}
	}

	if options.Database != "" && options.ORM != "" {
		provider, err := helpers.ProviderFactory(options.Database, options.ORM)
		if err != nil {
			return fmt.Errorf("error creating database provider: %v", err)
		}

		if err := helpers.GenerateDatabaseFile(ctx, initializersFolder, provider); err != nil {
			return fmt.Errorf("error generating database file: %v", err)
		}

		if err := helpers.GenerateMigrationFile(ctx, initializersFolder, provider); err != nil {
			return fmt.Errorf("error generating migration file: %v", err)
		}
	}

	// Generate .env.example
	if options.EnvFile && options.Database != "" {
		if err := helpers.GenerateEnvFile(ctx, projectDir, options.Database); err != nil {
			return fmt.Errorf("error generating .env.example: %v", err)
		}
	}

	// Generate Docker files
	if options.Docker {
		if err := helpers.GenerateDockerfile(ctx, projectDir); err != nil {
			return fmt.Errorf("error generating Dockerfile: %v", err)
		}
		if err := helpers.GenerateDockerCompose(ctx, projectDir, options.Database); err != nil {
			return fmt.Errorf("error generating docker-compose.yml: %v", err)
		}
	}

	return nil
}

func printDryRun(options *models.Options) {
	fmt.Println("\n[DRY RUN] The following project would be created:")
	fmt.Printf("  Project:    %s\n", options.ProjectName)
	fmt.Printf("  Module:     %s\n", options.ModulePath)
	fmt.Printf("  Framework:  %s\n", options.Framework)
	if options.Database != "" {
		fmt.Printf("  Database:   %s\n", options.Database)
		fmt.Printf("  ORM:        %s\n", options.ORM)
	} else {
		fmt.Println("  Database:   none")
	}
	fmt.Printf("  Docker:     %v\n", options.Docker)
	fmt.Printf("  .env file:  %v\n", options.EnvFile)
	fmt.Printf("  Git init:   %v\n", options.GitInit)

	fmt.Println("\n  Files that would be generated:")
	fmt.Printf("    %s/\n", options.ProjectName)
	fmt.Println("    ├── main.go")
	fmt.Println("    ├── go.mod")
	if options.Database != "" && options.ORM != "" {
		fmt.Println("    ├── initializers/")
		fmt.Println("    │   ├── database.go")
		fmt.Println("    │   └── migrations.go")
	}
	if options.EnvFile {
		fmt.Println("    ├── .env.example")
	}
	if options.Docker {
		fmt.Println("    ├── Dockerfile")
		fmt.Println("    └── docker-compose.yml")
	}
	fmt.Println("\n  No files were created.")
}

func runGoModCommands(projectName string) error {
	currentDir, _ := os.Getwd()
	projectDir := filepath.Join(currentDir, projectName)

	commands := [][]string{
		{"go", "mod", "tidy"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = projectDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s command failed: %v", cmdArgs, err)
		}
	}

	return nil
}

func runGoImports(projectDir string) error {
	cmd := exec.Command("goimports", "-w", projectDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			fmt.Println("goimports not found, falling back to gofmt")
			cmdFmt := exec.Command("gofmt", "-w", projectDir)
			cmdFmt.Stdout = os.Stdout
			cmdFmt.Stderr = os.Stderr
			if err := cmdFmt.Run(); err != nil {
				return fmt.Errorf("gofmt command failed for directory %s: %v", projectDir, err)
			}
			return nil
		}
		return fmt.Errorf("goimports command failed for directory %s: %v", projectDir, err)
	}

	return nil
}

func initGitRepo(projectName string) error {
	currentDir, _ := os.Getwd()
	projectDir := filepath.Join(currentDir, projectName)

	commands := [][]string{
		{"git", "init"},
		{"git", "add", "."},
		{"git", "commit", "-m", "Initial commit from go-sail"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = projectDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%v failed: %v", cmdArgs, err)
		}
	}

	return nil
}