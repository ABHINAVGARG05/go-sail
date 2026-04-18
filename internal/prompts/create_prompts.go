package prompts

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/TejasGhatte/go-sail/internal/errors"
	"github.com/TejasGhatte/go-sail/internal/initializers"
)

var frameworks = []string{"fiber", "gin", "echo"}

func SelectFramework(ctx context.Context) (string, error) {
	var framework string
	prompt := &survey.Select{
		Message: "Choose a Go framework:",
		Options: frameworks,
		Default: "fiber",
		Help:    "Select the framework you want to use for your project",
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- survey.AskOne(prompt, &framework)
	}()

	select {
	case <-ctx.Done():
		return "", errors.ErrInterrupt
	case err := <-errCh:
		if err != nil {
			return "", err
		}
		return framework, nil
	}
}

func SelectDatabase(ctx context.Context) (string, error) {
	// Build database list dynamically from config
	databases := []string{}
	for name := range initializers.Config.Databases {
		databases = append(databases, name)
	}
	databases = append(databases, "None")

	var database string
	prompt := &survey.Select{
		Message: "Choose a database (or None):",
		Options: databases,
		Default: "None",
		Help:    "Select the database you want to use, or 'None' if you don't need one",
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- survey.AskOne(prompt, &database)
	}()
	select {
	case <-ctx.Done():
		return "", errors.ErrInterrupt
	case err := <-errCh:
		if err != nil {
			return "", err
		}
		if database == "None" {
			return "", nil
		}
		return database, nil
	}
}

// SelectORM shows only ORM options that are valid for the selected database.
// If only one ORM is available (e.g. mongo-driver for mongodb), it auto-selects it.
func SelectORM(ctx context.Context, database string) (string, error) {
	// Get valid ORMs from the combinations config for this database
	combos, exists := initializers.Config.Combinations[database]
	if !exists {
		return "", fmt.Errorf("no ORM combinations found for database %q", database)
	}

	validORMs := []string{}
	for ormName := range combos {
		validORMs = append(validORMs, ormName)
	}

	// If only one valid ORM, auto-select it
	if len(validORMs) == 1 {
		fmt.Printf("Auto-selected ORM: %s (only valid option for %s)\n", validORMs[0], database)
		return validORMs[0], nil
	}

	validORMs = append(validORMs, "None")

	var orm string
	prompt := &survey.Select{
		Message: "Choose an ORM (or None):",
		Options: validORMs,
		Default: validORMs[0],
		Help:    "Select an ORM for database interactions, or 'None' if you don't need one",
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- survey.AskOne(prompt, &orm)
	}()

	select {
	case <-ctx.Done():
		return "", errors.ErrInterrupt
	case err := <-errCh:
		if err != nil {
			return "", err
		}
		if orm == "None" {
			return "", nil
		}
		return orm, nil
	}
}

// AskDockerSetup asks the user if they want Docker files generated.
func AskDockerSetup(ctx context.Context) (bool, error) {
	var wantDocker bool
	prompt := &survey.Confirm{
		Message: "Generate Dockerfile and docker-compose.yml?",
		Default: false,
		Help:    "Generates a multi-stage Dockerfile and docker-compose.yml pre-wired with your selected database",
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- survey.AskOne(prompt, &wantDocker)
	}()

	select {
	case <-ctx.Done():
		return false, errors.ErrInterrupt
	case err := <-errCh:
		if err != nil {
			return false, err
		}
		return wantDocker, nil
	}
}

// AskEnvSetup asks the user if they want a .env.example file generated.
func AskEnvSetup(ctx context.Context) (bool, error) {
	var wantEnv bool
	prompt := &survey.Confirm{
		Message: "Generate .env.example with database credentials template?",
		Default: true,
		Help:    "Creates a .env.example file with placeholder database credentials",
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- survey.AskOne(prompt, &wantEnv)
	}()

	select {
	case <-ctx.Done():
		return false, errors.ErrInterrupt
	case err := <-errCh:
		if err != nil {
			return false, err
		}
		return wantEnv, nil
	}
}
