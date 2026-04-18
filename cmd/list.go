package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TejasGhatte/go-sail/internal/initializers"
	"github.com/spf13/cobra"
)

var ListCommand = &cobra.Command{
	Use:   "list",
	Short: "List all supported frameworks, databases, and ORMs",
	Long:  "Displays all available frameworks, databases, ORM options, and their valid combinations.",
	Run: func(cmd *cobra.Command, args []string) {
		printSupportedOptions()
	},
}

func printSupportedOptions() {
	// Frameworks
	fmt.Println("Supported Frameworks:")
	fmt.Println(strings.Repeat("-", 40))
	frameworks := sortedKeys(initializers.Config.Repositories)
	for _, f := range frameworks {
		fmt.Printf("  • %s\n", f)
	}

	// Databases
	fmt.Println("\nSupported Databases:")
	fmt.Println(strings.Repeat("-", 40))
	databases := sortedKeys(initializers.Config.Databases)
	for _, db := range databases {
		cfg := initializers.Config.Databases[db]
		fmt.Printf("  • %-12s (driver: %s)\n", db, cfg.DriverPkg)
	}

	// ORMs
	fmt.Println("\nSupported ORMs:")
	fmt.Println(strings.Repeat("-", 40))
	orms := sortedKeys(initializers.Config.ORMs)
	for _, o := range orms {
		cfg := initializers.Config.ORMs[o]
		fmt.Printf("  • %-14s (import: %s)\n", o, cfg.ImportPath)
	}

	// Valid Combinations
	fmt.Println("\nValid Database + ORM Combinations:")
	fmt.Println(strings.Repeat("-", 40))
	for _, db := range databases {
		combos, exists := initializers.Config.Combinations[db]
		if !exists {
			continue
		}
		ormList := []string{}
		for ormName := range combos {
			ormList = append(ormList, ormName)
		}
		sort.Strings(ormList)
		fmt.Printf("  %s → %s\n", db, strings.Join(ormList, ", "))
	}
	fmt.Println()
}

// sortedKeys returns sorted keys from any map[string]T using generics-free approach
func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
