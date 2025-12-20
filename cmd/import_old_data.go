package cmd

import (
	"fmt"
	"os"
	"uspavalia/internal/config"
	"uspavalia/internal/database"
	"uspavalia/internal/models"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Old database schema structs (for reading legacy data)
type oldProfessor struct {
	ID     uint   `gorm:"column:id;primaryKey"`
	Nome   string `gorm:"column:nome"`
	UnitID uint   `gorm:"column:idunidade"`
	Usage  string `gorm:"column:uso"`
	Time   *int64 `gorm:"column:time"`
}

func (oldProfessor) TableName() string {
	return "professores"
}

type oldVote struct {
	ID               uint   `gorm:"column:id;primaryKey"`
	ClassProfessorID uint   `gorm:"column:APid"`
	UserID           string `gorm:"column:iduso"`
	Time             int64  `gorm:"column:time"`
	Nota             int    `gorm:"column:nota"`
	Tipo             int    `gorm:"column:tipo"`
}

func (oldVote) TableName() string {
	return "votos"
}

type oldClassProfessor struct {
	ID          uint   `gorm:"column:id;primaryKey"`
	ClassID     uint   `gorm:"column:idaula"`
	ProfessorID uint   `gorm:"column:idprofessor"`
	Usage       string `gorm:"column:uso"`
	Time        *int64 `gorm:"column:time"`
}

func (oldClassProfessor) TableName() string {
	return "aulaprofessor"
}

type oldDiscipline struct {
	ID     uint   `gorm:"column:id;primaryKey"`
	Nome   string `gorm:"column:nome"`
	Codigo string `gorm:"column:codigo"`
	UnitID uint   `gorm:"column:idunidade"`
	Usage  string `gorm:"column:uso"`
	Time   *int64 `gorm:"column:time"`
}

func (oldDiscipline) TableName() string {
	return "disciplinas"
}

type oldUnit struct {
	ID   uint   `gorm:"column:id;primaryKey"`
	Nome string `gorm:"column:NOME"`
}

func (oldUnit) TableName() string {
	return "unidades"
}

type oldComment struct {
	ID               uint   `gorm:"column:id;primaryKey"`
	UserID           string `gorm:"column:iduso"`
	Comantario       string `gorm:"column:comantario"`
	ClassProfessorID uint   `gorm:"column:aulaprofessorid"`
	Time             int64  `gorm:"column:time"`
}

func (oldComment) TableName() string {
	return "cometario"
}

type oldCommentVote struct {
	ID        uint   `gorm:"column:id;primaryKey"`
	CommentID uint   `gorm:"column:idcomentario"`
	Time      int64  `gorm:"column:time"`
	Vote      int    `gorm:"column:voto"`
	UserID    string `gorm:"column:iduso"`
}

func (oldCommentVote) TableName() string {
	return "votoscomentario"
}

var importOldDataCMD = &cobra.Command{
	Use:   "import-old-data",
	Short: "Import data from old MySQL database into the new database",
	Long: `Connects to the legacy PHP/MySQL database and imports all data into the new Go application database.

This command will:
1. Connect to the old MySQL database (read-only)
2. Read data from legacy tables (unidades, disciplinas, professores, etc.)
3. Map old data to new Go models
4. Insert into the new database

Configure the old database connection in .env:
  USPAVALIA_OLD_DATABASE_HOST=localhost
  USPAVALIA_OLD_DATABASE_PORT=3306
  USPAVALIA_OLD_DATABASE_USER=root
  USPAVALIA_OLD_DATABASE_PASSWORD=password
  USPAVALIA_OLD_DATABASE_NAME=uspavalia_old

Note: This is a one-time migration tool. Run it only on a fresh database.`,
	Run: runImportOldData,
}

func init() {
	rootCmd.AddCommand(importOldDataCMD)
}

func runImportOldData(cmd *cobra.Command, args []string) {
	fmt.Println("Starting import from old MySQL database")

	// Load configuration
	cfg := config.Load()

	// Validate old database configuration
	if cfg.OldDatabase.Host == "" || cfg.OldDatabase.Name == "" {
		fmt.Println("Error: Old database configuration is missing")
		fmt.Println("Please set USPAVALIA_OLD_DATABASE_HOST, USPAVALIA_OLD_DATABASE_NAME, etc. in .env")
		os.Exit(1)
	}

	// Connect to old database (read-only)
	oldDB, err := connectToOldDatabase(cfg)
	if err != nil {
		fmt.Printf("Error connecting to old database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to old database")

	// Connect to new database
	newDB, err := database.Initialize(cfg)
	if err != nil {
		fmt.Printf("Error connecting to new database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to new database")

	// Import data in order (respecting foreign key constraints)
	if err := importData(oldDB, newDB); err != nil {
		fmt.Printf("Error importing data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nImport completed successfully!")
}

func connectToOldDatabase(cfg *config.Config) (*gorm.DB, error) {
	// Build DSN for old MySQL database with read-only mode
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.OldDatabase.User,
		cfg.OldDatabase.Password,
		cfg.OldDatabase.Host,
		cfg.OldDatabase.Port,
		cfg.OldDatabase.Name,
	)

	// Open connection with read-only settings
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		// Disable write operations
		PrepareStmt: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Set connection to read-only mode
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Execute SET SESSION TRANSACTION READ ONLY
	if err := db.Exec("SET SESSION TRANSACTION READ ONLY").Error; err != nil {
		logrus.Warnf("Could not set read-only mode: %v (continuing anyway)", err)
	}

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func importData(oldDB, newDB *gorm.DB) error {
	// Import in order of foreign key dependencies
	importSteps := []struct {
		name     string
		function func(*gorm.DB, *gorm.DB) error
	}{
		{"Units", importUnits},
		{"Disciplines", importDisciplines},
		{"Professors", importProfessors},
		{"ClassProfessors", importClassProfessors},
		{"Votes", importVotes},
		{"Comments", importComments},
		{"CommentVotes", importCommentVotes},
	}

	for _, step := range importSteps {
		fmt.Printf("Importing %s...\n", step.name)
		if err := step.function(oldDB, newDB); err != nil {
			return fmt.Errorf("failed to import %s: %w", step.name, err)
		}
	}

	return nil
}

func importUnits(oldDB, newDB *gorm.DB) error {
	var oldUnits []oldUnit

	// Read from old database using old schema
	if err := oldDB.Find(&oldUnits).Error; err != nil {
		return fmt.Errorf("failed to read old units: %w", err)
	}

	if len(oldUnits) == 0 {
		fmt.Println("  No units to import")
		return nil
	}

	// Convert to new schema
	newUnits := make([]models.Unit, len(oldUnits))
	for i, oldU := range oldUnits {
		newUnits[i] = models.Unit{
			ID:   oldU.ID,
			Name: oldU.Nome,
		}
	}

	// Insert into new database in batches
	batchSize := 100
	for i := 0; i < len(newUnits); i += batchSize {
		end := i + batchSize
		if end > len(newUnits) {
			end = len(newUnits)
		}

		batch := newUnits[i:end]
		if err := newDB.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert units batch: %w", err)
		}
	}

	fmt.Printf("  Imported %d units\n", len(newUnits))
	return nil
}

func importDisciplines(oldDB, newDB *gorm.DB) error {
	var oldDisciplines []oldDiscipline

	// Read from old database using old schema
	if err := oldDB.Find(&oldDisciplines).Error; err != nil {
		return fmt.Errorf("failed to read old disciplines: %w", err)
	}

	if len(oldDisciplines) == 0 {
		fmt.Println("  No disciplines to import")
		return nil
	}

	// Convert to new schema
	newDisciplines := make([]models.Discipline, len(oldDisciplines))
	for i, oldD := range oldDisciplines {
		newDisciplines[i] = models.Discipline{
			ID:     oldD.ID,
			Name:   oldD.Nome,
			Code:   oldD.Codigo,
			UnitID: oldD.UnitID,
			Usage:  oldD.Usage,
			Time:   oldD.Time,
		}
	}

	// Insert into new database in batches
	batchSize := 100
	for i := 0; i < len(newDisciplines); i += batchSize {
		end := i + batchSize
		if end > len(newDisciplines) {
			end = len(newDisciplines)
		}

		batch := newDisciplines[i:end]
		if err := newDB.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert disciplines batch: %w", err)
		}
	}

	fmt.Printf("  Imported %d disciplines\n", len(newDisciplines))
	return nil
}

func importProfessors(oldDB, newDB *gorm.DB) error {
	var oldProfessors []oldProfessor

	// Read from old database using old schema
	if err := oldDB.Find(&oldProfessors).Error; err != nil {
		return fmt.Errorf("failed to read old professors: %w", err)
	}

	if len(oldProfessors) == 0 {
		fmt.Println("  No professors to import")
		return nil
	}

	// Convert to new schema and filter out professors with empty names
	validProfessors := make([]models.Professor, 0, len(oldProfessors))
	skipped := 0
	for _, oldProf := range oldProfessors {
		if oldProf.Nome != "" {
			newProf := models.Professor{
				ID:     oldProf.ID,
				Name:   oldProf.Nome,
				UnitID: oldProf.UnitID,
				Usage:  oldProf.Usage,
				Time:   oldProf.Time,
			}
			validProfessors = append(validProfessors, newProf)
		} else {
			skipped++
		}
	}

	if skipped > 0 {
		fmt.Printf("  Skipped %d professors with empty names\n", skipped)
	}

	if len(validProfessors) == 0 {
		fmt.Println("  No valid professors to import")
		return nil
	}

	// Insert into new database in batches
	batchSize := 100
	for i := 0; i < len(validProfessors); i += batchSize {
		end := i + batchSize
		if end > len(validProfessors) {
			end = len(validProfessors)
		}

		batch := validProfessors[i:end]
		if err := newDB.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert professors batch: %w", err)
		}
	}

	fmt.Printf("  Imported %d professors\n", len(validProfessors))
	return nil
}

func importClassProfessors(oldDB, newDB *gorm.DB) error {
	var oldClassProfessors []oldClassProfessor

	// Read from old database using old schema
	if err := oldDB.Find(&oldClassProfessors).Error; err != nil {
		return fmt.Errorf("failed to read old class professors: %w", err)
	}

	if len(oldClassProfessors) == 0 {
		fmt.Println("  No class professors to import")
		return nil
	}

	// Convert to new schema
	newClassProfessors := make([]models.ClassProfessor, len(oldClassProfessors))
	for i, oldCP := range oldClassProfessors {
		newClassProfessors[i] = models.ClassProfessor{
			ID:          oldCP.ID,
			ClassID:     oldCP.ClassID,
			ProfessorID: oldCP.ProfessorID,
			Usage:       oldCP.Usage,
			Time:        oldCP.Time,
		}
	}

	// Insert into new database in batches
	batchSize := 100
	for i := 0; i < len(newClassProfessors); i += batchSize {
		end := i + batchSize
		if end > len(newClassProfessors) {
			end = len(newClassProfessors)
		}

		batch := newClassProfessors[i:end]
		if err := newDB.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert class professors batch: %w", err)
		}
	}

	fmt.Printf("  Imported %d class professors\n", len(newClassProfessors))
	return nil
}

