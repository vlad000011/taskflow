package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// runMigrations applies ordered *.up.sql files once. Schema changes remain versioned in SQL.
func runMigrations(db *gorm.DB, dir string) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(255) PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`).Error; err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		version := strings.TrimSuffix(filepath.Base(file), ".up.sql")
		var count int64
		if err := db.Table("schema_migrations").Where("version = ?", version).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(content)).Error; err != nil {
				return fmt.Errorf("migration %s: %w", version, err)
			}
			return tx.Exec("INSERT INTO schema_migrations(version) VALUES (?)", version).Error
		}); err != nil {
			return err
		}
	}
	return nil
}
