package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateDatabaseFile generates the database initialization file
func GenerateDatabaseFile(ctx context.Context, folderPath string, provider Provider) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filename := filepath.Join(folderPath, "database.go")
	tmpl, err := template.New("database").Parse(`package initializers

import (
    "fmt"
    {{range .Imports}}
    {{.}}
    {{- end}}
)

var DB {{.DBVariable}}

func ConnectDB(){
    {{.ConnectionCode}}
}
`)
	if err != nil {
		return fmt.Errorf("error parsing database template: %v", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating database file: %v", err)
	}
	defer f.Close()

	data := struct {
		Imports        []string
		ConnectionCode string
		DBVariable     string
	}{
		Imports:        provider.GetImports(),
		ConnectionCode: provider.GetConnectionCode(),
		DBVariable:     provider.GetDBVariable(),
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		return fmt.Errorf("error executing database template: %v", err)
	}

	return nil
}

// GenerateMigrationFile generates the migration file
func GenerateMigrationFile(ctx context.Context, folderPath string, provider Provider) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filename := filepath.Join(folderPath, "migrations.go")
	tmpl, err := template.New("migration").Parse(`package initializers

import (
    "fmt"
    {{range .Imports}}
    {{.}}
    {{- end}}
)

func DBMigrate() error {
    {{.MigrationCode}}
    return nil
}
`)
	if err != nil {
		return fmt.Errorf("error parsing migration template: %v", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating migration file: %v", err)
	}
	defer f.Close()

	data := struct {
		Imports       []string
		MigrationCode string
	}{
		Imports:       provider.GetImports(),
		MigrationCode: provider.GetMigrationCode(),
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		return fmt.Errorf("error executing migration template: %v", err)
	}

	return nil
}

// GenerateEnvFile generates a .env.example file with database credential placeholders.
func GenerateEnvFile(ctx context.Context, projectDir, database string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	envContent := map[string]string{
		"postgres": `# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=your_database
DB_SSLMODE=disable

# Application
APP_PORT=8080
APP_ENV=development
`,
		"mysql": `# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=your_database

# Application
APP_PORT=8080
APP_ENV=development
`,
		"mongodb": `# Database Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=your_database

# Application
APP_PORT=8080
APP_ENV=development
`,
		"sqlite": `# Database Configuration
DB_PATH=./data.db

# Application
APP_PORT=8080
APP_ENV=development
`,
	}

	content, exists := envContent[database]
	if !exists {
		content = fmt.Sprintf("# Database: %s\n# Add your database configuration here\n\nAPP_PORT=8080\nAPP_ENV=development\n", database)
	}

	filename := filepath.Join(projectDir, ".env.example")
	return os.WriteFile(filename, []byte(content), 0644)
}

// GenerateDockerfile generates a multi-stage Dockerfile for the Go project.
func GenerateDockerfile(ctx context.Context, projectDir string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dockerfile := `# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/server .
COPY --from=builder /app/.env.example .env

EXPOSE 8080

CMD ["./server"]
`

	filename := filepath.Join(projectDir, "Dockerfile")
	return os.WriteFile(filename, []byte(dockerfile), 0644)
}

// GenerateDockerCompose generates a docker-compose.yml pre-wired with the selected database.
func GenerateDockerCompose(ctx context.Context, projectDir, database string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dbService := map[string]string{
		"postgres": `  db:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: your_username
      POSTGRES_PASSWORD: your_password
      POSTGRES_DB: your_database
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
`,
		"mysql": `  db:
    image: mysql:8.0
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: your_root_password
      MYSQL_DATABASE: your_database
      MYSQL_USER: your_username
      MYSQL_PASSWORD: your_password
    ports:
      - "3306:3306"
    volumes:
      - mysqldata:/var/lib/mysql

volumes:
  mysqldata:
`,
		"mongodb": `  db:
    image: mongo:7
    restart: unless-stopped
    environment:
      MONGO_INITDB_DATABASE: your_database
    ports:
      - "27017:27017"
    volumes:
      - mongodata:/data/db

volumes:
  mongodata:
`,
	}

	dbBlock, exists := dbService[database]
	if !exists {
		dbBlock = fmt.Sprintf("  # Add your %s service configuration here\n", database)
	}

	compose := fmt.Sprintf(`version: "3.8"

services:
  app:
    build: .
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - .env
    depends_on:
      - db

%s`, dbBlock)

	filename := filepath.Join(projectDir, "docker-compose.yml")
	return os.WriteFile(filename, []byte(compose), 0644)
}
