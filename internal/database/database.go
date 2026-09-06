package database

import (
	"fmt"
	"strings"
	"uspavalia/internal/config"
	"uspavalia/internal/models"

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
		&models.ScrapeRun{},
	)
}

// Dialect returns "sqlite" or "mysql" for the connected database.
func Dialect(db *gorm.DB) string {
	if db.Dialector.Name() == "sqlite" {
		return "sqlite"
	}
	return "mysql"
}

// secondsPerYear is the length of a year (365.25 days) used to age votes.
const secondsPerYear = 31557600

// VoteWeightSQL returns an SQL expression that weights a vote by its age:
// votes cast in the last year weigh 1, and the weight halves for every
// full year since the vote was cast (1/2, 1/4, 1/8, ...). timeCol is the
// column holding the vote's unix timestamp, e.g. "votes.time" or "v.time".
//
// It uses a bit shift instead of POWER() because the bundled SQLite build
// has no math functions. The shift is clamped to [0, 62] so votes with a
// timestamp in the future weigh 1 and very old ones cannot overflow.
func VoteWeightSQL(dialect, timeCol string) string {
	if dialect == "sqlite" {
		return fmt.Sprintf(
			"(1.0 / (1 << MAX(0, MIN(62, (CAST(strftime('%%s', 'now') AS INTEGER) - %s) / %d))))",
			timeCol, secondsPerYear,
		)
	}
	return fmt.Sprintf(
		"(1.0 / (1 << GREATEST(0, LEAST(62, FLOOR((UNIX_TIMESTAMP() - %s) / %d)))))",
		timeCol, secondsPerYear,
	)
}

// CreateViews (re)creates the ListaMedias and Melhores views.
//
// ListaMedias keeps the legacy plain average and count per class-professor
// and adds a recency-weighted average (weighted_avg) and the total weight
// (weight_sum), see VoteWeightSQL. Melhores ranks class-professors with at
// least MinVotesForTopRated votes by the weighted average, so recent votes
// take priority.
func CreateViews(db *gorm.DB) error {
	dbType := Dialect(db)
	weight := VoteWeightSQL(dbType, "votes.time")

	// The legacy column names "AVG(nota)" and "COUNT(*)" must be quoted;
	// SQLite uses double quotes and MySQL uses backticks.
	quote := func(name string) string {
		if dbType == "sqlite" {
			return `"` + name + `"`
		}
		return "`" + name + "`"
	}

	// Older SQLite versions lack CREATE OR REPLACE VIEW.
	createView := "CREATE OR REPLACE VIEW"
	if dbType == "sqlite" {
		createView = "CREATE VIEW"
		db.Exec("DROP VIEW IF EXISTS Melhores")
		db.Exec("DROP VIEW IF EXISTS ListaMedias")
	}

	listMediasSQL := fmt.Sprintf(`
		%s ListaMedias AS
		SELECT
			class_professor_id,
			AVG(score) AS %s,
			COUNT(*) AS %s,
			SUM(score * %s) / SUM(%s) AS weighted_avg,
			SUM(%s) AS weight_sum
		FROM votes
		WHERE type <> 5
		GROUP BY class_professor_id
	`, createView, quote("AVG(nota)"), quote("COUNT(*)"), weight, weight, weight)

	if err := db.Exec(listMediasSQL).Error; err != nil {
		return fmt.Errorf("failed to create ListaMedias view: %w", err)
	}

	melhoresSQL := fmt.Sprintf(`
		%s Melhores AS
		SELECT
			(l.weighted_avg * 2) AS media,
			l.%s AS votos,
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
		WHERE l.%s >= %d
		ORDER BY l.weighted_avg DESC, l.weight_sum DESC
	`, createView, quote("COUNT(*)"), quote("COUNT(*)"), models.MinVotesForTopRated)

	if err := db.Exec(melhoresSQL).Error; err != nil {
		return fmt.Errorf("failed to create Melhores view: %w", err)
	}

	logrus.Printf("Database views created successfully for %s", dbType)
	return nil
}