func importVotes(oldDB, newDB *gorm.DB) error {
	var oldVotes []oldVote

	// Read from old database using old schema
	if err := oldDB.Find(&oldVotes).Error; err != nil {
		return fmt.Errorf("failed to read old votes: %w", err)
	}

	if len(oldVotes) == 0 {
		fmt.Println("  No votes to import")
		return nil
	}

	// Convert to new schema
	newVotes := make([]models.Vote, len(oldVotes))
	for i, oldV := range oldVotes {
		newVotes[i] = models.Vote{
			ID:               oldV.ID,
			ClassProfessorID: oldV.ClassProfessorID,
			UserID:           oldV.UserID,
			Time:             oldV.Time,
			Score:            oldV.Nota,
			Type:             oldV.Tipo,
		}
	}

	// Insert into new database in batches
	batchSize := 500
	imported := 0
	for i := 0; i < len(newVotes); i += batchSize {
		end := i + batchSize
		if end > len(newVotes) {
			end = len(newVotes)
		}

		batch := newVotes[i:end]
		if err := newDB.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert votes batch: %w", err)
		}

		imported += len(batch)
		if imported%5000 == 0 {
			fmt.Printf("  Imported %d/%d votes...\n", imported, len(newVotes))
		}
	}

	fmt.Printf("  Imported %d votes\n", len(newVotes))
	return nil
}

