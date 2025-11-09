package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all SQL migrations in order
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	log.Info().Msg("Starting database migrations...")

	// Get all migration files
	migrationFiles, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	// Filter only .sql files and sort
	var sqlFiles []string
	for _, file := range migrationFiles {
		if strings.HasSuffix(file, ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}
	migrationFiles = sqlFiles

	// Sort migration files by name (001, 002, etc.)
	sort.Strings(migrationFiles)

	// Check if system_info table exists (indicates if migrations have been run)
	var tableExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'system_info'
		)
	`).Scan(&tableExists)
	if err != nil {
		log.Warn().Err(err).Msg("Could not check if migrations have been run, proceeding anyway")
		tableExists = false
	}

	if tableExists {
		// Check if all expected tables exist
		expectedTables := []string{"system_info", "agents", "trades", "portfolios", "decisions", "articles", "stocks"}
		var missingTables []string

		for _, tableName := range expectedTables {
			var exists bool
			err := pool.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = $1
				)
			`, tableName).Scan(&exists)
			if err != nil || !exists {
				missingTables = append(missingTables, tableName)
			}
		}

		if len(missingTables) == 0 {
			log.Info().Msg("Database already initialized with all tables, skipping migrations")
			return nil
		}

		log.Info().Strs("missing_tables", missingTables).Msg("Some tables are missing, will apply migrations")
	}

	// Apply each migration
	for _, migrationFile := range migrationFiles {
		log.Info().Str("file", migrationFile).Msg("Applying migration")

		// Read migration file
		content, err := migrationsFS.ReadFile(migrationFile)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationFile, err)
		}

		// Execute migration
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			// Check if error is due to table already existing (CREATE TABLE IF NOT EXISTS)
			if strings.Contains(err.Error(), "already exists") {
				log.Warn().Str("file", migrationFile).Msg("Migration already applied, skipping")
				continue
			}
			return fmt.Errorf("failed to apply migration %s: %w", migrationFile, err)
		}

		log.Info().Str("file", filepath.Base(migrationFile)).Msg("Migration applied successfully")
	}

	log.Info().Int("count", len(migrationFiles)).Msg("All migrations applied successfully")
	return nil
}

