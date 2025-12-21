package cmd

import (
	"uspavalia/internal/config"
	"uspavalia/internal/database"
	"uspavalia/internal/handlers"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	Long:  `Start the USP Avalia web server with all configured routes and middleware.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		db, err := database.Initialize(cfg)
		if err != nil {
			logrus.Fatal("Failed to initialize database:", err)
		}

		server := handlers.NewServer(cfg, db)
		logrus.Fatal(server.Start())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