func importComments(oldDB, newDB *gorm.DB) error {
	var oldComments []oldComment

	// Read from old database using old schema
	if err := oldDB.Find(&oldComments).Error; err != nil {
		return fmt.Errorf("failed to read old comments: %w", err)
	}

	if len(oldComments) == 0 {
		fmt.Println("  No comments to import")
		return nil
	}

	// Convert to new schema
	newComments := make([]models.Comment, len(oldComments))
	for i, oldC := range oldComments {
		newComments[i] = models.Comment{
			ID:               oldC.ID,
			UserID:           oldC.UserID,
			Content:          oldC.Comantario,
			ClassProfessorID: oldC.ClassProfessorID,
			Time:             oldC.Time,
		}
	}

	// Insert into new database in batches
	batchSize := 100
	for i := 0; i < len(newComments); i += batchSize {
		end := i + batchSize
		if end > len(newComments) {
			end = len(newComments)
		}

		batch := newComments[i:end]
		if err := newDB.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert comments batch: %w", err)
		}
	}

	fmt.Printf("  Imported %d comments\n", len(newComments))
	return nil
}

func importCommentVotes(oldDB, newDB *gorm.DB) error {
	var oldCommentVotes []oldCommentVote

	// Read from old database using old schema
	if err := oldDB.Find(&oldCommentVotes).Error; err != nil {
		return fmt.Errorf("failed to read old comment votes: %w", err)
	}

	if len(oldCommentVotes) == 0 {
		fmt.Println("  No comment votes to import")
		return nil
	}

	// Convert to new schema
	newCommentVotes := make([]models.CommentVote, len(oldCommentVotes))
	for i, oldCV := range oldCommentVotes {
		newCommentVotes[i] = models.CommentVote{
			ID:        oldCV.ID,
			CommentID: oldCV.CommentID,
			Time:      oldCV.Time,
			Vote:      oldCV.Vote,
			UserID:    oldCV.UserID,
		}
	}

	// Insert into new database in batches
	batchSize := 100
	for i := 0; i < len(newCommentVotes); i += batchSize {
		end := i + batchSize
		if end > len(newCommentVotes) {
			end = len(newCommentVotes)
		}

		batch := newCommentVotes[i:end]
		if err := newDB.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert comment votes batch: %w", err)
		}
	}

	fmt.Printf("  Imported %d comment votes\n", len(newCommentVotes))
	return nil
}
