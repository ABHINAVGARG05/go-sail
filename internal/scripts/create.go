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

func CreateProject(ctx context.Context, name string) error {
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
		orm, err = prompts.SelectORM(ctx)
		if err != nil {
			return err
		}
	}

	fmt.Println("Generating project with the following options:")
	fmt.Printf("Framework: %s, Database: %s, ORM: %s\n", framework, database, orm)

	options := &models.Options{
		ProjectName: name,
		Framework:   framework,
		Database:    database,
		ORM:         orm,
	}

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
	return nil
}

func PopulateDirectory(ctx context.Context, options *models.Options) error {
	if err := GitClone(ctx, options.ProjectName, options.Framework, initializers.Config.Repositories[options.Framework]); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}

	currentDir, _ := os.Getwd()
	folder := filepath.Join(currentDir, options.ProjectName, "initializers")

	if options.Database != "" && options.ORM != "" {
		provider, err := helpers.ProviderFactory(options.Database, options.ORM)
		if err != nil {
			return fmt.Errorf("error creating database provider: %v", err)
		}

		if err := helpers.GenerateDatabaseFile(ctx, folder, provider); err != nil {
			return fmt.Errorf("error generating database file: %v", err)
		}

		if err := helpers.GenerateMigrationFile(ctx, folder, provider); err != nil {
			return fmt.Errorf("error generating migration file: %v", err)
		}
	}
	return nil
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