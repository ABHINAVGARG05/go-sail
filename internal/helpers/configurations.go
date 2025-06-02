package helpers

import (
	"fmt"

	"github.com/TejasGhatte/go-sail/internal/initializers"
	"github.com/TejasGhatte/go-sail/internal/models"
)

// Provider interface defines methods for generating database connection and migration code
type Provider interface {
	GetImports() []string
	GetConnectionCode() string
	GetMigrationCode() string
	GetDBVariable() string
}

type DBProvider struct {
	Database      models.DatabaseConfig
	ORM           models.ORMConfig
	Combination   models.CombinationConfig
	MigrationCode string
}

func (p *DBProvider) GetImports() []string {
	importMap := make(map[string]bool)
	var imports []string

	// Include ORM import path if available
	if p.ORM.ImportPath != "" {
		importPath := fmt.Sprintf("%q", p.ORM.ImportPath)
		if !importMap[importPath] {
			imports = append(imports, importPath)
			importMap[importPath] = true
		}
	}
	if p.Database.Name == "mongodb" {
		extraImports := []string{"time", "context"}
		for _, imp := range extraImports {
			formatted := fmt.Sprintf("%q", imp)
			if !importMap[formatted] {
				imports = append(imports, formatted)
				importMap[formatted] = true
			}
		}

		for _, additionalImport := range p.Combination.AdditionalImports {
			importPath := fmt.Sprintf("%q", additionalImport)
			if !importMap[importPath] {
				imports = append(imports, importPath)
				importMap[importPath] = true
			}
		}
	}
	
	if p.Database.DriverPkg != "" {
		importPath := fmt.Sprintf("%q", p.Database.DriverPkg)
		if !importMap[importPath] {
			imports = append(imports, importPath)
			importMap[importPath] = true
		}
	}

	return imports
}

func (p *DBProvider) GetConnectionCode() string {
	if p.Database.Name == "mongodb" {
		return `var err error
    dsn := "mongodb://localhost:27017/your_database_name"
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
    if err != nil {
        panic(fmt.Errorf("failed to connect to MongoDB: %w", err))
    }
    
    if err := client.Ping(ctx, nil); err != nil {
        panic(fmt.Errorf("failed to ping MongoDB: %w", err))
    }
    
    DB = client.Database("your_database_name")
    fmt.Println("Connected to MongoDB")`
	}

	return fmt.Sprintf(`var err error
    dsn := fmt.Sprintf(%q, "your_username", "your_password", "your_database")
    DB, err = %s
    if err != nil {
        fmt.Println("failed to connect to database")
    }
    fmt.Println("Connected to database")`, p.Combination.DSNTemplate, p.Combination.InitFunc)
}

func (p *DBProvider) GetMigrationCode() string {
	return p.MigrationCode 
}

func (p *DBProvider) GetDBVariable() string {
	if p.Database.Name == "mongodb" {
		return "*mongo.Database"
	}
	return fmt.Sprintf("*%s.DB", p.ORM.Name)
}

func ProviderFactory(database, orm string) (Provider, error) {
	dbConfig, dbExists := initializers.Config.Databases[database]
	if !dbExists {
		return nil, fmt.Errorf("database configuration for %q not found", database)
	}

	ormConfig, ormExists := initializers.Config.ORMs[orm]
	if !ormExists {
		return nil, fmt.Errorf("ORM configuration for %q not found", orm)
	}

	combinationConfig, combinationExists := initializers.Config.Combinations[database][orm]
	if !combinationExists {
		return nil, fmt.Errorf("combination configuration for database %q and ORM %q not found", database, orm)
	}

	return &DBProvider{
		Database:      dbConfig,
		ORM:           ormConfig,
		Combination:   combinationConfig,
		MigrationCode: initializers.Config.MigrationCode[orm],
	}, nil
}
