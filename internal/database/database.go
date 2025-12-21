package database

import (
	"fmt"
	"strings"
	"uspavalia/internal/config"
	"uspavalia/internal/models"

	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Initialize(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Auto-detect database type based on environment
	dbType := cfg.Database.Type
	if dbType == "" {
		// Auto-detect: if MySQL env vars are set, use MySQL, otherwise SQLite
		if cfg.Database.Host != "" && cfg.Database.User != "" {
			dbType = "mysql"
		} else {
			dbType = "sqlite"
		}
	}

	gormConfig := &gorm.Config{}

	switch strings.ToLower(dbType) {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL database: %w", err)
		}
		logrus.Printf("Connected to MySQL database at %s:%d", cfg.Database.Host, cfg.Database.Port)

	case "sqlite":
		dbPath := cfg.Database.Path
		if dbPath == "" {
			dbPath = "./uspavalia.db"
		}
		db, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}
		logrus.Printf("Connected to SQLite database at %s", dbPath)

	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: mysql, sqlite)", dbType)
	}

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create database views after table migrations
	if err := CreateViews(db); err != nil {
		logrus.Printf("Warning: Failed to create database views: %v", err)
		// Don't fail the connection for view creation errors
	}

	logrus.Println("Database connection established and migrations completed")
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.LoginToken{},
		&models.Unit{},
		&models.Course{},
		&models.Discipline{},
		&models.Professor{},
		&models.ClassProfessor{},
		&models.Vote{},
		&models.Comment{},
		&models.CommentVote{},
		&models.ClassOffering{},
	)
}

func CreateViews(db *gorm.DB) error {
	// Detect database type
	var dbType string
	if sqlDB, err := db.DB(); err == nil {
		if driver := sqlDB.Driver(); driver != nil {
			switch driver.(type) {
			case *sqlite3.SQLiteDriver:
				dbType = "sqlite"
			default:
				dbType = "mysql"
			}
		}
	}

	// Create ListaMedias view
	var listMediasSQL string
	if dbType == "sqlite" {
		// SQLite version without CREATE OR REPLACE (not supported in older SQLite)
		// Drop view if exists, then create
		db.Exec("DROP VIEW IF EXISTS ListaMedias")
		listMediasSQL = `
			CREATE VIEW ListaMedias AS
			SELECT class_professor_id, AVG(score) AS "AVG(nota)", COUNT(*) AS "COUNT(*)"
			FROM votes
			WHERE type <> 5
			GROUP BY class_professor_id
		`
	} else {
		// MySQL version with backticks
		listMediasSQL = `
			CREATE OR REPLACE VIEW ListaMedias AS
			SELECT class_professor_id, AVG(score) AS 'AVG(nota)', COUNT(*) AS 'COUNT(*)'
			FROM votes
			WHERE type <> 5
			GROUP BY class_professor_id
		`
	}

	if err := db.Exec(listMediasSQL).Error; err != nil {
		return fmt.Errorf("failed to create ListaMedias view: %w", err)
	}

	// Create Melhores view
	var melhoresSQL string
	if dbType == "sqlite" {
		// SQLite version
		db.Exec("DROP VIEW IF EXISTS Melhores")
		melhoresSQL = `
			CREATE VIEW Melhores AS
			SELECT
				(l."AVG(nota)" * 2) AS media,
				l."COUNT(*)" AS votos,
				d.name AS materia,
				u.name AS unidade,
				d.code AS codigo,
				p.name AS professor,
				ap.id AS id
			FROM ListaMedias l
			JOIN class_professors ap ON l.class_professor_id = ap.id
			JOIN disciplines d ON ap.class_id = d.id
			JOIN units u ON d.unit_id = u.id
			JOIN professors p ON ap.professor_id = p.id
			WHERE l."COUNT(*)" >= 15
			ORDER BY l."AVG(nota)" DESC, l."COUNT(*)" DESC
		`
	} else {
		// MySQL version with backticks
		melhoresSQL = `
			CREATE OR REPLACE VIEW Melhores AS
			SELECT
				(ListaMedias.'AVG(nota)' * 2) AS media,
				ListaMedias.'COUNT(*)' AS votos,
				disciplines.name AS materia,
				units.name AS unidade,
				disciplines.code AS codigo,
				professors.name AS professor,
				class_professors.id AS id
			FROM ListaMedias
			JOIN class_professors ON ListaMedias.class_professor_id = class_professors.id
			JOIN disciplines ON class_professors.class_id = disciplines.id
			JOIN units ON disciplines.unit_id = units.id
			JOIN professors ON class_professors.professor_id = professors.id
			WHERE ListaMedias.'COUNT(*)' >= 15
			ORDER BY ListaMedias.'AVG(nota)' DESC, ListaMedias.'COUNT(*)' DESC
		`
	}

	if err := db.Exec(melhoresSQL).Error; err != nil {
		return fmt.Errorf("failed to create Melhores view: %w", err)
	}

	logrus.Printf("Database views created successfully for %s", dbType)
	return nil
}
