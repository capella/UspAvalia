package cmd

import (
	"fmt"
	"uspavalia/internal/config"
	"uspavalia/internal/database"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Runs GORM AutoMigrate to create/update database tables based on the current models.`,
	Run:   runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) {
	fmt.Println("Running database migrations...")

	cfg := config.Load()
	db, err := database.Initialize(cfg)
	if err != nil {
		fmt.Printf("Error: Failed to initialize database: %v\n", err)
		return
	}

	fmt.Println("✓ Database connection established")
	fmt.Println("✓ Migrations completed successfully")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run: ./uspavalia fetch-units --store")
	fmt.Println("  2. Run: ./uspavalia fetch-disciplines --store")
	fmt.Println("\nThis will populate your database with course and professor data.")

	// Get database instance to show table count
	sqlDB, err := db.DB()
	if err == nil {
		var tableCount int
		// Query varies by database type
		db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&tableCount)
		if tableCount == 0 {
			// Try SQLite query
			db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tableCount)
		}
		if tableCount > 0 {
			fmt.Printf("\nDatabase contains %d tables\n", tableCount)
		}
		sqlDB.Close()
	}
}
