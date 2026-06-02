// Usage:
//
//	go run cmd/grantadmin/main.go -email user@example.com [-role admin|user]
//	go run cmd/grantadmin/main.go -email admin@example.com -password secret123  # create if missing
package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	email := flag.String("email", "", "email of the user")
	role := flag.String("role", "admin", "role to assign (admin|user)")
	password := flag.String("password", "", "password to set when creating a missing user")
	flag.Parse()

	if *email == "" {
		log.Fatal("missing -email")
	}
	if *role != "admin" && *role != "user" {
		log.Fatalf("invalid -role %q (admin|user)", *role)
	}

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("warning: loading .env: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	addr := strings.ToLower(strings.TrimSpace(*email))

	// Try to promote an existing active user first.
	res, err := db.Exec(
		"UPDATE users SET role = $1 WHERE email = $2 AND is_deleted = false",
		*role, addr,
	)
	if err != nil {
		log.Fatalf("update role: %v", err)
	}

	if affected, _ := res.RowsAffected(); affected > 0 {
		log.Printf("→ %s is now %q", addr, *role)
		return
	}

	// No such user — create it. Use the given password or the default.
	pw := *password
	if pw == "" {
		pw = defaultPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO users (email, password_hash, role, is_active) VALUES ($1, $2, $3, true)",
		addr, string(hash), *role,
	)
	if err != nil {
		log.Fatalf("create user: %v", err)
	}

	log.Printf("→ created %s as %q (password: %s)", addr, *role, pw)
}

// defaultPassword is used when creating a user without an explicit -password.
const defaultPassword = "987456321"
