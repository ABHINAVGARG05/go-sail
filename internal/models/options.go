package models

type Options struct {
	ProjectName string
	Framework   string
	Database    string
	ORM         string
	ModulePath  string // custom go module path (--module flag)
	DryRun      bool   // preview mode, no files created (--dry-run flag)
	GitInit     bool   // auto-init git repo after scaffolding (--git-init flag)
	Docker      bool   // generate Dockerfile + docker-compose.yml
	EnvFile     bool   // generate .env.example
}