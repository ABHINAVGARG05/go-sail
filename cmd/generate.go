package cmd

import (
	"fmt"

	"github.com/TejasGhatte/go-sail/internal/errors"
	"github.com/TejasGhatte/go-sail/internal/scripts"
	"github.com/spf13/cobra"
)

var CreateProjectCommand *cobra.Command
var ProjectName string

// Flags
var (
	flagDryRun  bool
	flagModule  string
	flagGitInit bool
)

func init() {
	CreateProjectCommand = &cobra.Command{
		Use:   "create [project-name]",
		Short: "Creates a new Go project with your chosen framework, database, and ORM",
		Long: `Creates a new Go project by cloning a framework template and generating
database connection, migration, environment, and optionally Docker files.

Examples:
  go-sail create my-app
  go-sail create my-app --module github.com/myuser/my-app
  go-sail create my-app --dry-run
  go-sail create my-app --git-init`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ProjectName = args[0]
			ctx := cmd.Context()

			opts := scripts.CreateOptions{
				DryRun:     flagDryRun,
				ModulePath: flagModule,
				GitInit:    flagGitInit,
			}

			if err := scripts.CreateProject(ctx, ProjectName, opts); err != nil {
				if err == errors.ErrInterrupt {
					fmt.Println("Program Exited: interrupt")
				} else {
					fmt.Printf("Program Exited: %v\n", err)
				}
			}
		},
	}

	CreateProjectCommand.Flags().BoolVarP(&flagDryRun, "dry-run", "n", false, "Preview what would be generated without creating any files")
	CreateProjectCommand.Flags().StringVarP(&flagModule, "module", "m", "", "Custom Go module path (default: project name)")
	CreateProjectCommand.Flags().BoolVar(&flagGitInit, "git-init", false, "Initialize a git repository with an initial commit after scaffolding")
}
