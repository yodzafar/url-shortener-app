// Command migrate runs goose SQL migrations against the configured database.
//
// Usage: go run cmd/migrate/main.go -cmd up|down|reset|status|version|create [-name <name>]
package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	cmd := flag.String("cmd", "", "up | down | reset | status | version | create")
	name := flag.String("name", "", "migration name (for -cmd create)")
	flag.Parse()

	// Best-effort .env load; ignore if missing.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("warning: loading .env: %v", err)
	}

	dir := getenv("DB_MIGRATIONS_DIR", "migrations")
	dbURL := os.Getenv("DB_URL")

	if *cmd == "" {
		log.Fatal("missing -cmd (up|down|reset|status|version|create)")
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}

	// `create` does not need a DB connection.
	if *cmd == "create" {
		if *name == "" {
			log.Fatal("missing -name for create")
		}
		if err := goose.Create(nil, dir, *name, "sql"); err != nil {
			log.Fatalf("create migration: %v", err)
		}
		return
	}

	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	switch *cmd {
	case "up":
		err = goose.Up(db, dir)
	case "down":
		err = goose.Down(db, dir)
	case "reset":
		err = goose.Reset(db, dir)
	case "status":
		err = goose.Status(db, dir)
	case "version":
		err = goose.Version(db, dir)
	default:
		log.Fatalf("unknown -cmd %q", *cmd)
	}

	if err != nil {
		log.Fatalf("migrate %s: %v", *cmd, err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
