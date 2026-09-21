package database_test

import (
	"os"
	"path/filepath"
	"testing"

	"cards-api/internal/database"
)

func TestMigrationsApplyOnceAndAreIdempotent(t *testing.T) {
	db, err := database.Connect(filepath.Join(t.TempDir(), "cards.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for i := 0; i < 2; i++ {
		if err := database.RunMigrations(db, "../../migrations"); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}

	var applied int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if applied < 4 {
		t.Fatalf("expected at least 4 applied migrations, got %d", applied)
	}

	// Columns and indexes introduced by later migrations must exist.
	for _, stmt := range []string{
		"SELECT title_ar, title_en, template_ar, template_en FROM campaigns LIMIT 1",
		"SELECT lang, field_values FROM campaign_cards LIMIT 1",
		"SELECT name FROM sqlite_master WHERE type='index' AND name='idx_campaign_cards_slug_id'",
	} {
		rows, err := db.Query(stmt)
		if err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
		_ = rows.Close()
	}
}

// A failing migration must surface as an error (the server refuses to boot) and leave no partial state behind.
func TestFailedMigrationIsReportedAndRolledBack(t *testing.T) {
	dir := t.TempDir()
	write := func(name, sql string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(sql), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("000001_ok.up.sql", "CREATE TABLE ok (id INTEGER);")
	write("000002_broken.up.sql", "CREATE TABLE half (id INTEGER); THIS IS NOT SQL;")

	db, err := database.Connect(filepath.Join(t.TempDir(), "cards.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, dir); err == nil {
		t.Fatal("expected an error from the broken migration")
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE name = 'half'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("broken migration must roll back completely (count=%d, err=%v)", count, err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil || count != 1 {
		t.Fatalf("only the first migration should be recorded (count=%d, err=%v)", count, err)
	}
}
