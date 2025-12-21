package cmd

import (
	"fmt"
	"os"
	"uspavalia/internal/config"
	"uspavalia/internal/database"
	"uspavalia/internal/models"

	"github.com/spf13/cobra"
)

var (
	fetchUnitsStore bool
)

var fetchUnitsCmd = &cobra.Command{
	Use:   "fetch-units",
	Short: "Fetch all USP teaching units from Jupiter Web",
	Long:  `Fetches all teaching units (departments) from USP Jupiter Web with their codes and names.`,
	Run:   runFetchUnits,
}

func init() {
	rootCmd.AddCommand(fetchUnitsCmd)
	fetchUnitsCmd.Flags().BoolVar(&fetchUnitsStore, "store", false, "Store units in database")
}

func runFetchUnits(cmd *cobra.Command, args []string) {
	fmt.Println("Fetching teaching units from Jupiter Web...")

	units, err := getTeachingUnits()
	if err != nil {
		fmt.Printf("Error: Failed to fetch units: %v\n", err)
		os.Exit(1)
	}

	if len(units) == 0 {
		fmt.Println("No units found.")
		return
	}

	fmt.Printf("\nFound %d teaching units:\n\n", len(units))
	fmt.Printf("%-8s | %s\n", "Code", "Name")
	fmt.Println("---------+--------------------------------------------------")
	for _, unit := range units {
		fmt.Printf("%-8s | %s\n", unit.Code, unit.Name)
	}

	if fetchUnitsStore {
		fmt.Println("\nStoring units in database...")

		cfg := config.Load()
		db, err := database.Initialize(cfg)
		if err != nil {
			fmt.Printf("Error: Failed to initialize database: %v\n", err)
			os.Exit(1)
		}

		stored := 0
		for _, unit := range units {
			// Check if unit already exists by ID
			// We need to parse the code as uint
			var unitID uint
			fmt.Sscanf(unit.Code, "%d", &unitID)

			dbUnit := models.Unit{
				ID:   unitID,
				Name: unit.Name,
			}

			// Use FirstOrCreate to avoid duplicates
			result := db.Where(models.Unit{ID: unitID}).FirstOrCreate(&dbUnit)
			if result.Error != nil {
				fmt.Printf("Warning: Failed to store unit %s: %v\n", unit.Code, result.Error)
				continue
			}

			// Update name if it changed
			if result.RowsAffected == 0 {
				db.Model(&dbUnit).Update("NOME", unit.Name)
			}
			stored++
		}

		fmt.Printf("Successfully stored %d units in database.\n", stored)
	}
}
